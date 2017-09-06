package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/registry/handlers"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"

	registryclient "github.com/openshift/origin/pkg/dockerregistry/server/client"
	registrytest "github.com/openshift/origin/pkg/dockerregistry/testutil"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	imageapiv1 "github.com/openshift/origin/pkg/image/apis/image/v1"
)

func createTestRegistryServer(t *testing.T, ctx context.Context) *httptest.Server {
	ctx = WithTestPassthroughToUpstream(ctx, true)

	// pullthrough middleware will attempt to pull from this registry instance
	remoteRegistryApp := handlers.NewApp(ctx, &configuration.Configuration{
		Loglevel: "debug",
		Auth: map[string]configuration.Parameters{
			fakeAuthorizerName: {"realm": fakeAuthorizerName},
		},
		Storage: configuration.Storage{
			"inmemory": configuration.Parameters{},
			"cache": configuration.Parameters{
				"blobdescriptor": "inmemory",
			},
			"delete": configuration.Parameters{
				"enabled": true,
			},
			"maintenance": configuration.Parameters{
				"uploadpurging": map[interface{}]interface{}{
					"enabled": false,
				},
			},
		},
	})

	remoteRegistryServer := httptest.NewServer(remoteRegistryApp)

	serverURL, err := url.Parse(remoteRegistryServer.URL)
	if err != nil {
		t.Fatalf("error parsing server url: %v", err)
	}
	os.Setenv("OPENSHIFT_DEFAULT_REGISTRY", serverURL.Host)

	return remoteRegistryServer
}

