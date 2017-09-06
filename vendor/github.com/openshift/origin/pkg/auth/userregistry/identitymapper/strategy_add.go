package identitymapper

import (
	kerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"

	"github.com/openshift/origin/pkg/user"
	userapi "github.com/openshift/origin/pkg/user/apis/user"
	userclient "github.com/openshift/origin/pkg/user/generated/internalclientset/typed/user/internalversion"
)

var _ = UserForNewIdentityGetter(&StrategyAdd{})

// StrategyAdd associates a new identity with a user with the identity's preferred username,
// adding to any existing identities associated with the user
type StrategyAdd struct {
	user        userclient.UserResourceInterface
	initializer user.Initializer
}

func NewStrategyAdd(user userclient.UserResourceInterface, initializer user.Initializer) UserForNewIdentityGetter {
	return &StrategyAdd{user, initializer}
}

func (s *StrategyAdd) UserForNewIdentity(ctx apirequest.Context, preferredUserName string, identity *userapi.Identity) (*userapi.User, error) {

	persistedUser, err := s.user.Get(preferredUserName, metav1.GetOptions{})

	switch {
	case kerrs.IsNotFound(err):
		// CreateUser a new user
		desiredUser := &userapi.User{}
		desiredUser.Name = preferredUserName
		desiredUser.Identities = []string{identity.Name}
		s.initializer.InitializeUser(identity, desiredUser)
		return s.user.Create(desiredUser)

	case err == nil:
		// If the existing user already references our identity, we're done
		if sets.NewString(persistedUser.Identities...).Has(identity.Name) {
			return persistedUser, nil
		}

		// Otherwise add our identity and update
		persistedUser.Identities = append(persistedUser.Identities, identity.Name)
		// If our newly added identity is the only one, initialize the user
		if len(persistedUser.Identities) == 1 {
			s.initializer.InitializeUser(identity, persistedUser)
		}
		return s.user.Update(persistedUser)

	default:
		// Fail on errors other than "not found"
		return nil, err
	}
}
