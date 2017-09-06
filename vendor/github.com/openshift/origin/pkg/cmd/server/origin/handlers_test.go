package origin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	metainternal "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	apifilters "k8s.io/apiserver/pkg/endpoints/filters"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	apiserver "k8s.io/apiserver/pkg/server"
	kapi "k8s.io/kubernetes/pkg/api"

	authenticationapi "github.com/openshift/origin/pkg/auth/api"
	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	serverhandlers "github.com/openshift/origin/pkg/cmd/server/handlers"
	kubernetes "github.com/openshift/origin/pkg/cmd/server/kubernetes/master"
	userapi "github.com/openshift/origin/pkg/user/apis/user"
)

type impersonateAuthorizer struct{}

func (impersonateAuthorizer) Authorize(a authorizer.Attributes) (allowed bool, reason string, err error) {
	user := a.GetUser()
	if user == nil {
		return false, "missing user", nil
	}

	switch {
	case user.GetName() == "system:admin":
		return true, "", nil

	case user.GetName() == "tester":
		return false, "", fmt.Errorf("works on my machine")

	case user.GetName() == "deny-me":
		return false, "denied", nil
	}

	if len(user.GetGroups()) > 0 && user.GetGroups()[0] == "wheel" && a.GetVerb() == "impersonate" && a.GetResource() == "systemusers" {
		return true, "", nil
	}

	if len(user.GetGroups()) > 0 && user.GetGroups()[0] == "sa-impersonater" && a.GetVerb() == "impersonate" && a.GetResource() == "serviceaccounts" {
		return true, "", nil
	}

	if len(user.GetGroups()) > 0 && user.GetGroups()[0] == "regular-impersonater" && a.GetVerb() == "impersonate" && a.GetResource() == "users" {
		return true, "", nil
	}

	if len(user.GetGroups()) > 1 && user.GetGroups()[1] == "group-impersonater" && a.GetVerb() == "impersonate" && a.GetResource() == "groups" {
		return true, "", nil
	}
	if len(user.GetGroups()) > 1 && user.GetGroups()[1] == "system-group-impersonater" && a.GetVerb() == "impersonate" && a.GetResource() == "systemgroups" {
		return true, "", nil
	}

	return false, "deny by default", nil
}

func (impersonateAuthorizer) GetAllowedSubjects(attributes authorizer.Attributes) (sets.String, sets.String, error) {
	return nil, nil, nil
}

type groupCache struct {
}

func (*groupCache) ListGroups(ctx apirequest.Context, options *metainternal.ListOptions) (*userapi.GroupList, error) {
	return &userapi.GroupList{}, nil
}
func (*groupCache) GetGroup(ctx apirequest.Context, name string, options *metav1.GetOptions) (*userapi.Group, error) {
	return nil, nil
}
func (*groupCache) CreateGroup(ctx apirequest.Context, group *userapi.Group) (*userapi.Group, error) {
	return nil, nil
}
func (*groupCache) UpdateGroup(ctx apirequest.Context, group *userapi.Group) (*userapi.Group, error) {
	return nil, nil
}
func (*groupCache) DeleteGroup(ctx apirequest.Context, name string) error {
	return nil
}
func (*groupCache) WatchGroups(ctx apirequest.Context, options *metainternal.ListOptions) (watch.Interface, error) {
	return watch.NewFake(), nil
}

