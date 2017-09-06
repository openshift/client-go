package login

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kclientcmd "k8s.io/client-go/tools/clientcmd"
	kclientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/term"

	"github.com/openshift/origin/pkg/cmd/flagtypes"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/cli/config"
)

var (
	loginLong = templates.LongDesc(`
		Log in to your server and save login for subsequent use

		First-time users of the client should run this command to connect to a server,
		establish an authenticated session, and save connection to the configuration file. The
		default configuration will be saved to your home directory under
		".kube/config".

		The information required to login -- like username and password, a session token, or
		the server details -- can be provided through flags. If not provided, the command will
		prompt for user input as needed.`)

	loginExample = templates.Examples(`
		# Log in interactively
	  %[1]s login

	  # Log in to the given server with the given certificate authority file
	  %[1]s login localhost:8443 --certificate-authority=/path/to/cert.crt

	  # Log in to the given server with the given credentials (will not prompt interactively)
	  %[1]s login localhost:8443 --username=myuser --password=mypass`)
)

// NewCmdLogin implements the OpenShift cli login command
func NewCmdLogin(fullName string, f *osclientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &LoginOptions{
		Reader: reader,
		Out:    out,
	}

	cmds := &cobra.Command{
		Use:     "login [URL]",
		Short:   "Log in to a server",
		Long:    loginLong,
		Example: fmt.Sprintf(loginExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args, fullName); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Validate(args, kcmdutil.GetFlagString(cmd, "server")); err != nil {
				kcmdutil.CheckErr(err)
			}

			err := RunLogin(cmd, options)

			if kapierrors.IsUnauthorized(err) {
				fmt.Fprintln(out, "Login failed (401 Unauthorized)")

				if err, isStatusErr := err.(*kapierrors.StatusError); isStatusErr {
					if details := err.Status().Details; details != nil {
						for _, cause := range details.Causes {
							fmt.Fprintln(out, cause.Message)
						}
					}
				}

				os.Exit(1)

			} else {
				kcmdutil.CheckErr(err)
			}
		},
	}

	// Login is the only command that can negotiate a session token against the auth server using basic auth
	cmds.Flags().StringVarP(&options.Username, "username", "u", "", "Username, will prompt if not provided")
	cmds.Flags().StringVarP(&options.Password, "password", "p", "", "Password, will prompt if not provided")

	return cmds
}

func (o *LoginOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string, commandName string) error {
	kubeconfig, err := f.OpenShiftClientConfig().RawConfig()
	o.StartingKubeConfig = &kubeconfig
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// build a valid object to use if we failed on a non-existent file
		o.StartingKubeConfig = kclientcmdapi.NewConfig()
	}

	unparsedTimeout := kcmdutil.GetFlagString(cmd, "request-timeout")
	timeout, err := kclientcmd.ParseTimeout(unparsedTimeout)
	if err != nil {
		return err
	}
	o.RequestTimeout = timeout

	o.CommandName = commandName
	if o.CommandName == "" {
		o.CommandName = "oc"
	}

	parsedDefaultClusterURL, err := url.Parse(defaultClusterURL)
	if err != nil {
		return err
	}
	addr := flagtypes.Addr{Value: parsedDefaultClusterURL.Host, DefaultScheme: parsedDefaultClusterURL.Scheme, AllowPrefix: true}.Default()

	if serverFlag := kcmdutil.GetFlagString(cmd, "server"); len(serverFlag) > 0 {
		if err := addr.Set(serverFlag); err != nil {
			return err
		}
		o.Server = addr.String()

	} else if len(args) == 1 {
		if err := addr.Set(args[0]); err != nil {
			return err
		}
		o.Server = addr.String()

	} else if len(o.Server) == 0 {
		if defaultContext, defaultContextExists := o.StartingKubeConfig.Contexts[o.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if cluster, exists := o.StartingKubeConfig.Clusters[defaultContext.Cluster]; exists {
				o.Server = cluster.Server
			}
		}
	}

	o.CertFile = kcmdutil.GetFlagString(cmd, "client-certificate")
	o.KeyFile = kcmdutil.GetFlagString(cmd, "client-key")

	o.CAFile = kcmdutil.GetFlagString(cmd, "certificate-authority")
	o.InsecureTLS = kcmdutil.GetFlagBool(cmd, "insecure-skip-tls-verify")
	o.Token = kcmdutil.GetFlagString(cmd, "token")

	o.DefaultNamespace, _, _ = f.DefaultNamespace()

	o.PathOptions = config.NewPathOptions(cmd)

	return nil
}

func (o LoginOptions) Validate(args []string, serverFlag string) error {
	if len(args) > 1 {
		return errors.New("Only the server URL may be specified as an argument")
	}

	if (len(serverFlag) > 0) && (len(args) == 1) {
		return errors.New("--server and passing the server URL as an argument are mutually exclusive")
	}

	if (len(o.Server) == 0) && !term.IsTerminal(o.Reader) {
		return errors.New("A server URL must be specified")
	}

	if len(o.Username) > 0 && len(o.Token) > 0 {
		return errors.New("--token and --username are mutually exclusive")
	}

	if o.StartingKubeConfig == nil {
		return errors.New("Must have a config file already created")
	}

	return nil
}

// RunLogin contains all the necessary functionality for the OpenShift cli login command
func RunLogin(cmd *cobra.Command, options *LoginOptions) error {
	if err := options.GatherInfo(); err != nil {
		return err
	}

	newFileCreated, err := options.SaveConfig()
	if err != nil {
		return err
	}

	if newFileCreated {
		fmt.Fprintf(options.Out, "Welcome! See '%s help' to get started.\n", options.CommandName)
	}
	return nil
}
