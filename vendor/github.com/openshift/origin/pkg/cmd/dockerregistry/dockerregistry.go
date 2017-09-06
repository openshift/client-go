package dockerregistry

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	logrus_logstash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/docker/go-units"
	gorillahandlers "github.com/gorilla/handlers"

	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/health"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/handlers"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/factory"
	"github.com/docker/distribution/uuid"
	distversion "github.com/docker/distribution/version"

	_ "github.com/docker/distribution/registry/auth/htpasswd"
	_ "github.com/docker/distribution/registry/auth/token"

	_ "github.com/docker/distribution/registry/proxy"
	_ "github.com/docker/distribution/registry/storage/driver/azure"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	_ "github.com/docker/distribution/registry/storage/driver/gcs"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/docker/distribution/registry/storage/driver/oss"
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	_ "github.com/docker/distribution/registry/storage/driver/swift"

	kubeversion "k8s.io/kubernetes/pkg/version"

	"github.com/openshift/origin/pkg/cmd/server/crypto"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/dockerregistry/server"
	"github.com/openshift/origin/pkg/dockerregistry/server/api"
	"github.com/openshift/origin/pkg/dockerregistry/server/audit"
	"github.com/openshift/origin/pkg/dockerregistry/server/client"
	registryconfig "github.com/openshift/origin/pkg/dockerregistry/server/configuration"
	"github.com/openshift/origin/pkg/dockerregistry/server/maxconnections"
	"github.com/openshift/origin/pkg/dockerregistry/server/prune"
	"github.com/openshift/origin/pkg/version"
)

var pruneMode = flag.String("prune", "", "prune blobs from the storage and exit (check, delete)")

func versionFields() log.Fields {
	return log.Fields{
		"distribution_version": distversion.Version,
		"kubernetes_version":   kubeversion.Get(),
		"openshift_version":    version.Get(),
	}
}

// ExecutePruner runs the pruner.
func ExecutePruner(configFile io.Reader, dryRun bool) {
	config, _, err := registryconfig.Parse(configFile)
	if err != nil {
		log.Fatalf("error parsing configuration file: %s", err)
	}

	// A lot of installations have the 'debug' log level in their config files,
	// but it's too verbose for pruning. Therefore we ignore it, but we still
	// respect overrides using environment variables.
	config.Loglevel = ""
	config.Log.Level = configuration.Loglevel(os.Getenv("REGISTRY_LOG_LEVEL"))
	if len(config.Log.Level) == 0 {
		config.Log.Level = "warning"
	}

	ctx := context.Background()
	ctx, err = configureLogging(ctx, config)
	if err != nil {
		log.Fatalf("error configuring logging: %s", err)
	}

	startPrune := "start prune"
	var registryOptions []storage.RegistryOption
	if dryRun {
		startPrune += " (dry-run mode)"
	} else {
		registryOptions = append(registryOptions, storage.EnableDelete)
	}
	log.WithFields(versionFields()).Info(startPrune)

	registryClient := client.NewRegistryClient(clientcmd.NewConfig().BindToFile())

	storageDriver, err := factory.Create(config.Storage.Type(), config.Storage.Parameters())
	if err != nil {
		log.Fatalf("error creating storage driver: %s", err)
	}

	registry, err := storage.NewRegistry(ctx, storageDriver, registryOptions...)
	if err != nil {
		log.Fatalf("error creating registry: %s", err)
	}

	stats, err := prune.Prune(ctx, storageDriver, registry, registryClient, dryRun)
	if err != nil {
		log.Error(err)
	}
	if dryRun {
		fmt.Printf("Would delete %d blobs\n", stats.Blobs)
		fmt.Printf("Would free up %s of disk space\n", units.BytesSize(float64(stats.DiskSpace)))
		fmt.Println("Use -prune=delete to actually delete the data")
	} else {
		fmt.Printf("Deleted %d blobs\n", stats.Blobs)
		fmt.Printf("Freed up %s of disk space\n", units.BytesSize(float64(stats.DiskSpace)))
	}
	if err != nil {
		os.Exit(1)
	}
}