func TestPullthroughManifests(t *testing.T) {
	namespace := "fuser"
	repo := "zapp"
	repoName := fmt.Sprintf("%s/%s", namespace, repo)
	tag := "latest"

	installFakeAccessController(t)
	setPassthroughBlobDescriptorServiceFactory()

	remoteRegistryServer := createTestRegistryServer(t, context.Background())
	defer remoteRegistryServer.Close()

	serverURL, err := url.Parse(remoteRegistryServer.URL)
	if err != nil {
		t.Fatalf("error parsing server url: %v", err)
	}

	ms1dgst, ms1canonical, _, ms1manifest, err := registrytest.CreateAndUploadTestManifest(
		registrytest.ManifestSchema1, 2, serverURL, nil, repoName, "schema1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ms1payload, err := ms1manifest.Payload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("ms1dgst=%s, ms1manifest: %s", ms1dgst, ms1canonical)

	image, err := registrytest.NewImageForManifest(repoName, string(ms1payload), "", false)
	if err != nil {
		t.Fatal(err)
	}
	image.DockerImageReference = fmt.Sprintf("%s/%s/%s@%s", serverURL.Host, namespace, repo, image.Name)
	image.DockerImageManifest = ""

	fos, client, imageClient := registrytest.NewFakeOpenShiftWithClient()
	registrytest.AddImageStream(t, fos, namespace, repo, map[string]string{
		imageapi.InsecureRepositoryAnnotation: "true",
	})
	registrytest.AddImage(t, fos, image, namespace, repo, tag)

	for _, tc := range []struct {
		name                  string
		manifestDigest        digest.Digest
		localData             map[digest.Digest]distribution.Manifest
		expectedLocalCalls    map[string]int
		expectedError         bool
		expectedNotFoundError bool
	}{
		{
			name:           "manifest digest from local store",
			manifestDigest: ms1dgst,
			localData: map[digest.Digest]distribution.Manifest{
				ms1dgst: ms1manifest,
			},
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},
		{
			name:           "manifest served from remote repository",
			manifestDigest: digest.Digest(image.Name),
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},
		{
			name:                  "unknown manifest digest",
			manifestDigest:        unknownBlobDigest,
			expectedNotFoundError: true,
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},
	} {
		localManifestService := newTestManifestService(repoName, tc.localData)

		repo := newTestRepository(t, namespace, repo, testRepositoryOptions{
			client:            registryclient.NewFakeRegistryAPIClient(client, nil, imageClient),
			enablePullThrough: true,
		})

		ptms := &pullthroughManifestService{
			ManifestService: localManifestService,
			repo:            repo,
		}

		ctx := WithTestPassthroughToUpstream(context.Background(), false)
		manifestResult, err := ptms.Get(ctx, tc.manifestDigest)
		switch err.(type) {
		case distribution.ErrManifestUnknownRevision:
			if !tc.expectedNotFoundError {
				t.Fatalf("[%s] unexpected error: %#+v", tc.name, err)
			}
		case nil:
			if tc.expectedError || tc.expectedNotFoundError {
				t.Fatalf("[%s] unexpected successful response", tc.name)
			}
		default:
			if tc.expectedError {
				break
			}
			t.Fatalf("[%s] unexpected error: %#+v", tc.name, err)
		}

		if tc.localData != nil {
			if manifestResult != nil && manifestResult != tc.localData[tc.manifestDigest] {
				t.Fatalf("[%s] unexpected result returned", tc.name)
			}
		}

		for name, count := range localManifestService.calls {
			expectCount, exists := tc.expectedLocalCalls[name]
			if !exists {
				t.Errorf("[%s] expected no calls to method %s of local manifest service, got %d", tc.name, name, count)
			}
			if count != expectCount {
				t.Errorf("[%s] unexpected number of calls to method %s of local manifest service, got %d", tc.name, name, count)
			}
		}
	}
}

func TestPullthroughManifestInsecure(t *testing.T) {
	namespace := "fuser"
	repo := "zapp"
	repoName := fmt.Sprintf("%s/%s", namespace, repo)

	installFakeAccessController(t)
	setPassthroughBlobDescriptorServiceFactory()

	remoteRegistryServer := createTestRegistryServer(t, context.Background())
	defer remoteRegistryServer.Close()

	serverURL, err := url.Parse(remoteRegistryServer.URL)
	if err != nil {
		t.Fatalf("error parsing server url: %v", err)
	}

	ms1dgst, ms1canonical, _, ms1manifest, err := registrytest.CreateAndUploadTestManifest(
		registrytest.ManifestSchema1, 2, serverURL, nil, repoName, "schema1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ms1payload, err := ms1manifest.Payload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("ms1dgst=%s, ms1manifest: %s", ms1dgst, ms1canonical)
	ms2dgst, ms2canonical, ms2config, ms2manifest, err := registrytest.CreateAndUploadTestManifest(
		registrytest.ManifestSchema2, 2, serverURL, nil, repoName, "schema2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ms2payload, err := ms2manifest.Payload()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("ms2dgst=%s, ms2manifest: %s", ms2dgst, ms2canonical)

	ms1img, err := registrytest.NewImageForManifest(repoName, string(ms1payload), "", false)
	if err != nil {
		t.Fatal(err)
	}
	ms1img.DockerImageReference = fmt.Sprintf("%s/%s/%s@%s", serverURL.Host, namespace, repo, ms1img.Name)
	ms1img.DockerImageManifest = ""
	ms2img, err := registrytest.NewImageForManifest(repoName, string(ms2payload), ms2config, false)
	if err != nil {
		t.Fatal(err)
	}
	ms2img.DockerImageReference = fmt.Sprintf("%s/%s/%s@%s", serverURL.Host, namespace, repo, ms2img.Name)
	ms2img.DockerImageManifest = ""

	for _, tc := range []struct {
		name                string
		manifestDigest      digest.Digest
		localData           map[digest.Digest]distribution.Manifest
		fakeOpenShiftInit   func(fos *registrytest.FakeOpenShift)
		expectedManifest    distribution.Manifest
		expectedLocalCalls  map[string]int
		expectedErrorString string
	}{

		{
			name:           "fetch schema 1 with allowed insecure",
			manifestDigest: ms1dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				registrytest.AddImageStream(t, fos, namespace, repo, map[string]string{
					imageapi.InsecureRepositoryAnnotation: "true",
				})
				registrytest.AddImage(t, fos, ms1img, namespace, repo, "schema1")
			},
			expectedManifest: ms1manifest,
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},

		{
			name:           "fetch schema 2 with allowed insecure",
			manifestDigest: ms2dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				registrytest.AddImageStream(t, fos, namespace, repo, map[string]string{
					imageapi.InsecureRepositoryAnnotation: "true",
				})
				registrytest.AddImage(t, fos, ms2img, namespace, repo, "schema2")
			},
			expectedManifest: ms2manifest,
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},

		{
			name:           "explicit forbid insecure",
			manifestDigest: ms1dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				registrytest.AddImageStream(t, fos, namespace, repo, map[string]string{
					imageapi.InsecureRepositoryAnnotation: "false",
				})
				registrytest.AddImage(t, fos, ms1img, namespace, repo, "schema1")
			},
			expectedErrorString: "server gave HTTP response to HTTPS client",
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},

		{
			name:           "implicit forbid insecure",
			manifestDigest: ms1dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				registrytest.AddImageStream(t, fos, namespace, repo, nil)
				registrytest.AddImage(t, fos, ms1img, namespace, repo, "schema1")
			},
			expectedErrorString: "server gave HTTP response to HTTPS client",
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},

		{
			name:           "pullthrough from insecure tag",
			manifestDigest: ms1dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				image, err := registrytest.NewImageForManifest(repoName, string(ms1payload), "", false)
				if err != nil {
					t.Fatal(err)
				}
				image.DockerImageReference = fmt.Sprintf("%s/%s/%s@%s", serverURL.Host, namespace, repo, ms1dgst)
				image.DockerImageManifest = ""

				registrytest.AddUntaggedImage(t, fos, image)
				registrytest.AddImageStream(t, fos, namespace, repo, nil)
				registrytest.AddImageStreamTag(t, fos, ms1img, namespace, repo, &imageapiv1.TagReference{
					Name:         "schema1",
					ImportPolicy: imageapiv1.TagImportPolicy{Insecure: true},
				})
			},
			expectedManifest: ms1manifest,
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},

		{
			name:           "pull insecure if either image stream is insecure or the tag",
			manifestDigest: ms2dgst,
			fakeOpenShiftInit: func(fos *registrytest.FakeOpenShift) {
				image, err := registrytest.NewImageForManifest(repoName, string(ms2payload), ms2config, false)
				if err != nil {
					t.Fatal(err)
				}
				image.DockerImageReference = fmt.Sprintf("%s/%s/%s@%s", serverURL.Host, namespace, repo, image.Name)
				image.DockerImageManifest = ""

				registrytest.AddUntaggedImage(t, fos, image)
				registrytest.AddImageStream(t, fos, namespace, repo, map[string]string{
					imageapi.InsecureRepositoryAnnotation: "true",
				})
				registrytest.AddImageStreamTag(t, fos, image, namespace, repo, &imageapiv1.TagReference{
					Name: "schema2",
					// the value doesn't override is annotation because we cannot determine whether the
					// value is explicit or just the default
					ImportPolicy: imageapiv1.TagImportPolicy{Insecure: false},
				})
			},
			expectedManifest: ms2manifest,
			expectedLocalCalls: map[string]int{
				"Get": 1,
			},
		},
	} {
		fos, client, imageClient := registrytest.NewFakeOpenShiftWithClient()

		tc.fakeOpenShiftInit(fos)

		localManifestService := newTestManifestService(repoName, tc.localData)

		ctx := WithTestPassthroughToUpstream(context.Background(), false)
		repo := newTestRepository(t, namespace, repo, testRepositoryOptions{
			client:            registryclient.NewFakeRegistryAPIClient(client, nil, imageClient),
			enablePullThrough: true,
		})
		ctx = withRepository(ctx, repo)

		ptms := &pullthroughManifestService{
			ManifestService: localManifestService,
			repo:            repo,
		}

		manifestResult, err := ptms.Get(ctx, tc.manifestDigest)
		switch err.(type) {
		case nil:
			if len(tc.expectedErrorString) > 0 {
				t.Errorf("[%s] unexpected successful response", tc.name)
				continue
			}
		default:
			if len(tc.expectedErrorString) > 0 {
				if !strings.Contains(err.Error(), tc.expectedErrorString) {
					t.Fatalf("expected error string %q, got instead: %s (%#+v)", tc.expectedErrorString, err.Error(), err)
				}
				break
			}
			t.Fatalf("[%s] unexpected error: %#+v (%s)", tc.name, err, err.Error())
		}

		if tc.localData != nil {
			if manifestResult != nil && manifestResult != tc.localData[tc.manifestDigest] {
				t.Errorf("[%s] unexpected result returned", tc.name)
				continue
			}
		}

		registrytest.AssertManifestsEqual(t, tc.name, manifestResult, tc.expectedManifest)

		for name, count := range localManifestService.calls {
			expectCount, exists := tc.expectedLocalCalls[name]
			if !exists {
				t.Errorf("[%s] expected no calls to method %s of local manifest service, got %d", tc.name, name, count)
			}
			if count != expectCount {
				t.Errorf("[%s] unexpected number of calls to method %s of local manifest service, got %d", tc.name, name, count)
			}
		}
	}
}