func TestImpersonationFilter(t *testing.T) {
	testCases := []struct {
		name                string
		user                user.Info
		impersonationString string
		impersonationGroups []string
		expectedUser        user.Info
		expectedCode        int
	}{
		{
			name: "not-impersonating",
			user: &user.DefaultInfo{
				Name: "tester",
			},
			expectedUser: &user.DefaultInfo{
				Name: "tester",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "impersonating-error",
			user: &user.DefaultInfo{
				Name: "tester",
			},
			impersonationString: "anyone",
			expectedUser: &user.DefaultInfo{
				Name: "tester",
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "disallowed-group",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel"},
			},
			impersonationString: "system:admin",
			impersonationGroups: []string{"some-group"},
			expectedUser: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel"},
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "disallowed-system-group",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "group-impersonater"},
			},
			impersonationString: "system:admin",
			impersonationGroups: []string{"some-group", "system:group"},
			expectedUser: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "group-impersonater"},
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "disallowed-group-2",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "system-group-impersonater"},
			},
			impersonationString: "system:admin",
			impersonationGroups: []string{"some-group", "system:group"},
			expectedUser: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "system-group-impersonater"},
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "allowed-group",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "group-impersonater"},
			},
			impersonationString: "system:admin",
			impersonationGroups: []string{"some-group"},
			expectedUser: &user.DefaultInfo{
				Name:   "system:admin",
				Groups: []string{"some-group"},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "allowed-system-group",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel", "system-group-impersonater"},
			},
			impersonationString: "system:admin",
			impersonationGroups: []string{"some-system:group"},
			expectedUser: &user.DefaultInfo{
				Name:   "system:admin",
				Groups: []string{"some-system:group"},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "allowed-systemusers-impersonation",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"wheel"},
			},
			impersonationString: "system:admin",
			expectedUser: &user.DefaultInfo{
				Name:   "system:admin",
				Groups: []string{"system:authenticated"},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "allowed-users-impersonation",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"regular-impersonater"},
			},
			impersonationString: "tester",
			expectedUser: &user.DefaultInfo{
				Name:   "tester",
				Groups: []string{"system:authenticated", "system:authenticated:oauth"},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "disallowed-impersonating",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"sa-impersonater"},
			},
			impersonationString: "tester",
			expectedUser: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"sa-impersonater"},
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "allowed-sa-impersonating",
			user: &user.DefaultInfo{
				Name:   "dev",
				Groups: []string{"sa-impersonater"},
			},
			impersonationString: "system:serviceaccount:foo:default",
			expectedUser: &user.DefaultInfo{
				Name:   "system:serviceaccount:foo:default",
				Groups: []string{"system:serviceaccounts", "system:serviceaccounts:foo", "system:authenticated"},
			},
			expectedCode: http.StatusOK,
		},
	}

	config := MasterConfig{}
	config.RequestContextMapper = apirequest.NewRequestContextMapper()
	config.Authorizer = impersonateAuthorizer{}
	var ctx apirequest.Context
	var actualUser user.Info
	var lock sync.Mutex

	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		currentCtx, _ := config.RequestContextMapper.Get(req)
		user, exists := apirequest.UserFrom(currentCtx)
		if !exists {
			actualUser = nil
			return
		}

		actualUser = user
	})
	handler := func(delegate http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Recovered %v", r)
				}
			}()
			lock.Lock()
			defer lock.Unlock()
			config.RequestContextMapper.Update(req, ctx)
			currentCtx, _ := config.RequestContextMapper.Get(req)

			user, exists := apirequest.UserFrom(currentCtx)
			if !exists {
				actualUser = nil
				return
			} else {
				actualUser = user
			}

			delegate.ServeHTTP(w, req)
		})
	}(serverhandlers.ImpersonationFilter(doNothingHandler, config.Authorizer, fakeGroupCache{}, config.RequestContextMapper))
	handler = apirequest.WithRequestContext(handler, config.RequestContextMapper)

	server := httptest.NewServer(handler)
	defer server.Close()

	for _, tc := range testCases {
		func() {
			lock.Lock()
			defer lock.Unlock()
			ctx = apirequest.WithUser(apirequest.NewContext(), tc.user)
		}()

		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		req.Header.Add(authenticationapi.ImpersonateUserHeader, tc.impersonationString)
		for _, group := range tc.impersonationGroups {
			req.Header.Add(authenticationapi.ImpersonateGroupHeader, group)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
		if resp.StatusCode != tc.expectedCode {
			t.Errorf("%s: expected %v, actual %v", tc.name, tc.expectedCode, resp.StatusCode)
			continue
		}

		if !reflect.DeepEqual(actualUser, tc.expectedUser) {
			t.Errorf("%s: expected %#v, actual %#v", tc.name, tc.expectedUser, actualUser)
			continue
		}
	}
}