// Execute runs the Docker registry.
func Execute(configFile io.Reader) {
	if len(*pruneMode) != 0 {
		var dryRun bool
		switch *pruneMode {
		case "delete":
			dryRun = false
		case "check":
			dryRun = true
		default:
			log.Fatal("invalid value for the -prune option")
		}
		ExecutePruner(configFile, dryRun)
		return
	}

	dockerConfig, extraConfig, err := registryconfig.Parse(configFile)
	if err != nil {
		log.Fatalf("error parsing configuration file: %s", err)
	}
	setDefaultMiddleware(dockerConfig)
	setDefaultLogParameters(dockerConfig)

	ctx := context.Background()
	ctx = server.WithConfiguration(ctx, extraConfig)
	ctx, err = configureLogging(ctx, dockerConfig)
	if err != nil {
		log.Fatalf("error configuring logger: %v", err)
	}

	registryClient := client.NewRegistryClient(clientcmd.NewConfig().BindToFile())
	ctx = server.WithRegistryClient(ctx, registryClient)

	readLimiter := newLimiter(extraConfig.Requests.Read)
	writeLimiter := newLimiter(extraConfig.Requests.Write)
	ctx = server.WithWriteLimiter(ctx, writeLimiter)

	log.WithFields(versionFields()).Info("start registry")
	// inject a logger into the uuid library. warns us if there is a problem
	// with uuid generation under low entropy.
	uuid.Loggerf = context.GetLogger(ctx).Warnf

	// add parameters for the auth middleware
	if dockerConfig.Auth.Type() == server.OpenShiftAuth {
		if dockerConfig.Auth[server.OpenShiftAuth] == nil {
			dockerConfig.Auth[server.OpenShiftAuth] = make(configuration.Parameters)
		}
		dockerConfig.Auth[server.OpenShiftAuth][server.AccessControllerOptionParams] = server.AccessControllerParams{
			Logger:         context.GetLogger(ctx),
			RegistryClient: registryClient,
		}
	}

	app := handlers.NewApp(ctx, dockerConfig)

	// Add a token handling endpoint
	if options, usingOpenShiftAuth := dockerConfig.Auth[server.OpenShiftAuth]; usingOpenShiftAuth {
		tokenRealm, err := server.TokenRealm(options)
		if err != nil {
			context.GetLogger(app).Fatalf("error setting up token auth: %s", err)
		}
		err = app.NewRoute().Methods("GET").PathPrefix(tokenRealm.Path).Handler(server.NewTokenHandler(ctx, registryClient)).GetError()
		if err != nil {
			context.GetLogger(app).Fatalf("error setting up token endpoint at %q: %v", tokenRealm.Path, err)
		}
		context.GetLogger(app).Debugf("configured token endpoint at %q", tokenRealm.String())
	}

	// TODO add https scheme
	adminRouter := app.NewRoute().PathPrefix(api.AdminPrefix).Subrouter()
	pruneAccessRecords := func(*http.Request) []auth.Access {
		return []auth.Access{
			{
				Resource: auth.Resource{
					Type: "admin",
				},
				Action: "prune",
			},
		}
	}

	app.RegisterRoute(
		// DELETE /admin/blobs/<digest>
		adminRouter.Path(api.AdminPath).Methods("DELETE"),
		// handler
		server.BlobDispatcher,
		// repo name not required in url
		handlers.NameNotRequired,
		// custom access records
		pruneAccessRecords,
	)

	// Registry extensions endpoint provides extra functionality to handle the image
	// signatures.
	server.RegisterSignatureHandler(app)

	// Registry extensions endpoint provides prometheus metrics.
	if extraConfig.Metrics.Enabled {
		if len(extraConfig.Metrics.Secret) == 0 {
			context.GetLogger(app).Fatalf("openshift.metrics.secret field cannot be empty when metrics are enabled")
		}
		server.RegisterMetricHandler(app)
	}

	// Advertise features supported by OpenShift
	if app.Config.HTTP.Headers == nil {
		app.Config.HTTP.Headers = http.Header{}
	}
	app.Config.HTTP.Headers.Set("X-Registry-Supports-Signatures", "1")

	app.RegisterHealthChecks()
	handler := http.Handler(app)
	handler = limit(readLimiter, writeLimiter, handler)
	handler = alive("/", handler)
	// TODO: temporarily keep for backwards compatibility; remove in the future
	handler = alive("/healthz", handler)
	handler = health.Handler(handler)
	handler = panicHandler(handler)
	handler = gorillahandlers.CombinedLoggingHandler(os.Stdout, handler)

	if dockerConfig.HTTP.TLS.Certificate == "" {
		context.GetLogger(app).Infof("listening on %v", dockerConfig.HTTP.Addr)
		if err := http.ListenAndServe(dockerConfig.HTTP.Addr, handler); err != nil {
			context.GetLogger(app).Fatalln(err)
		}
	} else {
		var (
			minVersion   uint16
			cipherSuites []uint16
		)
		if s := os.Getenv("REGISTRY_HTTP_TLS_MINVERSION"); len(s) > 0 {
			minVersion, err = crypto.TLSVersion(s)
			if err != nil {
				context.GetLogger(app).Fatalln(fmt.Errorf("invalid TLS version %q specified in REGISTRY_HTTP_TLS_MINVERSION: %v (valid values are %q)", s, err, crypto.ValidTLSVersions()))
			}
		}
		if s := os.Getenv("REGISTRY_HTTP_TLS_CIPHERSUITES"); len(s) > 0 {
			for _, cipher := range strings.Split(s, ",") {
				cipherSuite, err := crypto.CipherSuite(cipher)
				if err != nil {
					context.GetLogger(app).Fatalln(fmt.Errorf("invalid cipher suite %q specified in REGISTRY_HTTP_TLS_CIPHERSUITES: %v (valid suites are %q)", s, err, crypto.ValidCipherSuites()))
				}
				cipherSuites = append(cipherSuites, cipherSuite)
			}
		}

		tlsConf := crypto.SecureTLSConfig(&tls.Config{
			ClientAuth:   tls.NoClientCert,
			MinVersion:   minVersion,
			CipherSuites: cipherSuites,
		})

		if len(dockerConfig.HTTP.TLS.ClientCAs) != 0 {
			pool := x509.NewCertPool()

			for _, ca := range dockerConfig.HTTP.TLS.ClientCAs {
				caPem, err := ioutil.ReadFile(ca)
				if err != nil {
					context.GetLogger(app).Fatalln(err)
				}

				if ok := pool.AppendCertsFromPEM(caPem); !ok {
					context.GetLogger(app).Fatalln(fmt.Errorf("Could not add CA to pool"))
				}
			}

			for _, subj := range pool.Subjects() {
				context.GetLogger(app).Debugf("CA Subject: %s", string(subj))
			}

			tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConf.ClientCAs = pool
		}

		context.GetLogger(app).Infof("listening on %v, tls", dockerConfig.HTTP.Addr)
		server := &http.Server{
			Addr:      dockerConfig.HTTP.Addr,
			Handler:   handler,
			TLSConfig: tlsConf,
		}

		if err := server.ListenAndServeTLS(dockerConfig.HTTP.TLS.Certificate, dockerConfig.HTTP.TLS.Key); err != nil {
			context.GetLogger(app).Fatalln(err)
		}
	}
}

