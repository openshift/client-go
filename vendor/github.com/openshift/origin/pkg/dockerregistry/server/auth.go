package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	context "github.com/docker/distribution/context"
	registryauth "github.com/docker/distribution/registry/auth"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationapi "k8s.io/kubernetes/pkg/apis/authorization/v1"

	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	"github.com/openshift/origin/pkg/util/httprequest"

	"github.com/openshift/origin/pkg/dockerregistry/server/audit"
	"github.com/openshift/origin/pkg/dockerregistry/server/client"
)

type deferredErrors map[string]error

func (d deferredErrors) Add(namespace string, name string, err error) {
	d[namespace+"/"+name] = err
}
func (d deferredErrors) Get(namespace string, name string) (error, bool) {
	err, exists := d[namespace+"/"+name]
	return err, exists
}
func (d deferredErrors) Empty() bool {
	return len(d) == 0
}

const (
	OpenShiftAuth = "openshift"

	defaultTokenPath = "/openshift/token"
	defaultUserName  = "anonymous"

	RealmKey      = "realm"
	TokenRealmKey = "tokenrealm"

	// AccessControllerOptionParams is an option name for passing
	// AccessControllerParams to AccessController.
	AccessControllerOptionParams = "_params"
)

func init() {
	registryauth.Register(OpenShiftAuth, registryauth.InitFunc(newAccessController))
}

// WithUserInfoLogger creates a new context with provided user infomation.
func WithUserInfoLogger(ctx context.Context, username, userid string) context.Context {
	ctx = context.WithValue(ctx, audit.AuditUserEntry, username)
	if len(userid) > 0 {
		ctx = context.WithValue(ctx, audit.AuditUserIDEntry, userid)
	}
	return context.WithLogger(ctx, context.GetLogger(ctx,
		audit.AuditUserEntry,
		audit.AuditUserIDEntry,
	))
}

type AccessController struct {
	realm          string
	tokenRealm     *url.URL
	registryClient client.RegistryClient
	auditLog       bool
}

var _ registryauth.AccessController = &AccessController{}

type authChallenge struct {
	realm string
	err   error
}

var _ registryauth.Challenge = &authChallenge{}

type tokenAuthChallenge struct {
	realm   string
	service string
	err     error
}

var _ registryauth.Challenge = &tokenAuthChallenge{}

// Errors used and exported by this package.
var (
	// Challenging errors
	ErrTokenRequired         = errors.New("authorization header required")
	ErrTokenInvalid          = errors.New("failed to decode credentials")
	ErrOpenShiftAccessDenied = errors.New("access denied")

	// Non-challenging errors
	ErrNamespaceRequired   = errors.New("repository namespace required")
	ErrUnsupportedAction   = errors.New("unsupported action")
	ErrUnsupportedResource = errors.New("unsupported resource")
)

// TokenRealm returns the template URL to use as the token realm redirect.
// An empty scheme/host in the returned URL means to match the scheme/host on incoming requests.
func TokenRealm(options map[string]interface{}) (*url.URL, error) {
	if options[TokenRealmKey] == nil {
		// If not specified, default to "/openshift/token", auto-detecting the scheme and host
		return &url.URL{Path: defaultTokenPath}, nil
	}

	tokenRealmString, ok := options[TokenRealmKey].(string)
	if !ok {
		return nil, fmt.Errorf("%s config option must be a string, got %T", TokenRealmKey, options[TokenRealmKey])
	}

	tokenRealm, err := url.Parse(tokenRealmString)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL in %s config option: %v", TokenRealmKey, err)
	}
	if len(tokenRealm.RawQuery) > 0 || len(tokenRealm.Fragment) > 0 {
		return nil, fmt.Errorf("%s config option may not contain query parameters or a fragment", TokenRealmKey)
	}
	if len(tokenRealm.Path) > 0 {
		return nil, fmt.Errorf("%s config option may not contain a path (%q was specified)", TokenRealmKey, tokenRealm.Path)
	}

	// pin to "/openshift/token"
	tokenRealm.Path = defaultTokenPath

	return tokenRealm, nil
}

