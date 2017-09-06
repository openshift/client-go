package server

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/cache"
	"github.com/docker/distribution/registry/storage/cache/memory"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/docker/libtrust"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	clientgotesting "k8s.io/client-go/testing"
	kapi "k8s.io/kubernetes/pkg/api"

	registryclient "github.com/openshift/origin/pkg/dockerregistry/server/client"
	registrytest "github.com/openshift/origin/pkg/dockerregistry/testutil"
	imageapi "github.com/openshift/origin/pkg/image/apis/image"
	imageapiv1 "github.com/openshift/origin/pkg/image/apis/image/v1"
)

const (
	// testImageLayerCount says how many layers to generate per image
	testImageLayerCount = 2
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestRepositoryBlobStat(t *testing.T) {
	quotaEnforcing = &quotaEnforcingConfig{}

	ctx := context.Background()
	// this driver holds all the testing blobs in memory during the whole test run
	driver := inmemory.New()
	// generate two images and store their blobs in the driver
	testImages, err := populateTestStorage(t, driver, true, 1, map[string]int{"nm/is:latest": 1, "nm/repo:missing-layer-links": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// generate an image and store its blobs in the driver; the resulting image will lack managed by openshift
	// annotation
	testImages, err = populateTestStorage(t, driver, false, 1, map[string]int{"nm/unmanaged:missing-layer-links": 1}, testImages)
	if err != nil {
		t.Fatal(err)
	}

	// remove layer repository links from two of the above images; keep the uploaded blobs in the global
	// blostore though
	for _, name := range []string{"nm/repo:missing-layer-links", "nm/unmanaged:missing-layer-links"} {
		repoName := strings.Split(name, ":")[0]
		for _, layer := range testImages[name][0].DockerImageLayers {
			dgst := digest.Digest(layer.Name)
			alg, hex := dgst.Algorithm(), dgst.Hex()
			err := driver.Delete(ctx, fmt.Sprintf("/docker/registry/v2/repositories/%s/_layers/%s/%s", repoName, alg, hex))
			if err != nil {
				t.Fatalf("failed to delete layer link %q from repository %q: %v", layer.Name, repoName, err)
			}
		}
	}

	// generate random images without storing its blobs in the driver
	etcdOnlyImages := map[string]*imageapiv1.Image{}
	for _, d := range []struct {
		name    string
		managed bool
	}{{"nm/is", true}, {"registry.org:5000/user/app", false}} {
		img, err := registrytest.NewImageForManifest(d.name, registrytest.SampleImageManifestSchema1, "", d.managed)
		if err != nil {
			t.Fatal(err)
		}
		etcdOnlyImages[d.name] = img
	}

	for _, tc := range []struct {
		name               string
		stat               string
		images             []imageapiv1.Image
		imageStreams       []imageapiv1.ImageStream
		pullthrough        bool
		skipAuth           bool
		deferredErrors     deferredErrors
		expectedDescriptor distribution.Descriptor
		expectedError      error
		expectedActions    []clientAction
	}{
		{
			name:               "local stat",
			stat:               "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name,
			imageStreams:       []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}},
			expectedDescriptor: testNewDescriptorForLayer(testImages["nm/is:latest"][0].DockerImageLayers[0]),
		},

		{
			name:   "blob only tagged in image stream",
			stat:   "nm/repo@" + testImages["nm/repo:missing-layer-links"][0].DockerImageLayers[1].Name,
			images: []imageapiv1.Image{*testImages["nm/repo:missing-layer-links"][0]},
			imageStreams: []imageapiv1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nm",
						Name:      "repo",
					},
					Status: imageapiv1.ImageStreamStatus{
						Tags: []imageapiv1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imageapiv1.TagEvent{
									{
										Image: testImages["nm/repo:missing-layer-links"][0].Name,
									},
								},
							},
						},
					},
				},
			},
			expectedDescriptor: testNewDescriptorForLayer(testImages["nm/repo:missing-layer-links"][0].DockerImageLayers[1]),
			expectedActions:    []clientAction{{"get", "imagestreams"}, {"get", "images"}},
		},

		{
			name:   "blob referenced only by not managed image with pullthrough on",
			stat:   "nm/unmanaged@" + testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1].Name,
			images: []imageapiv1.Image{*testImages["nm/unmanaged:missing-layer-links"][0]},
			imageStreams: []imageapiv1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nm",
						Name:      "unmanaged",
					},
					Status: imageapiv1.ImageStreamStatus{
						Tags: []imageapiv1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imageapiv1.TagEvent{
									{
										Image: testImages["nm/unmanaged:missing-layer-links"][0].Name,
									},
								},
							},
						},
					},
				},
			},
			pullthrough:        true,
			expectedDescriptor: testNewDescriptorForLayer(testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1]),
			expectedActions:    []clientAction{{"get", "imagestreams"}, {"get", "images"}},
		},

		{
			// TODO: this should err out because of missing image stream.
			// Unfortunately, it's not the case. Until we start storing layer links in etcd, we depend on
			// local layer links.
			name:               "layer link present while image stream not found",
			stat:               "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name,
			images:             []imageapiv1.Image{*testImages["nm/is:latest"][0]},
			expectedDescriptor: testNewDescriptorForLayer(testImages["nm/is:latest"][0].DockerImageLayers[0]),
		},

		{
			name:   "blob only tagged by not managed image with pullthrough off",
			stat:   "nm/repo@" + testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1].Name,
			images: []imageapiv1.Image{*testImages["nm/unmanaged:missing-layer-links"][0]},
			imageStreams: []imageapiv1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nm",
						Name:      "repo",
					},
					Status: imageapiv1.ImageStreamStatus{
						Tags: []imageapiv1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imageapiv1.TagEvent{
									{
										Image: testImages["nm/unmanaged:missing-layer-links"][0].DockerImageLayers[1].Name,
									},
								},
							},
						},
					},
				},
			},
			expectedError:   distribution.ErrBlobUnknown,
			expectedActions: []clientAction{{"get", "imagestreams"}, {"get", "images"}},
		},

		{
			name:   "blob not stored locally but referred in image stream",
			stat:   "nm/is@" + etcdOnlyImages["nm/is"].DockerImageLayers[1].Name,
			images: []imageapiv1.Image{*etcdOnlyImages["nm/is"]},
			imageStreams: []imageapiv1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nm",
						Name:      "is",
					},
					Status: imageapiv1.ImageStreamStatus{
						Tags: []imageapiv1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imageapiv1.TagEvent{
									{
										Image: etcdOnlyImages["nm/is"].Name,
									},
								},
							},
						},
					},
				},
			},
			expectedError: distribution.ErrBlobUnknown,
		},

		{
			name:   "blob does not exist",
			stat:   "nm/repo@" + etcdOnlyImages["nm/is"].DockerImageLayers[0].Name,
			images: []imageapiv1.Image{*testImages["nm/is:latest"][0]},
			imageStreams: []imageapiv1.ImageStream{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nm",
						Name:      "repo",
					},
					Status: imageapiv1.ImageStreamStatus{
						Tags: []imageapiv1.NamedTagEventList{
							{
								Tag: "latest",
								Items: []imageapiv1.TagEvent{
									{
										Image: testImages["nm/is:latest"][0].Name,
									},
								},
							},
						},
					},
				},
			},
			expectedError: distribution.ErrBlobUnknown,
		},

		{
			name:          "auth not performed",
			stat:          "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name,
			imageStreams:  []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}},
			skipAuth:      true,
			expectedError: fmt.Errorf("openshift.auth.completed missing from context"),
		},

		{
			name:           "deferred error",
			stat:           "nm/is@" + testImages["nm/is:latest"][0].DockerImageLayers[0].Name,
			imageStreams:   []imageapiv1.ImageStream{{ObjectMeta: metav1.ObjectMeta{Namespace: "nm", Name: "is"}}},
			deferredErrors: deferredErrors{"nm/is": ErrOpenShiftAccessDenied},
			expectedError:  ErrOpenShiftAccessDenied,
		},
	} {
		ref, err := reference.Parse(tc.stat)
		if err != nil {
			t.Errorf("[%s] failed to parse blob reference %q: %v", tc.name, tc.stat, err)
			continue
		}
		canonical, ok := ref.(reference.Canonical)
		if !ok {
			t.Errorf("[%s] not a canonical reference %q", tc.name, ref.String())
			continue
		}

		cachedLayers, err = newDigestToRepositoryCache(defaultDigestToRepositoryCacheSize)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		if !tc.skipAuth {
			ctx = withAuthPerformed(ctx)
		}
		if tc.deferredErrors != nil {
			ctx = withDeferredErrors(ctx, tc.deferredErrors)
		}

		fos, client, imageClient := registrytest.NewFakeOpenShiftWithClient()

		for _, is := range tc.imageStreams {
			_, err = fos.CreateImageStream(is.Namespace, &is)
			if err != nil {
				t.Fatal(err)
			}
		}

		for _, image := range tc.images {
			_, err = fos.CreateImage(&image)
			if err != nil {
				t.Fatal(err)
			}
		}

		reg, err := newTestRegistry(ctx, registryclient.NewFakeRegistryAPIClient(client, nil, imageClient), driver, defaultBlobRepositoryCacheTTL, tc.pullthrough, true)
		if err != nil {
			t.Errorf("[%s] unexpected error: %v", tc.name, err)
			continue
		}

		repo, err := reg.Repository(ctx, canonical)
		if err != nil {
			t.Errorf("[%s] unexpected error: %v", tc.name, err)
			continue
		}

		desc, err := repo.Blobs(ctx).Stat(ctx, canonical.Digest())
		if err != nil && tc.expectedError == nil {
			t.Errorf("[%s] got unexpected stat error: %v", tc.name, err)
			continue
		}
		if err == nil && tc.expectedError != nil {
			t.Errorf("[%s] got unexpected non-error", tc.name)
			continue
		}
		if !reflect.DeepEqual(err, tc.expectedError) {
			t.Errorf("[%s] got unexpected error: %s", tc.name, diff.ObjectGoPrintDiff(err, tc.expectedError))
			continue
		}
		if tc.expectedError == nil && !reflect.DeepEqual(desc, tc.expectedDescriptor) {
			t.Errorf("[%s] got unexpected descriptor: %s", tc.name, diff.ObjectGoPrintDiff(desc, tc.expectedDescriptor))
		}

		compareActions(t, tc.name, imageClient.Actions(), tc.expectedActions)
	}
}