func TestPullthroughManifestDockerReference(t *testing.T) {
	namespace := "user"
	repo1 := "repo1"
	repo2 := "repo2"
	tag := "latest"

	type testServer struct {
		*httptest.Server
		name    string
		touched bool
	}

	startServer := func(name string) *testServer {
		s := &testServer{
			name: name,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.touched = true
			http.Error(w, "dummy implementation", http.StatusInternalServerError)
		})

		s.Server = httptest.NewServer(handler)
		return s
	}

	dockerImageReference := func(s *testServer, rest string) string {
		serverURL, err := url.Parse(s.Server.URL)
		if err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%s/%s", serverURL.Host, rest)
	}

	server1 := startServer("server1")
	defer server1.Close()

	server2 := startServer("server2")
	defer server2.Close()

	img, err := registrytest.CreateRandomImage(namespace, "dummy")
	if err != nil {
		t.Fatal(err)
	}
	img.DockerImageManifest = ""

	image1 := *img
	image1.DockerImageReference = dockerImageReference(server1, "repo/name")

	image2 := *img
	image2.DockerImageReference = dockerImageReference(server2, "foo/bar")

	fos, client, imageClient := registrytest.NewFakeOpenShiftWithClient()
	registrytest.AddImageStream(t, fos, namespace, repo1, map[string]string{
		imageapi.InsecureRepositoryAnnotation: "true",
	})
	registrytest.AddImageStream(t, fos, namespace, repo2, map[string]string{
		imageapi.InsecureRepositoryAnnotation: "true",
	})
	registrytest.AddImage(t, fos, &image1, namespace, repo1, tag)
	registrytest.AddImage(t, fos, &image2, namespace, repo2, tag)

	for _, tc := range []struct {
		name             string
		repoName         string
		touchedServers   []*testServer
		untouchedServers []*testServer
	}{
		{
			name:             "server 1",
			repoName:         repo1,
			touchedServers:   []*testServer{server1},
			untouchedServers: []*testServer{server2},
		},
		{
			name:             "server 2",
			repoName:         repo2,
			touchedServers:   []*testServer{server2},
			untouchedServers: []*testServer{server1},
		},
	} {
		for _, s := range append(tc.touchedServers, tc.untouchedServers...) {
			s.touched = false
		}

		r := newTestRepository(t, namespace, tc.repoName, testRepositoryOptions{
			client:            registryclient.NewFakeRegistryAPIClient(client, nil, imageClient),
			enablePullThrough: true,
		})

		ptms := &pullthroughManifestService{
			ManifestService: newTestManifestService(tc.repoName, nil),
			repo:            r,
		}

		ctx := context.Background()
		ctx = withRepository(ctx, r)
		ptms.Get(ctx, digest.Digest(img.Name))

		for _, s := range tc.touchedServers {
			if !s.touched {
				t.Errorf("[%s] %s not touched", tc.name, s.name)
			}
		}

		for _, s := range tc.untouchedServers {
			if s.touched {
				t.Errorf("[%s] %s touched", tc.name, s.name)
			}
		}
	}
}