type fakeGroupCache struct{}

func (fakeGroupCache) GroupsFor(_ string) ([]*userapi.Group, error) {
	return []*userapi.Group{}, nil
}

var (
	currentOCKubeResources                 = "oc/v1.2.0 (linux/amd64) kubernetes/bc4550d"
	currentOCOriginResources               = "oc/v1.1.3 (linux/amd64) openshift/b348c2f"
	currentOpenshiftKubectlKubeResources   = "openshift/v1.2.0 (linux/amd64) kubernetes/bc4550d"
	currentOpenshiftKubectlOriginResources = "openshift/v1.1.3 (linux/amd64) openshift/b348c2f"
	currentOADMKubeResources               = "oadm/v1.2.0 (linux/amd64) kubernetes/bc4550d"
	currentOADMOriginResources             = "oadm/v1.1.3 (linux/amd64) openshift/b348c2f"
	currentVersionUserAgents               = []string{
		currentOCKubeResources, currentOCOriginResources, currentOpenshiftKubectlKubeResources, currentOpenshiftKubectlOriginResources, currentOADMKubeResources, currentOADMOriginResources}

	olderOCKubeResources                 = "oc/v1.1.10 (linux/amd64) kubernetes/bc4550d"
	olderOCOriginResources               = "oc/v1.1.1 (linux/amd64) openshift/b348c2f"
	oldestOCOriginResources              = "oc/v1.0.1 (linux/amd64) openshift/b348c2f"
	olderOpenshiftKubectlKubeResources   = "openshift/v1.1.10 (linux/amd64) kubernetes/bc4550d"
	olderOpenshiftKubectlOriginResources = "openshift/v1.1.1 (linux/amd64) openshift/b348c2f"
	olderOADMKubeResources               = "oadm/v1.1.10 (linux/amd64) kubernetes/bc4550d"
	olderOADMOriginResources             = "oadm/v1.1.1 (linux/amd64) openshift/b348c2f"
	olderVersionUserAgents               = []string{
		olderOCKubeResources, olderOCOriginResources, oldestOCOriginResources, olderOpenshiftKubectlKubeResources, olderOpenshiftKubectlOriginResources, olderOADMKubeResources, olderOADMOriginResources}

	newerOCKubeResources                 = "oc/v1.2.1 (linux/amd64) kubernetes/bc4550d"
	newerOCOriginResources               = "oc/v1.1.4 (linux/amd64) openshift/b348c2f"
	newerOpenshiftKubectlKubeResources   = "openshift/v1.2.1 (linux/amd64) kubernetes/bc4550d"
	newerOpenshiftKubectlOriginResources = "openshift/v1.1.4 (linux/amd64) openshift/b348c2f"
	newerOADMKubeResources               = "oadm/v1.2.1 (linux/amd64) kubernetes/bc4550d"
	newerOADMOriginResources             = "oadm/v1.1.4 (linux/amd64) openshift/b348c2f"
	newerVersionUserAgents               = []string{
		newerOCKubeResources, newerOCOriginResources, newerOpenshiftKubectlKubeResources, newerOpenshiftKubectlOriginResources, newerOADMKubeResources, newerOADMOriginResources}

	notOCVersion = "something else"

	openshiftServerVersion = `v1\.1\.3`
	kubeServerVersion      = `v1\.2\.0`
)