func TestRepositoryBlobStatCacheEviction(t *testing.T) {
	const blobRepoCacheTTL = time.Millisecond * 500

	quotaEnforcing = &quotaEnforcingConfig{}
	ctx := withAuthPerformed(context.Background())

	// this driver holds all the testing blobs in memory during the whole test run
	driver := inmemory.New()
	// generate two images and store their blobs in the driver
	testImages, err := populateTestStorage(t, driver, true, 1, map[string]int{"nm/is:latest": 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	testImage := testImages["nm/is:latest"][0]

	blob1Desc := testNewDescriptorForLayer(testImage.DockerImageLayers[0])
	blob1Dgst := blob1Desc.Digest
	blob2Desc := testNewDescriptorForLayer(testImage.DockerImageLayers[1])
	blob2Dgst := blob2Desc.Digest

	// remove repo layer repo link of the image's second blob
	alg, hex := blob2Dgst.Algorithm(), blob2Dgst.Hex()
	err = driver.Delete(ctx, fmt.Sprintf("/docker/registry/v2/repositories/%s/_layers/%s/%s", "nm/is", alg, hex))

	cachedLayers, err = newDigestToRepositoryCache(defaultDigestToRepositoryCacheSize)
	if err != nil {
		t.Fatal(err)
	}

	fos, client, imageClient := registrytest.NewFakeOpenShiftWithClient()
	registrytest.AddImageStream(t, fos, "nm", "is", nil)
	registrytest.AddImage(t, fos, testImage, "nm", "is", "latest")

	reg, err := newTestRegistry(ctx, registryclient.NewFakeRegistryAPIClient(client, nil, imageClient), driver, blobRepoCacheTTL, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ref, err := reference.ParseNamed("nm/is")
	if err != nil {
		t.Errorf("failed to parse blob reference %q: %v", "nm/is", err)
	}

	repo, err := reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// hit the layer repo link - cache the result
	desc, err := repo.Blobs(ctx).Stat(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob1Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob1Desc)
	}

	compareActions(t, "no actions expected", imageClient.Actions(), []clientAction{})

	// remove layer repo link, delete the association from cache as well
	err = repo.Blobs(ctx).Delete(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	// query etcd
	repo, err = reg.Repository(ctx, ref) // the repository needs to be recreated since it caches image streams and images
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob1Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob1Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob1Desc)
	}

	expectedActions := []clientAction{{"get", "imagestreams"}, {"get", "images"}}
	compareActions(t, "1st roundtrip to etcd", imageClient.Actions(), expectedActions)

	// remove the underlying blob
	vacuum := storage.NewVacuum(ctx, driver)
	err = vacuum.RemoveBlob(blob1Dgst.String())
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	// fail because the blob isn't stored locally
	desc, err = repo.Blobs(ctx).Stat(ctx, blob1Dgst)
	if err == nil {
		t.Fatalf("got unexpected non error: %v", err)
	}
	if err != distribution.ErrBlobUnknown {
		t.Fatalf("got unexpected error: %#+v", err)
	}

	// cache hit - don't query etcd
	repo, err = reg.Repository(ctx, ref) // the repository needs to be recreated since it caches image streams and images
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob2Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}

	compareActions(t, "no etcd query", imageClient.Actions(), expectedActions)

	lastStatTimestamp := time.Now()

	// hit the cache
	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob2Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}

	// cache hit - no additional etcd query
	compareActions(t, "no roundrip to etcd", imageClient.Actions(), expectedActions)

	t.Logf("sleeping %s while waiting for eviction of blob %q from cache", blobRepoCacheTTL.String(), blob2Dgst.String())
	time.Sleep(blobRepoCacheTTL - (time.Now().Sub(lastStatTimestamp)))

	repo, err = reg.Repository(ctx, ref)
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err != nil {
		t.Fatalf("got unexpected stat error: %v", err)
	}
	if !reflect.DeepEqual(desc, blob2Desc) {
		t.Fatalf("got unexpected descriptor: %#+v != %#+v", desc, blob2Desc)
	}

	expectedActions = append(expectedActions, []clientAction{{"get", "imagestreams"}, {"get", "images"}}...)
	compareActions(t, "2nd roundtrip to etcd", imageClient.Actions(), expectedActions)

	err = vacuum.RemoveBlob(blob2Dgst.String())
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	// fail because the blob isn't stored locally
	desc, err = repo.Blobs(ctx).Stat(ctx, blob2Dgst)
	if err == nil {
		t.Fatalf("got unexpected non error: %v", err)
	}
	if err != distribution.ErrBlobUnknown {
		t.Fatalf("got unexpected error: %#+v", err)
	}
}