type testManifestService struct {
	name  string
	data  map[digest.Digest]distribution.Manifest
	calls map[string]int
}

var _ distribution.ManifestService = &testManifestService{}

func newTestManifestService(name string, data map[digest.Digest]distribution.Manifest) *testManifestService {
	b := make(map[digest.Digest]distribution.Manifest)
	for d, content := range data {
		b[d] = content
	}
	return &testManifestService{
		name:  name,
		data:  b,
		calls: make(map[string]int),
	}
}

func (t *testManifestService) Exists(ctx context.Context, dgst digest.Digest) (bool, error) {
	t.calls["Exists"]++
	_, exists := t.data[dgst]
	return exists, nil
}

func (t *testManifestService) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	t.calls["Get"]++
	content, exists := t.data[dgst]
	if !exists {
		return nil, distribution.ErrManifestUnknownRevision{
			Name:     t.name,
			Revision: dgst,
		}
	}
	return content, nil
}

func (t *testManifestService) Put(ctx context.Context, manifest distribution.Manifest, options ...distribution.ManifestServiceOption) (digest.Digest, error) {
	t.calls["Put"]++
	_, payload, err := manifest.Payload()
	if err != nil {
		return "", err
	}
	dgst := digest.FromBytes(payload)
	t.data[dgst] = manifest
	return dgst, nil
}