// AccessControllerParams is the parameters for newAccessController
type AccessControllerParams struct {
	Logger         context.Logger
	RegistryClient client.RegistryClient
}

func newAccessController(options map[string]interface{}) (registryauth.AccessController, error) {
	params, ok := options[AccessControllerOptionParams].(AccessControllerParams)
	if !ok {
		return nil, fmt.Errorf("no parameters provided to Origin Auth handler")
	}

	params.Logger.Info("Using Origin Auth handler")

	realm, err := getStringOption("", RealmKey, "origin", options)
	if err != nil {
		return nil, err
	}

	tokenRealm, err := TokenRealm(options)
	if err != nil {
		return nil, err
	}

	ac := &AccessController{
		realm:          realm,
		tokenRealm:     tokenRealm,
		registryClient: params.RegistryClient,
	}

	if audit, ok := options["audit"]; ok {
		auditOptions := make(map[string]interface{})

		for k, v := range audit.(map[interface{}]interface{}) {
			if s, ok := k.(string); ok {
				auditOptions[s] = v
			}
		}

		ac.auditLog, err = getBoolOption("", "enabled", false, auditOptions)
		if err != nil {
			return nil, err
		}
	}

	return ac, nil
}

// Error returns the internal error string for this authChallenge.
func (ac *authChallenge) Error() string {
	return ac.err.Error()
}

// SetHeaders sets the basic challenge header on the response.
func (ac *authChallenge) SetHeaders(w http.ResponseWriter) {
	// WWW-Authenticate response challenge header.
	// See https://tools.ietf.org/html/rfc6750#section-3
	str := fmt.Sprintf("Basic realm=%s", ac.realm)
	if ac.err != nil {
		str = fmt.Sprintf("%s,error=%q", str, ac.Error())
	}
	w.Header().Set("WWW-Authenticate", str)
}

// Error returns the internal error string for this authChallenge.
func (ac *tokenAuthChallenge) Error() string {
	return ac.err.Error()
}

// SetHeaders sets the bearer challenge header on the response.
func (ac *tokenAuthChallenge) SetHeaders(w http.ResponseWriter) {
	// WWW-Authenticate response challenge header.
	// See https://docs.docker.com/registry/spec/auth/token/#/how-to-authenticate and https://tools.ietf.org/html/rfc6750#section-3
	str := fmt.Sprintf("Bearer realm=%q", ac.realm)
	if ac.service != "" {
		str += fmt.Sprintf(",service=%q", ac.service)
	}
	w.Header().Set("WWW-Authenticate", str)
}

// wrapErr wraps errors related to authorization in an authChallenge error that will present a WWW-Authenticate challenge response
func (ac *AccessController) wrapErr(ctx context.Context, err error) error {
	switch err {
	case ErrTokenRequired:
		// Challenge for errors that involve missing tokens
		if ac.tokenRealm == nil {
			// Send the basic challenge if we don't have a place to redirect
			return &authChallenge{realm: ac.realm, err: err}
		}

		if len(ac.tokenRealm.Scheme) > 0 && len(ac.tokenRealm.Host) > 0 {
			// Redirect to token auth if we've been given an absolute URL
			return &tokenAuthChallenge{realm: ac.tokenRealm.String(), err: err}
		}

		// Auto-detect scheme/host from request
		req, reqErr := context.GetRequest(ctx)
		if reqErr != nil {
			return reqErr
		}
		scheme, host := httprequest.SchemeHost(req)
		tokenRealmCopy := *ac.tokenRealm
		if len(tokenRealmCopy.Scheme) == 0 {
			tokenRealmCopy.Scheme = scheme
		}
		if len(tokenRealmCopy.Host) == 0 {
			tokenRealmCopy.Host = host
		}
		return &tokenAuthChallenge{realm: tokenRealmCopy.String(), err: err}
	case ErrTokenInvalid, ErrOpenShiftAccessDenied:
		// Challenge for errors that involve tokens or access denied
		return &authChallenge{realm: ac.realm, err: err}
	case ErrNamespaceRequired, ErrUnsupportedAction, ErrUnsupportedResource:
		// Malformed or unsupported request, no challenge
		return err
	default:
		// By default, just return the error, this gets surfaced as a bad request / internal error, but no challenge
		return err
	}
}