type clientAction struct {
	verb     string
	resource string
}

func storeTestImage(
	ctx context.Context,
	reg distribution.Namespace,
	imageReference reference.NamedTagged,
	schemaVersion int,
	managedByOpenShift bool,
) (*imageapiv1.Image, error) {
	repo, err := reg.Repository(ctx, imageReference)
	if err != nil {
		return nil, fmt.Errorf("unexpected error getting repo %q: %v", imageReference.Name(), err)
	}

	var (
		m  distribution.Manifest
		m1 schema1.Manifest
	)
	switch schemaVersion {
	case 1:
		m1 = schema1.Manifest{
			Versioned: manifest.Versioned{
				SchemaVersion: 1,
			},
			Name: imageReference.Name(),
			Tag:  imageReference.Tag(),
		}
	case 2:
		// TODO
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported manifest version %d", schemaVersion)
	}

	for i := 0; i < testImageLayerCount; i++ {
		payload, err := registrytest.CreateRandomTarFile()
		if err != nil {
			return nil, fmt.Errorf("unexpected error generating test layer file: %v", err)
		}

		wr, err := repo.Blobs(ctx).Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating test upload: %v", err)
		}
		defer wr.Close()

		n, err := io.Copy(wr, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("unexpected error copying to upload: %v", err)
		}

		dgst := digest.FromBytes(payload)

		if schemaVersion == 1 {
			m1.FSLayers = append(m1.FSLayers, schema1.FSLayer{BlobSum: dgst})
			m1.History = append(m1.History, schema1.History{V1Compatibility: fmt.Sprintf(`{"size":%d}`, n)})
		} // TODO v2

		if _, err := wr.Commit(ctx, distribution.Descriptor{Digest: dgst, MediaType: schema1.MediaTypeManifestLayer}); err != nil {
			return nil, fmt.Errorf("unexpected error finishing upload: %v", err)
		}
	}

	var dgst digest.Digest
	var payload []byte

	if schemaVersion == 1 {
		pk, err := libtrust.GenerateECP256PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("unexpected error generating private key: %v", err)
		}

		m, err = schema1.Sign(&m1, pk)
		if err != nil {
			return nil, fmt.Errorf("error signing manifest: %v", err)
		}

		_, payload, err = m.Payload()
		if err != nil {
			return nil, fmt.Errorf("error getting payload %#v", err)
		}

		dgst = digest.FromBytes(payload)
	} //TODO v2

	image := &imageapi.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: dgst.String(),
		},
		DockerImageManifest:  string(payload),
		DockerImageReference: imageReference.Name() + "@" + dgst.String(),
	}

	if managedByOpenShift {
		image.Annotations = map[string]string{imageapi.ManagedByOpenShiftAnnotation: "true"}
	}

	if schemaVersion == 1 {
		signedManifest := m.(*schema1.SignedManifest)
		signatures, err := signedManifest.Signatures()
		if err != nil {
			return nil, err
		}

		for _, signDigest := range signatures {
			image.DockerImageSignatures = append(image.DockerImageSignatures, signDigest)
		}
	}

	if err := imageapi.ImageWithMetadata(image); err != nil {
		return nil, err
	}
	newImage := imageapiv1.Image{}
	if err := kapi.Scheme.Converter().Convert(image, &newImage, 0, nil); err != nil {
		return nil, err
	}

	if err := imageapiv1.ImageWithMetadata(&newImage); err != nil {
		return nil, fmt.Errorf("failed to fill image with metadata: %v", err)
	}

	return &newImage, nil
}