func (t *testManifestService) Delete(ctx context.Context, dgst digest.Digest) error {
	t.calls["Delete"]++
	return fmt.Errorf("method not implemented")
}

const etcdDigest = "sha256:958608f8ecc1dc62c93b6c610f3a834dae4220c9642e6e8b4e0f2b3ad7cbd238"
const etcdManifest = `{
   "schemaVersion": 1, 
   "tag": "latest", 
   "name": "coreos/etcd", 
   "architecture": "amd64", 
   "fsLayers": [
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      }, 
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      }, 
      {
         "blobSum": "sha256:2560187847cadddef806eaf244b7755af247a9dbabb90ca953dd2703cf423766"
      }, 
      {
         "blobSum": "sha256:744b46d0ac8636c45870a03830d8d82c20b75fbfb9bc937d5e61005d23ad4cfe"
      }
   ], 
   "history": [
      {
         "v1Compatibility": "{\"id\":\"fe50ac14986497fa6b5d2cc24feb4a561d01767bc64413752c0988cb70b0b8b9\",\"parent\":\"a5a18474fa96a3c6e240bc88e41de2afd236520caf904356ad9d5f8d875c3481\",\"created\":\"2015-12-30T22:29:13.967754365Z\",\"container\":\"c8d0f1a274b5f52fa5beb280775ef07cf18ec0f95e5ae42fbad01157e2614d42\",\"container_config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"ExposedPorts\":{\"2379/tcp\":{},\"2380/tcp\":{},\"4001/tcp\":{},\"7001/tcp\":{}},\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ENTRYPOINT \\u0026{[\\\"/etcd\\\"]}\"],\"Image\":\"a5a18474fa96a3c6e240bc88e41de2afd236520caf904356ad9d5f8d875c3481\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":[\"/etcd\"],\"OnBuild\":null,\"Labels\":{}},\"docker_version\":\"1.9.1\",\"config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"ExposedPorts\":{\"2379/tcp\":{},\"2380/tcp\":{},\"4001/tcp\":{},\"7001/tcp\":{}},\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"a5a18474fa96a3c6e240bc88e41de2afd236520caf904356ad9d5f8d875c3481\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":[\"/etcd\"],\"OnBuild\":null,\"Labels\":{}},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      }, 
      {
         "v1Compatibility": "{\"id\":\"a5a18474fa96a3c6e240bc88e41de2afd236520caf904356ad9d5f8d875c3481\",\"parent\":\"796d581500e960cc02095dcdeccf55db215b8e54c57e3a0b11392145ffe60cf6\",\"created\":\"2015-12-30T22:29:13.504159783Z\",\"container\":\"080708d544f85052a46fab72e701b4358c1b96cb4b805a5b2d66276fc2aaf85d\",\"container_config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"ExposedPorts\":{\"2379/tcp\":{},\"2380/tcp\":{},\"4001/tcp\":{},\"7001/tcp\":{}},\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) EXPOSE 2379/tcp 2380/tcp 4001/tcp 7001/tcp\"],\"Image\":\"796d581500e960cc02095dcdeccf55db215b8e54c57e3a0b11392145ffe60cf6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"docker_version\":\"1.9.1\",\"config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"ExposedPorts\":{\"2379/tcp\":{},\"2380/tcp\":{},\"4001/tcp\":{},\"7001/tcp\":{}},\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"796d581500e960cc02095dcdeccf55db215b8e54c57e3a0b11392145ffe60cf6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      }, 
      {
         "v1Compatibility": "{\"id\":\"796d581500e960cc02095dcdeccf55db215b8e54c57e3a0b11392145ffe60cf6\",\"parent\":\"309c960c7f875411ae2ee2bfb97b86eee5058f3dad77206dd0df4f97df8a77fa\",\"created\":\"2015-12-30T22:29:12.912813629Z\",\"container\":\"f28be899c9b8680d4cf8585e663ad20b35019db062526844e7cfef117ce9037f\",\"container_config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ADD file:e330b1da49d993059975e46560b3bd360691498b0f2f6e00f39fc160cf8d4ec3 in /\"],\"Image\":\"309c960c7f875411ae2ee2bfb97b86eee5058f3dad77206dd0df4f97df8a77fa\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"docker_version\":\"1.9.1\",\"config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"309c960c7f875411ae2ee2bfb97b86eee5058f3dad77206dd0df4f97df8a77fa\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":13502144}"
      }, 
      {
         "v1Compatibility": "{\"id\":\"309c960c7f875411ae2ee2bfb97b86eee5058f3dad77206dd0df4f97df8a77fa\",\"created\":\"2015-12-30T22:29:12.346834862Z\",\"container\":\"1b97abade59e4b5b935aede236980a54fb500cd9ee5bd4323c832c6d7b3ffc6e\",\"container_config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ADD file:74912593c6783292c4520514f5cc9313acbd1da0f46edee0fdbed2a24a264d6f in /\"],\"Image\":\"\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"docker_version\":\"1.9.1\",\"config\":{\"Hostname\":\"1b97abade59e\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":15141568}"
      }
   ], 
   "signatures": [
      {
         "header": {
            "alg": "RS256", 
            "jwk": {
               "e": "AQAB", 
               "kty": "RSA", 
               "n": "yB40ou1GMvIxYs1jhxWaeoDiw3oa0_Q2UJThUPtArvO0tRzaun9FnSphhOEHIGcezfq95jy-3MN-FIjmsWgbPHY8lVDS38fF75aCw6qkholwqjmMtUIgPNYoMrg0rLUE5RRyJ84-hKf9Fk7V3fItp1mvCTGKaS3ze-y5dTTrfbNGE7qG638Dla2Fuz-9CNgRQj0JH54o547WkKJC-pG-j0jTDr8lzsXhrZC7lJas4yc-vpt3D60iG4cW_mkdtIj52ZFEgHZ56sUj7AhnNVly0ZP9W1hmw4xEHDn9WLjlt7ivwARVeb2qzsNdguUitcI5hUQNwpOVZ_O3f1rUIL_kRw"
            }
         }, 
         "protected": "eyJmb3JtYXRUYWlsIjogIkNuMCIsICJmb3JtYXRMZW5ndGgiOiA1OTI2LCAidGltZSI6ICIyMDE2LTAxLTAyVDAyOjAxOjMzWiJ9", 
         "signature": "DrQ43UWeit-thDoRGTCP0Gd2wL5K2ecyPhHo_au0FoXwuKODja0tfwHexB9ypvFWngk-ijXuwO02x3aRIZqkWpvKLxxzxwkrZnPSje4o_VrFU4z5zwmN8sJw52ODkQlW38PURIVksOxCrb0zRl87yTAAsUAJ_4UUPNltZSLnhwy-qPb2NQ8ghgsONcBxRQrhPFiWNkxDKZ3kjvzYyrXDxTcvwK3Kk_YagZ4rCOhH1B7mAdVSiSHIvvNV5grPshw_ipAoqL2iNMsxWxLjYZl9xSJQI2asaq3fvh8G8cZ7T-OahDUos_GyhnIj39C-9ouqdJqMUYFETqbzRCR6d36CpQ"
      }
   ]
}`