// Authorized handles checking whether the given request is authorized
// for actions on resources allowed by openshift.
// Sources of access records:
//   origin/pkg/cmd/dockerregistry/dockerregistry.go#Execute
//   docker/distribution/registry/handlers/app.go#appendAccessRecords
func (ac *AccessController) Authorized(ctx context.Context, accessRecords ...registryauth.Access) (context.Context, error) {
	req, err := context.GetRequest(ctx)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}

	bearerToken, err := getOpenShiftAPIToken(ctx, req)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}

	osClient, err := ac.registryClient.ClientFromToken(bearerToken)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}

	// In case of docker login, hits endpoint /v2
	if len(bearerToken) > 0 && !isMetricsBearerToken(ctx, bearerToken) {
		user, userid, err := verifyOpenShiftUser(ctx, osClient)
		if err != nil {
			return nil, ac.wrapErr(ctx, err)
		}
		ctx = WithUserInfoLogger(ctx, user, userid)
	} else {
		ctx = WithUserInfoLogger(ctx, defaultUserName, "")
	}

	if ac.auditLog {
		// TODO: setup own log formatter.
		ctx = audit.WithLogger(ctx, audit.GetLogger(ctx))
	}

	// pushChecks remembers which ns/name pairs had push access checks done
	pushChecks := map[string]bool{}
	// possibleCrossMountErrors holds errors which may be related to cross mount errors
	possibleCrossMountErrors := deferredErrors{}

	verifiedPrune := false

	// Validate all requested accessRecords
	// Only return failure errors from this loop. Success should continue to validate all records
	for _, access := range accessRecords {
		context.GetLogger(ctx).Debugf("Origin auth: checking for access to %s:%s:%s", access.Resource.Type, access.Resource.Name, access.Action)

		switch access.Resource.Type {
		case "repository":
			imageStreamNS, imageStreamName, err := getNamespaceName(access.Resource.Name)
			if err != nil {
				return nil, ac.wrapErr(ctx, err)
			}

			verb := ""
			switch access.Action {
			case "push":
				verb = "update"
				pushChecks[imageStreamNS+"/"+imageStreamName] = true
			case "pull":
				verb = "get"
			case "*":
				verb = "prune"
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}

			switch verb {
			case "prune":
				if verifiedPrune {
					continue
				}
				if err := verifyPruneAccess(ctx, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
				verifiedPrune = true
			default:
				if err := verifyImageStreamAccess(ctx, imageStreamNS, imageStreamName, verb, osClient); err != nil {
					if access.Action != "pull" {
						return nil, ac.wrapErr(ctx, err)
					}
					possibleCrossMountErrors.Add(imageStreamNS, imageStreamName, ac.wrapErr(ctx, err))
				}
			}

		case "signature":
			namespace, name, err := getNamespaceName(access.Resource.Name)
			if err != nil {
				return nil, ac.wrapErr(ctx, err)
			}
			switch access.Action {
			case "get":
				if err := verifyImageStreamAccess(ctx, namespace, name, access.Action, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			case "put":
				if err := verifyImageSignatureAccess(ctx, namespace, name, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			}

		case "metrics":
			switch access.Action {
			case "get":
				if !isMetricsBearerToken(ctx, bearerToken) {
					return nil, ac.wrapErr(ctx, ErrOpenShiftAccessDenied)
				}
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}

		case "admin":
			switch access.Action {
			case "prune":
				if verifiedPrune {
					continue
				}
				if err := verifyPruneAccess(ctx, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
				verifiedPrune = true
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}
		default:
			return nil, ac.wrapErr(ctx, ErrUnsupportedResource)
		}
	}

	// deal with any possible cross-mount errors
	for namespaceAndName, err := range possibleCrossMountErrors {
		// If we have no push requests, this can't be a cross-mount request, so error
		if len(pushChecks) == 0 {
			return nil, err
		}
		// If we also requested a push to this ns/name, this isn't a cross-mount request, so error
		if pushChecks[namespaceAndName] {
			return nil, err
		}
	}

	// Conditionally add auth errors we want to handle later to the context
	if !possibleCrossMountErrors.Empty() {
		context.GetLogger(ctx).Debugf("Origin auth: deferring errors: %#v", possibleCrossMountErrors)
		ctx = withDeferredErrors(ctx, possibleCrossMountErrors)
	}
	// Always add a marker to the context so we know auth was run
	ctx = withAuthPerformed(ctx)

	return withUserClient(ctx, osClient), nil
}

func getOpenShiftAPIToken(ctx context.Context, req *http.Request) (string, error) {
	token := ""

	authParts := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(authParts) != 2 {
		return "", ErrTokenRequired
	}

	switch strings.ToLower(authParts[0]) {
	case "bearer":
		// This is either a direct API token, or a token issued by our docker token handler
		token = authParts[1]
		// Recognize the token issued to anonymous users by our docker token handler
		if token == anonymousToken {
			token = ""
		}

	case "basic":
		_, password, ok := req.BasicAuth()
		if !ok || len(password) == 0 {
			return "", ErrTokenInvalid
		}
		token = password

	default:
		return "", ErrTokenRequired
	}

	return token, nil
}

func verifyOpenShiftUser(ctx context.Context, c client.UsersInterfacer) (string, string, error) {
	userInfo, err := c.Users().Get("~", metav1.GetOptions{})
	if err != nil {
		context.GetLogger(ctx).Errorf("Get user failed with error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return "", "", ErrOpenShiftAccessDenied
		}
		return "", "", err
	}

	return userInfo.GetName(), string(userInfo.GetUID()), nil
}

func verifyWithSAR(ctx context.Context, resource, namespace, name, verb string, c client.SelfSubjectAccessReviewsNamespacer) error {
	sar := authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     imageapi.GroupName,
				Resource:  resource,
				Name:      name,
			},
		},
	}
	response, err := c.SelfSubjectAccessReviews().Create(&sar)
	if err != nil {
		context.GetLogger(ctx).Errorf("OpenShift client error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return ErrOpenShiftAccessDenied
		}
		return err
	}

	if !response.Status.Allowed {
		context.GetLogger(ctx).Errorf("OpenShift access denied: %s", response.Status.Reason)
		return ErrOpenShiftAccessDenied
	}

	return nil
}