func populateTestStorage(
	t *testing.T,
	driver driver.StorageDriver,
	setManagedByOpenShift bool,
	schemaVersion int,
	repoImages map[string]int,
	testImages map[string][]*imageapiv1.Image,
) (map[string][]*imageapiv1.Image, error) {
	ctx := context.Background()
	reg, err := storage.NewRegistry(ctx, driver)
	if err != nil {
		t.Fatalf("error creating registry: %v", err)
	}

	result := make(map[string][]*imageapiv1.Image)
	for key, value := range testImages {
		images := make([]*imageapiv1.Image, len(value))
		copy(images, value)
		result[key] = images
	}

	for imageReference := range repoImages {
		parsed, err := reference.Parse(imageReference)
		if err != nil {
			t.Fatalf("failed to parse reference %q: %v", imageReference, err)
		}
		namedTagged, ok := parsed.(reference.NamedTagged)
		if !ok {
			t.Fatalf("expected NamedTagged reference, not %T", parsed)
		}

		imageCount := repoImages[imageReference]

		for i := 0; i < imageCount; i++ {
			img, err := storeTestImage(ctx, reg, namedTagged, schemaVersion, setManagedByOpenShift)
			if err != nil {
				t.Fatal(err)
			}
			arr := result[imageReference]
			t.Logf("created image %s@%s image with layers:", namedTagged.Name(), img.Name)
			for _, l := range img.DockerImageLayers {
				t.Logf("  %s of size %d", l.Name, l.LayerSize)
			}
			result[imageReference] = append(arr, img)
		}
	}

	return result, nil
}

