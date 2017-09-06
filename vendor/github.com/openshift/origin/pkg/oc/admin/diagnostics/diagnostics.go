package diagnostics

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	kutilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/cmd/flagtypes"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/cmd/util/variable"
	"github.com/openshift/origin/pkg/diagnostics/log"
	netutil "github.com/openshift/origin/pkg/diagnostics/networkpod/util"
	"github.com/openshift/origin/pkg/diagnostics/types"
	"github.com/openshift/origin/pkg/oc/admin/diagnostics/options"
	"github.com/openshift/origin/pkg/oc/cli/config"
)

// DiagnosticsOptions holds values received from command line flags as well as
// other objects generated for the command to operate.
type DiagnosticsOptions struct {
	// list of diagnostic names to limit what is run
	RequestedDiagnostics []string
	// specify locations of host config files
	MasterConfigLocation string
	NodeConfigLocation   string
	// specify context name to be used for cluster-admin access
	ClientClusterContext string
	// indicate this is an openshift host despite lack of other indicators
	IsHost bool
	// specify the image template to use for DiagnosticPod
	ImageTemplate variable.ImageTemplate
	// When true, prevent diagnostics from changing API state (e.g. creating something)
	PreventModification bool
	// We need a factory for creating clients. Creating a factory
	// creates flags as a byproduct, most of which we don't want.
	// The command creates these and binds only the flags we want.
	ClientFlags *flag.FlagSet
	Factory     *osclientcmd.Factory
	// LogOptions determine globally what the user wants to see and how.
	LogOptions *log.LoggerOptions
	// The Logger is built with the options and should be used for all diagnostic output.
	Logger *log.Logger
	// Options specific to network diagnostics
	NetworkOptions *NetworkDiagnosticsOptions
}

// NetworkDiagnosticsOptions holds additional values received from command line flags that
// are specify to network diagnostics.
type NetworkDiagnosticsOptions struct {
	// Path to store network diagnostic results in case of errors
	LogDir string
	// Image to use for network diagnostic pod
	PodImage string
	// Image to use for network diagnostic test pod
	TestPodImage string
	// Protocol used to connect network diagnostic test pod
	TestPodProtocol string
	// Serving port on the network diagnostic test pod
	TestPodPort int
}

const (
	DiagnosticsRecommendedName = "diagnostics"

	// Standard locations for the host config files OpenShift uses.
	StandardMasterConfigPath string = "/etc/origin/master/master-config.yaml"
	StandardNodeConfigPath   string = "/etc/origin/node/node-config.yaml"
)

var (
	longDescription = templates.LongDesc(`
		This utility helps troubleshoot and diagnose known problems. It runs
		diagnostics using a client and/or the state of a running master /
		node host.

		    %[1]s

		If run without flags, it will check for standard config files for
		client, master, and node, and if found, use them for diagnostics.
		You may also specify config files explicitly with flags, in which case
		you will receive an error if they are not found. For example:

		    %[1]s --master-config=/etc/origin/master/master-config.yaml

		* If master/node config files are not found and the --host flag is not
		  present, host diagnostics are skipped.
		* If the client has cluster-admin access, this access enables cluster
		  diagnostics to run which regular users cannot.
		* If a client config file is not found, client and cluster diagnostics
		  are skipped.

		Diagnostics may be individually run by passing diagnostic name as arguments.

		    %[1]s <DiagnosticName>

		The available diagnostic names are: %[2]s.`)
)