func verifyImageStreamAccess(ctx context.Context, namespace, imageRepo, verb string, c client.SelfSubjectAccessReviewsNamespacer) error {
	return verifyWithSAR(ctx, "imagestreams/layers", namespace, imageRepo, verb, c)
}

func verifyImageSignatureAccess(ctx context.Context, namespace, imageRepo string, c client.SelfSubjectAccessReviewsNamespacer) error {
	return verifyWithSAR(ctx, "imagesignatures", namespace, imageRepo, "create", c)
}

func verifyPruneAccess(ctx context.Context, c client.SelfSubjectAccessReviewsNamespacer) error {
	sar := authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Verb:     "delete",
				Group:    imageapi.GroupName,
				Resource: "images",
			},
		},
	}
	response, err := c.SelfSubjectAccessReviews().Create(&sar)
	if err != nil {
		context.GetLogger(ctx).Errorf("OpenShift client error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return ErrOpenShiftAccessDenied
		}
		return err
	}
	if !response.Status.Allowed {
		context.GetLogger(ctx).Errorf("OpenShift access denied: %s", response.Status.Reason)
		return ErrOpenShiftAccessDenied
	}
	return nil
}

func isMetricsBearerToken(ctx context.Context, token string) bool {
	config := ConfigurationFrom(ctx)
	if config.Metrics.Enabled {
		return config.Metrics.Secret == token
	}
	return false
}