// variants I know I have to worry about
// 1. oc kube resources: oc/v1.2.0 (linux/amd64) kubernetes/bc4550d
// 2. oc openshift resources: oc/v1.1.3 (linux/amd64) openshift/b348c2f
// 3. openshift kubectl kube resources:  openshift/v1.2.0 (linux/amd64) kubernetes/bc4550d
// 4. openshift kubectl openshift resources: openshift/v1.1.3 (linux/amd64) openshift/b348c2f
// 5. oadm kube resources: oadm/v1.2.0 (linux/amd64) kubernetes/bc4550d
// 6. oadm openshift resources: oadm/v1.1.3 (linux/amd64) openshift/b348c2f
// 7. openshift cli kube resources: openshift/v1.2.0 (linux/amd64) kubernetes/bc4550d
// 8. openshift cli openshift resources: openshift/v1.1.3 (linux/amd64) openshift/b348c2f
// var (
// 	kubeStyleUserAgent      = regexp.MustCompile(`\w+/v([\w\.]+) \(.+/.+\) kubernetes/\w{7}`)
// 	openshiftStyleUserAgent = regexp.MustCompile(`\w+/v([\w\.]+) \(.+/.+\) openshift/\w{7}`)
// )

type versionSkewTestCase struct {
	name           string
	userAgents     []string
	failureMessage string
	methods        []string
}

func (tc versionSkewTestCase) Run(url string, t *testing.T) {
	for _, method := range tc.methods {
		for _, userAgent := range tc.userAgents {
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tc.name, err)
				return
			}
			req.Header.Add("User-Agent", userAgent)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tc.name, err)
				return
			}
			if len(tc.failureMessage) == 0 {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s: %s: unexpected status: %v", tc.name, userAgent, resp.StatusCode)
					return
				}

			} else {
				if resp.StatusCode != http.StatusForbidden {
					t.Errorf("%s: %s: unexpected status: %v", tc.name, userAgent, resp.StatusCode)
					return
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tc.name, err)
					return
				}

				if !strings.Contains(string(body), tc.failureMessage) {
					t.Errorf("%s: expected %v, got %v", tc.name, tc.failureMessage, string(body))
					return
				}
			}
		}
	}

}

func TestVersionSkewFilterDenyOld(t *testing.T) {
	verbs := []string{"PATCH", "POST"}
	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	config := MasterConfig{}
	config.Options.PolicyConfig.UserAgentMatchingConfig.DeniedClients = []configapi.UserAgentDenyRule{
		{UserAgentMatchRule: configapi.UserAgentMatchRule{Regex: `\w+/v1\.1\.10 \(.+/.+\) kubernetes/\w{7}`, HTTPVerbs: verbs}, RejectionMessage: "rejected for reasons!"},
		{UserAgentMatchRule: configapi.UserAgentMatchRule{Regex: `\w+/v(?:(?:1\.1\.1)|(?:1\.0\.1)) \(.+/.+\) openshift/\w{7}`, HTTPVerbs: verbs}, RejectionMessage: "rejected for reasons!"},
	}
	requestContextMapper := apirequest.NewRequestContextMapper()
	handler := config.versionSkewFilter(doNothingHandler, requestContextMapper)
	server := httptest.NewServer(testHandlerChain(handler, requestContextMapper))
	defer server.Close()

	testCases := []versionSkewTestCase{
		{
			name:       "missing",
			userAgents: []string{""},
			methods:    verbs,
		},
		{
			name:       "not oc",
			userAgents: []string{notOCVersion},
			methods:    verbs,
		},
		{
			name:           "older",
			userAgents:     olderVersionUserAgents,
			failureMessage: "rejected for reasons!",
			methods:        verbs,
		},
		{
			name:       "newer",
			userAgents: newerVersionUserAgents,
			methods:    verbs,
		},
		{
			name:       "exact",
			userAgents: currentVersionUserAgents,
			methods:    verbs,
		},
	}

	for _, tc := range testCases {
		tc.Run(server.URL+"/api/v1/namespaces", t)
	}
}

