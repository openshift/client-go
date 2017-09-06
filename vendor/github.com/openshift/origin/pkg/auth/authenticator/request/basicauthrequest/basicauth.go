package basicauthrequest

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/golang/glog"
	"github.com/openshift/origin/pkg/auth/authenticator"
)

type basicAuthRequestHandler struct {
	provider              string
	passwordAuthenticator authenticator.Password
	removeHeader          bool
}

func NewBasicAuthAuthentication(provider string, passwordAuthenticator authenticator.Password, removeHeader bool) authenticator.Request {
	return &basicAuthRequestHandler{provider, passwordAuthenticator, removeHeader}
}

func (authHandler *basicAuthRequestHandler) AuthenticateRequest(req *http.Request) (user.Info, bool, error) {
	username, password, hasBasicAuth, err := getBasicAuthInfo(req)
	if err != nil {
		return nil, false, err
	}
	if !hasBasicAuth {
		return nil, false, nil
	}

	user, ok, err := authHandler.passwordAuthenticator.AuthenticatePassword(username, password)
	if ok && authHandler.removeHeader {
		req.Header.Del("Authorization")
	}

	switch {
	case err != nil:
		glog.Errorf(`Error authenticating login %q with provider %q: %v`, username, authHandler.provider, err)
	case !ok:
		glog.V(4).Infof(`Login with provider %q failed for login %q`, authHandler.provider, username)
	case ok:
		glog.V(4).Infof(`Login with provider %q succeeded for login %q: %#v`, authHandler.provider, username, user)
	}
	return user, ok, err
}

// getBasicAuthInfo returns the username and password in the request's basic-auth Authorization header,
// a boolean indicating whether the request had a valid basic-auth header, and any error encountered
// attempting to extract the basic-auth data.
func getBasicAuthInfo(r *http.Request) (string, string, bool, error) {
	// Retrieve the Authorization header and check whether it contains basic auth information
	const basicScheme string = "Basic "
	auth := r.Header.Get("Authorization")

	if !strings.HasPrefix(auth, basicScheme) {
		return "", "", false, nil
	}

	str, err := base64.StdEncoding.DecodeString(auth[len(basicScheme):])
	if err != nil {
		return "", "", false, errors.New("No valid base64 data in basic auth scheme found")
	}

	cred := strings.SplitN(string(str), ":", 2)
	if len(cred) < 2 {
		return "", "", false, errors.New("Invalid Authorization header")
	}

	return cred[0], cred[1], true, nil
}