// NewCmdDiagnostics is the base command for running any diagnostics.
func NewCmdDiagnostics(name string, fullName string, out io.Writer) *cobra.Command {
	o := &DiagnosticsOptions{
		RequestedDiagnostics: []string{},
		LogOptions:           &log.LoggerOptions{Out: out},
		ImageTemplate:        variable.NewDefaultImageTemplate(),
		NetworkOptions:       &NetworkDiagnosticsOptions{},
	}

	cmd := &cobra.Command{
		Use:   name,
		Short: "Diagnose common cluster problems",
		Long:  fmt.Sprintf(longDescription, fullName, strings.Join(availableDiagnostics().List(), ", ")),
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(args))

			kcmdutil.CheckErr(o.Validate())

			failed, err, warnCount, errorCount := o.RunDiagnostics()
			o.Logger.Summary(warnCount, errorCount)

			kcmdutil.CheckErr(err)
			if failed {
				os.Exit(255)
			}

		},
	}
	cmd.SetOutput(out) // for output re: usage / help

	o.ClientFlags = flag.NewFlagSet("client", flag.ContinueOnError) // hide the extensive set of client flags
	o.Factory = osclientcmd.New(o.ClientFlags)                      // that would otherwise be added to this command
	cmd.Flags().AddFlag(o.ClientFlags.Lookup(config.OpenShiftConfigFlagName))
	cmd.Flags().AddFlag(o.ClientFlags.Lookup("context")) // TODO: find k8s constant
	cmd.Flags().StringVar(&o.ClientClusterContext, options.FlagClusterContextName, "", "Client context to use for cluster administrator")
	cmd.Flags().StringVar(&o.MasterConfigLocation, options.FlagMasterConfigName, "", "Path to master config file (implies --host)")
	cmd.Flags().StringVar(&o.NodeConfigLocation, options.FlagNodeConfigName, "", "Path to node config file (implies --host)")
	cmd.Flags().BoolVar(&o.IsHost, options.FlagIsHostName, false, "If true, look for systemd and journald units even without master/node config")
	cmd.Flags().StringVar(&o.ImageTemplate.Format, options.FlagImageTemplateName, o.ImageTemplate.Format, "Image template for DiagnosticPod to use in creating a pod")
	cmd.Flags().BoolVar(&o.ImageTemplate.Latest, options.FlagLatestImageName, false, "If true, when expanding the image template, use latest version, not release version")
	cmd.Flags().BoolVar(&o.PreventModification, options.FlagPreventModificationName, false, "If true, may be set to prevent diagnostics making any changes via the API")
	cmd.Flags().StringVar(&o.NetworkOptions.LogDir, options.FlagNetworkDiagLogDir, netutil.NetworkDiagDefaultLogDir, "Path to store network diagnostic results in case of errors")
	cmd.Flags().StringVar(&o.NetworkOptions.PodImage, options.FlagNetworkDiagPodImage, netutil.NetworkDiagDefaultPodImage, "Image to use for network diagnostic pod")
	cmd.Flags().StringVar(&o.NetworkOptions.TestPodImage, options.FlagNetworkDiagTestPodImage, netutil.NetworkDiagDefaultTestPodImage, "Image to use for network diagnostic test pod")
	cmd.Flags().StringVar(&o.NetworkOptions.TestPodProtocol, options.FlagNetworkDiagTestPodProtocol, netutil.NetworkDiagDefaultTestPodProtocol, "Protocol used to connect to network diagnostic test pod")
	cmd.Flags().IntVar(&o.NetworkOptions.TestPodPort, options.FlagNetworkDiagTestPodPort, netutil.NetworkDiagDefaultTestPodPort, "Serving port on the network diagnostic test pod")
	flagtypes.GLog(cmd.Flags())
	options.BindLoggerOptionFlags(cmd.Flags(), o.LogOptions, options.RecommendedLoggerOptionFlags())

	return cmd
}