func newTestRegistry(
	ctx context.Context,
	osClient registryclient.Interface,
	storageDriver driver.StorageDriver,
	blobrepositorycachettl time.Duration,
	pullthrough bool,
	useBlobDescriptorCacheProvider bool,
) (*testRegistry, error) {
	if storageDriver == nil {
		storageDriver = inmemory.New()
	}
	dockerStorageDriver = storageDriver

	opts := []storage.RegistryOption{
		storage.BlobDescriptorServiceFactory(&blobDescriptorServiceFactory{}),
		storage.EnableDelete,
		storage.EnableRedirect,
	}
	if useBlobDescriptorCacheProvider {
		cacheProvider := cache.BlobDescriptorCacheProvider(memory.NewInMemoryBlobDescriptorCacheProvider())
		opts = append(opts, storage.BlobDescriptorCacheProvider(cacheProvider))
	}

	reg, err := storage.NewRegistry(ctx, dockerStorageDriver, opts...)
	if err != nil {
		return nil, err
	}
	dockerRegistry = reg

	return &testRegistry{
		Namespace:              dockerRegistry,
		osClient:               osClient,
		blobrepositorycachettl: blobrepositorycachettl,
		pullthrough:            pullthrough,
	}, nil
}

type testRegistry struct {
	distribution.Namespace
	osClient               registryclient.Interface
	pullthrough            bool
	blobrepositorycachettl time.Duration
}