func TestVersionSkewFilterDenySkewed(t *testing.T) {
	verbs := []string{"PUT", "DELETE"}
	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	config := MasterConfig{}
	config.Options.PolicyConfig.UserAgentMatchingConfig.RequiredClients = []configapi.UserAgentMatchRule{
		{Regex: `\w+/` + kubeServerVersion + ` \(.+/.+\) kubernetes/\w{7}`, HTTPVerbs: verbs},
		{Regex: `\w+/` + openshiftServerVersion + ` \(.+/.+\) openshift/\w{7}`, HTTPVerbs: verbs},
	}
	config.Options.PolicyConfig.UserAgentMatchingConfig.DefaultRejectionMessage = "rejected for reasons!"
	requestContextMapper := apirequest.NewRequestContextMapper()
	handler := config.versionSkewFilter(doNothingHandler, requestContextMapper)
	server := httptest.NewServer(testHandlerChain(handler, requestContextMapper))
	defer server.Close()

	testCases := []versionSkewTestCase{
		{
			name:           "missing",
			userAgents:     []string{""},
			failureMessage: "rejected for reasons!",
			methods:        verbs,
		},
		{
			name:           "not oc",
			userAgents:     []string{notOCVersion},
			failureMessage: "rejected for reasons!",
			methods:        verbs,
		},
		{
			name:           "older",
			userAgents:     olderVersionUserAgents,
			failureMessage: "rejected for reasons!",
			methods:        verbs,
		},
		{
			name:           "newer",
			userAgents:     newerVersionUserAgents,
			failureMessage: "rejected for reasons!",
			methods:        verbs,
		},
		{
			name:       "current",
			userAgents: currentVersionUserAgents,
			methods:    verbs,
		},
	}

	for _, tc := range testCases {
		tc.Run(server.URL+"/api/v1/namespaces", t)
	}
}

func TestVersionSkewFilterSkippedOnNonAPIRequest(t *testing.T) {
	verbs := []string{"PUT", "DELETE"}
	doNothingHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	config := MasterConfig{}
	config.Options.PolicyConfig.UserAgentMatchingConfig.RequiredClients = []configapi.UserAgentMatchRule{
		{Regex: `\w+/` + kubeServerVersion + ` \(.+/.+\) kubernetes/\w{7}`, HTTPVerbs: verbs},
		{Regex: `\w+/` + openshiftServerVersion + ` \(.+/.+\) openshift/\w{7}`, HTTPVerbs: verbs},
	}
	config.Options.PolicyConfig.UserAgentMatchingConfig.DefaultRejectionMessage = "rejected for reasons!"

	requestContextMapper := apirequest.NewRequestContextMapper()
	handler := config.versionSkewFilter(doNothingHandler, requestContextMapper)
	server := httptest.NewServer(testHandlerChain(handler, requestContextMapper))
	defer server.Close()

	testCases := []versionSkewTestCase{
		{
			name:       "missing",
			userAgents: []string{""},
			methods:    verbs,
		},
		{
			name:       "not oc",
			userAgents: []string{notOCVersion},
			methods:    verbs,
		},
		{
			name:       "older",
			userAgents: olderVersionUserAgents,
			methods:    verbs,
		},
		{
			name:       "newer",
			userAgents: newerVersionUserAgents,
			methods:    verbs,
		},
		{
			name:       "current",
			userAgents: currentVersionUserAgents,
			methods:    verbs,
		},
	}

	for _, tc := range testCases {
		tc.Run(server.URL+"/api/v1", t)
	}
}

func testHandlerChain(handler http.Handler, contextMapper apirequest.RequestContextMapper) http.Handler {
	kgenericconfig := apiserver.NewConfig(kapi.Codecs)
	kgenericconfig.LegacyAPIGroupPrefixes = kubernetes.LegacyAPIGroupPrefixes

	handler = apifilters.WithRequestInfo(handler, apiserver.NewRequestInfoResolver(kgenericconfig), contextMapper)
	handler = apirequest.WithRequestContext(handler, contextMapper)
	return handler
}