// Complete fills in DiagnosticsOptions needed if the command is actually invoked.
func (o *DiagnosticsOptions) Complete(args []string) error {
	var err error
	o.Logger, err = o.LogOptions.NewLogger()
	if err != nil {
		return err
	}

	// If not given master/client config file locations, check if the defaults exist
	// and adjust the options accordingly:
	if len(o.MasterConfigLocation) == 0 {
		if _, err := os.Stat(StandardMasterConfigPath); !os.IsNotExist(err) {
			o.MasterConfigLocation = StandardMasterConfigPath
		}
	}
	if len(o.NodeConfigLocation) == 0 {
		if _, err := os.Stat(StandardNodeConfigPath); !os.IsNotExist(err) {
			o.NodeConfigLocation = StandardNodeConfigPath
		}
	}

	if len(o.NetworkOptions.LogDir) == 0 {
		o.NetworkOptions.LogDir = netutil.NetworkDiagDefaultLogDir
	} else {
		logdir, err := filepath.Abs(o.NetworkOptions.LogDir)
		if err != nil {
			return err
		}
		if path, err := os.Stat(o.NetworkOptions.LogDir); err == nil && !path.Mode().IsDir() {
			return fmt.Errorf("Network log path %q exists but is not a directory", o.NetworkOptions.LogDir)
		}
		o.NetworkOptions.LogDir = logdir
	}
	if len(o.NetworkOptions.PodImage) == 0 {
		o.NetworkOptions.PodImage = netutil.NetworkDiagDefaultPodImage
	}
	if len(o.NetworkOptions.TestPodImage) == 0 {
		o.NetworkOptions.TestPodImage = netutil.NetworkDiagDefaultTestPodImage
	}

	supportedProtocols := sets.NewString(string(kapi.ProtocolTCP), string(kapi.ProtocolUDP))
	if !supportedProtocols.Has(o.NetworkOptions.TestPodProtocol) {
		return fmt.Errorf("invalid protocol for network diagnostic test pod. Supported protocols: %s", strings.Join(supportedProtocols.List(), ","))
	}
	if kvalidation.IsValidPortNum(o.NetworkOptions.TestPodPort) != nil {
		return fmt.Errorf("invalid port for network diagnostic test pod. Must be in the range 1 to 65535.")
	}

	o.RequestedDiagnostics = append(o.RequestedDiagnostics, args...)
	if len(o.RequestedDiagnostics) == 0 {
		o.RequestedDiagnostics = availableDiagnostics().Difference(defaultSkipDiagnostics()).List()
	}

	return nil
}

func (o *DiagnosticsOptions) Validate() error {
	available := availableDiagnostics()

	if common := available.Intersection(sets.NewString(o.RequestedDiagnostics...)); len(common) == 0 {
		o.Logger.Error("CED3012", log.EvalTemplate("CED3012", "None of the requested diagnostics are available:\n  {{.requested}}\nPlease try from the following:\n  {{.available}}",
			log.Hash{"requested": o.RequestedDiagnostics, "available": available.List()}))
		return fmt.Errorf("No requested diagnostics are available: requested=%s available=%s", strings.Join(o.RequestedDiagnostics, " "), strings.Join(available.List(), " "))

	} else if len(common) < len(o.RequestedDiagnostics) {
		o.Logger.Error("CED3013", log.EvalTemplate("CED3013", `
Of the requested diagnostics:
    {{.requested}}
only these are available:
    {{.common}}
The list of all possible is:
    {{.available}}
		`, log.Hash{"requested": o.RequestedDiagnostics, "common": common.List(), "available": available.List()}))

		return fmt.Errorf("Not all requested diagnostics are available: missing=%s requested=%s available=%s",
			strings.Join(sets.NewString(o.RequestedDiagnostics...).Difference(available).List(), " "),
			strings.Join(o.RequestedDiagnostics, " "),
			strings.Join(available.List(), " "))
	}

	return nil
}

func availableDiagnostics() sets.String {
	available := sets.NewString()
	available.Insert(availableClientDiagnostics.List()...)
	available.Insert(availableClusterDiagnostics.List()...)
	available.Insert(availableEtcdDiagnostics.List()...)
	available.Insert(availableHostDiagnostics.List()...)
	return available
}

func defaultSkipDiagnostics() sets.String {
	available := sets.NewString()
	available.Insert(defaultSkipEtcdDiagnostics.List()...)
	return available
}