var _ distribution.Namespace = &testRegistry{}

func (reg *testRegistry) Repository(ctx context.Context, ref reference.Named) (distribution.Repository, error) {
	repo, err := reg.Namespace.Repository(ctx, ref)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(ref.Name(), "/", 3)
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to parse repository name %q", ref.Name())
	}

	nm, name := parts[0], parts[1]

	isGetter := &cachedImageStreamGetter{
		ctx:          ctx,
		namespace:    nm,
		name:         name,
		isNamespacer: reg.osClient,
	}

	r := &repository{
		Repository: repo,

		ctx:              ctx,
		registryOSClient: reg.osClient,
		registryAddr:     "localhost:5000",
		namespace:        nm,
		name:             name,
		blobrepositorycachettl: reg.blobrepositorycachettl,
		imageStreamGetter:      isGetter,
		cachedImages:           make(map[digest.Digest]*imageapiv1.Image),
		cachedLayers:           cachedLayers,
		pullthrough:            reg.pullthrough,
	}

	if reg.pullthrough {
		r.remoteBlobGetter = NewBlobGetterService(
			nm,
			name,
			defaultBlobRepositoryCacheTTL,
			isGetter.get,
			reg.osClient,
			cachedLayers)
	}

	return r, nil
}

func testNewDescriptorForLayer(layer imageapiv1.ImageLayer) distribution.Descriptor {
	return distribution.Descriptor{
		Digest:    digest.Digest(layer.Name),
		MediaType: "application/octet-stream",
		Size:      layer.LayerSize,
	}
}

func compareActions(t *testing.T, testCaseName string, actions []clientgotesting.Action, expectedActions []clientAction) {
	for i, action := range actions {
		if i >= len(expectedActions) {
			t.Errorf("[%s] got unexpected client action: %#+v", testCaseName, action)
			continue
		}
		expected := expectedActions[i]
		if !action.Matches(expected.verb, expected.resource) {
			t.Errorf("[%s] expected client action %s[%s], got instead: %#+v", testCaseName, expected.verb, expected.resource, action)
		}
	}
	for i := len(actions); i < len(expectedActions); i++ {
		expected := expectedActions[i]
		t.Errorf("[%s] expected action %s[%s] did not happen (%#v)", testCaseName, expected.verb, expected.resource, actions)
	}
}
