package reaper

import (
	"time"

	"github.com/golang/glog"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/security/legacyclient"
)

func NewUserReaper(
	userClient client.UsersInterface,
	groupClient client.GroupsInterface,
	clusterBindingClient client.ClusterRoleBindingsInterface,
	bindingClient client.RoleBindingsNamespacer,
	authorizationsClient client.OAuthClientAuthorizationsInterface,
	sccClient legacyclient.SecurityContextConstraintInterface,
) kubectl.Reaper {
	return &UserReaper{
		userClient:           userClient,
		groupClient:          groupClient,
		clusterBindingClient: clusterBindingClient,
		bindingClient:        bindingClient,
		authorizationsClient: authorizationsClient,
		sccClient:            sccClient,
	}
}

type UserReaper struct {
	userClient           client.UsersInterface
	groupClient          client.GroupsInterface
	clusterBindingClient client.ClusterRoleBindingsInterface
	bindingClient        client.RoleBindingsNamespacer
	authorizationsClient client.OAuthClientAuthorizationsInterface
	sccClient            legacyclient.SecurityContextConstraintInterface
}

// Stop on a reaper is actually used for deletion.  In this case, we'll delete referencing identities, clusterBindings, and bindings,
// then delete the user
func (r *UserReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *metav1.DeleteOptions) error {
	removedSubject := kapi.ObjectReference{Kind: "User", Name: name}

	if err := reapClusterBindings(removedSubject, r.clusterBindingClient); err != nil {
		return err
	}

	if err := reapNamespacedBindings(removedSubject, r.bindingClient); err != nil {
		return err
	}

	// Remove the user from sccs
	sccs, err := r.sccClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, scc := range sccs.Items {
		retainedUsers := []string{}
		for _, user := range scc.Users {
			if user != name {
				retainedUsers = append(retainedUsers, user)
			}
		}
		if len(retainedUsers) != len(scc.Users) {
			updatedSCC := scc
			updatedSCC.Users = retainedUsers
			if _, err := r.sccClient.Update(&updatedSCC); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update scc/%s: %v", scc.Name, err)
			}
		}
	}

	// Remove the user from groups
	groups, err := r.groupClient.Groups().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, group := range groups.Items {
		retainedUsers := []string{}
		for _, user := range group.Users {
			if user != name {
				retainedUsers = append(retainedUsers, user)
			}
		}
		if len(retainedUsers) != len(group.Users) {
			updatedGroup := group
			updatedGroup.Users = retainedUsers
			if _, err := r.groupClient.Groups().Update(&updatedGroup); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update groups/%s: %v", group.Name, err)
			}
		}
	}

	// Remove the user's OAuthClientAuthorizations
	// Once https://github.com/kubernetes/kubernetes/pull/28112 is fixed, use a field selector
	// to filter on the userName, rather than fetching all authorizations and filtering client-side
	authorizations, err := r.authorizationsClient.OAuthClientAuthorizations().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, authorization := range authorizations.Items {
		if authorization.UserName == name {
			if err := r.authorizationsClient.OAuthClientAuthorizations().Delete(authorization.Name); err != nil && !kerrors.IsNotFound(err) {
				return err
			}
		}
	}

	// Intentionally leave identities that reference the user
	// The user does not "own" the identities
	// If the admin wants to remove the identities, that is a distinct operation

	// Remove the user
	if err := r.userClient.Users().Delete(name); err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	return nil
}