// RunDiagnostics builds diagnostics based on the options and executes them, returning a summary.
func (o DiagnosticsOptions) RunDiagnostics() (bool, error, int, int) {
	failed := false
	warnings := []error{}
	errors := []error{}
	diagnostics := []types.Diagnostic{}

	func() { // don't trust discovery/build of diagnostics; wrap panic nicely in case of developer error
		defer func() {
			if r := recover(); r != nil {
				failed = true
				stack := debug.Stack()
				errors = append(errors, fmt.Errorf("While building the diagnostics, a panic was encountered.\nThis is a bug in diagnostics. Error and stack trace follow: \n%v\n%s", r, stack))
			}
		}()
		detected, detectWarnings, detectErrors := o.detectClientConfig() // may log and return problems
		for _, warn := range detectWarnings {
			warnings = append(warnings, warn)
		}
		for _, err := range detectErrors {
			errors = append(errors, err)
		}
		if !detected { // there just plain isn't any client config file available
			o.Logger.Notice("CED3014", "No client configuration specified; skipping client and cluster diagnostics.")
		} else if rawConfig, err := o.buildRawConfig(); err != nil { // client config is totally broken - won't parse etc (problems may have been detected and logged)
			o.Logger.Error("CED3015", fmt.Sprintf("Client configuration failed to load; skipping client and cluster diagnostics due to error: %s", err.Error()))
			errors = append(errors, err)
		} else {
			clientDiags, ok, err := o.buildClientDiagnostics(rawConfig)
			failed = failed || !ok
			if ok {
				diagnostics = append(diagnostics, clientDiags...)
			}
			if err != nil {
				errors = append(errors, err)
			}

			clusterDiags, ok, err := o.buildClusterDiagnostics(rawConfig)
			failed = failed || !ok
			if ok {
				diagnostics = append(diagnostics, clusterDiags...)
			}
			if err != nil {
				errors = append(errors, err)
			}
		}

		etcdDiags, ok, err := o.buildEtcdDiagnostics()
		failed = failed || !ok
		if ok {
			diagnostics = append(diagnostics, etcdDiags...)
		}

		hostDiags, ok, err := o.buildHostDiagnostics()
		failed = failed || !ok
		if ok {
			diagnostics = append(diagnostics, hostDiags...)
		}
		if err != nil {
			errors = append(errors, err)
		}
	}()

	if failed {
		return failed, kutilerrors.NewAggregate(errors), len(warnings), len(errors)
	}

	failed, err, numWarnings, numErrors := o.Run(diagnostics)
	numWarnings += len(warnings)
	numErrors += len(errors)
	return failed, err, numWarnings, numErrors
}

// Run performs the actual execution of diagnostics once they're built.
func (o DiagnosticsOptions) Run(diagnostics []types.Diagnostic) (bool, error, int, int) {
	warnCount := 0
	errorCount := 0
	for _, diagnostic := range diagnostics {
		func() { // wrap diagnostic panic nicely in case of developer error
			defer func() {
				if r := recover(); r != nil {
					errorCount += 1
					stack := debug.Stack()
					o.Logger.Error("CED3017",
						fmt.Sprintf("While running the %s diagnostic, a panic was encountered.\nThis is a bug in diagnostics. Error and stack trace follow: \n%s\n%s",
							diagnostic.Name(), fmt.Sprintf("%v", r), stack))
				}
			}()

			if canRun, reason := diagnostic.CanRun(); !canRun {
				if reason == nil {
					o.Logger.Notice("CED3018", fmt.Sprintf("Skipping diagnostic: %s\nDescription: %s", diagnostic.Name(), diagnostic.Description()))
				} else {
					o.Logger.Notice("CED3019", fmt.Sprintf("Skipping diagnostic: %s\nDescription: %s\nBecause: %s", diagnostic.Name(), diagnostic.Description(), reason.Error()))
				}
				return
			}

			o.Logger.Notice("CED3020", fmt.Sprintf("Running diagnostic: %s\nDescription: %s", diagnostic.Name(), diagnostic.Description()))
			r := diagnostic.Check()
			for _, entry := range r.Logs() {
				o.Logger.LogEntry(entry)
			}
			warnCount += len(r.Warnings())
			errorCount += len(r.Errors())
		}()
	}
	return errorCount > 0, nil, warnCount, errorCount
}