// configureLogging prepares the context with a logger using the
// configuration.
func configureLogging(ctx context.Context, config *configuration.Configuration) (context.Context, error) {
	if config.Log.Level == "" && config.Log.Formatter == "" {
		// If no config for logging is set, fallback to deprecated "Loglevel".
		log.SetLevel(logLevel(config.Loglevel))
		ctx = context.WithLogger(ctx, context.GetLogger(ctx))
		return ctx, nil
	}

	log.SetLevel(logLevel(config.Log.Level))

	formatter := config.Log.Formatter
	if formatter == "" {
		formatter = "text" // default formatter
	}

	switch formatter {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "logstash":
		log.SetFormatter(&logrus_logstash.LogstashFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		// just let the library use default on empty string.
		if config.Log.Formatter != "" {
			return ctx, fmt.Errorf("unsupported logging formatter: %q", config.Log.Formatter)
		}
	}

	if config.Log.Formatter != "" {
		log.Debugf("using %q logging formatter", config.Log.Formatter)
	}

	if len(config.Log.Fields) > 0 {
		// build up the static fields, if present.
		var fields []interface{}
		for k := range config.Log.Fields {
			fields = append(fields, k)
		}

		ctx = context.WithValues(ctx, config.Log.Fields)
		ctx = context.WithLogger(ctx, context.GetLogger(ctx, fields...))
	}

	return ctx, nil
}

func logLevel(level configuration.Loglevel) log.Level {
	l, err := log.ParseLevel(string(level))
	if err != nil {
		l = log.InfoLevel
		log.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	return l
}

func newLimiter(c registryconfig.RequestsLimits) maxconnections.Limiter {
	if c.MaxRunning <= 0 {
		return nil
	}
	return maxconnections.NewLimiter(c.MaxRunning, c.MaxInQueue, c.MaxWaitInQueue)
}

func limit(readLimiter, writeLimiter maxconnections.Limiter, handler http.Handler) http.Handler {
	readHandler := handler
	if readLimiter != nil {
		readHandler = maxconnections.New(readLimiter, readHandler)
	}

	writeHandler := handler
	if writeLimiter != nil {
		writeHandler = maxconnections.New(writeLimiter, writeHandler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToUpper(r.Method) {
		case "GET", "HEAD", "OPTIONS":
			readHandler.ServeHTTP(w, r)
		default:
			writeHandler.ServeHTTP(w, r)
		}
	})
}

// alive simply wraps the handler with a route that always returns an http 200
// response when the path is matched. If the path is not matched, the request
// is passed to the provided handler. There is no guarantee of anything but
// that the server is up. Wrap with other handlers (such as health.Handler)
// for greater affect.
func alive(path string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// panicHandler add a HTTP handler to web app. The handler recover the happening
// panic. logrus.Panic transmits panic message to pre-config log hooks, which is
// defined in config.yml.
func panicHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Panic(fmt.Sprintf("%v", err))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func setDefaultMiddleware(config *configuration.Configuration) {
	// Default to openshift middleware for relevant types
	// This allows custom configs based on old default configs to continue to work
	if config.Middleware == nil {
		config.Middleware = map[string][]configuration.Middleware{}
	}
	for _, middlewareType := range []string{"registry", "repository", "storage"} {
		found := false
		for _, middleware := range config.Middleware[middlewareType] {
			if middleware.Name == "openshift" {
				found = true
				break
			}
		}
		if found {
			continue
		}
		config.Middleware[middlewareType] = append(config.Middleware[middlewareType], configuration.Middleware{
			Name: "openshift",
		})
		log.Errorf("obsolete configuration detected, please add openshift %s middleware into registry config file", middlewareType)
	}
	return
}

func setDefaultLogParameters(config *configuration.Configuration) {
	if len(config.Log.Fields) == 0 {
		config.Log.Fields = make(map[string]interface{})
	}
	config.Log.Fields[audit.LogEntryType] = audit.DefaultLoggerType
}
