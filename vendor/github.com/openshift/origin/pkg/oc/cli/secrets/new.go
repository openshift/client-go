package secrets

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	kapi "k8s.io/kubernetes/pkg/api"
	kcoreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
)

const NewSecretRecommendedCommandName = "new"

var (
	newLong = templates.LongDesc(`
    Create a new secret based on a file or directory

    Key files can be specified using their file path, in which case a default name will be given to them, or optionally
    with a name and file path, in which case the given name will be used. Specifying a directory will create a secret
    using with all valid keys in that directory.`)

	newExample = templates.Examples(`
    # Create a new secret named my-secret with a key named ssh-privatekey
    %[1]s my-secret ~/.ssh/ssh-privatekey

    # Create a new secret named my-secret with keys named ssh-privatekey and ssh-publickey instead of the names of the keys on disk
    %[1]s my-secret ssh-privatekey=~/.ssh/id_rsa ssh-publickey=~/.ssh/id_rsa.pub

    # Create a new secret named my-secret with keys for each file in the folder "bar"
    %[1]s my-secret path/to/bar

    # Create a new .dockercfg secret named my-secret
    %[1]s my-secret path/to/.dockercfg

    # Create a new .docker/config.json secret named my-secret
    %[1]s my-secret .dockerconfigjson=path/to/.docker/config.json`)
)

type CreateSecretOptions struct {
	// Name of the resulting secret
	Name string

	// SecretTypeName is the type to use when creating the secret.  It is checked against known types.
	SecretTypeName string

	// Files/Directories to read from.
	// Directory sources are listed and any direct file children included (but subfolders are not traversed)
	Sources []string

	SecretsInterface kcoreclient.SecretInterface

	// Writer to write warnings to
	Stderr io.Writer

	Out io.Writer

	// Controls whether to output warnings
	Quiet bool

	AllowUnknownTypes bool
}

func NewCmdCreateSecret(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := NewCreateSecretOptions()
	options.Out = out

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s NAME [KEY=]SOURCE ...", name),
		Short:   "Create a new secret based on a key file or on files within a directory",
		Long:    newLong,
		Example: fmt.Sprintf(newExample, fullName),
		Run: func(c *cobra.Command, args []string) {
			if err := options.Complete(args, f); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(c, err.Error()))
			}

			if err := options.Validate(); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(c, err.Error()))
			}

			if len(kcmdutil.GetFlagString(c, "output")) != 0 {
				secret, err := options.BundleSecret()
				kcmdutil.CheckErr(err)

				mapper, _ := f.Object()
				kcmdutil.CheckErr(f.PrintObject(c, false, mapper, secret, out))
				return
			}

			_, err := options.CreateSecret()
			kcmdutil.CheckErr(err)
		},
	}

	cmd.Flags().BoolVarP(&options.Quiet, "quiet", "q", options.Quiet, "If true, suppress warnings")
	cmd.Flags().BoolVar(&options.AllowUnknownTypes, "confirm", options.AllowUnknownTypes, "If true, allow unknown secret types.")
	cmd.Flags().StringVar(&options.SecretTypeName, "type", "", "The type of secret")
	kcmdutil.AddPrinterFlags(cmd)

	return cmd
}

func NewCreateSecretOptions() *CreateSecretOptions {
	return &CreateSecretOptions{
		Stderr:  os.Stderr,
		Sources: []string{},
	}
}

func (o *CreateSecretOptions) Complete(args []string, f *clientcmd.Factory) error {
	// Fill name from args[0]
	if len(args) > 0 {
		o.Name = args[0]
	}

	// Add sources from args[1:...] in addition to -f
	if len(args) > 1 {
		o.Sources = append(o.Sources, args[1:]...)
	}

	if f != nil {
		_, kubeClient, err := f.Clients()
		if err != nil {
			return err
		}
		namespace, _, err := f.DefaultNamespace()
		if err != nil {
			return err
		}
		o.SecretsInterface = kubeClient.Core().Secrets(namespace)
	}

	return nil
}

func (o *CreateSecretOptions) Validate() error {
	if len(o.Name) == 0 {
		return errors.New("secret name is required")
	}
	if len(o.Sources) == 0 {
		return errors.New("at least one source file or directory must be specified")
	}

	if !o.AllowUnknownTypes {
		switch o.SecretTypeName {
		case string(kapi.SecretTypeOpaque), "":
			// this is ok
		default:
			found := false
			for _, secretType := range KnownSecretTypes {
				if o.SecretTypeName == string(secretType.Type) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("unknown secret type %q; use --confirm to use it anyway", o.SecretTypeName)
			}
		}
	}

	return nil
}

func (o *CreateSecretOptions) CreateSecret() (*kapi.Secret, error) {
	secret, err := o.BundleSecret()
	if err != nil {
		return nil, err
	}

	persistedSecret, err := o.SecretsInterface.Create(secret)
	if err == nil {
		fmt.Fprintf(o.Out, "secret/%s\n", persistedSecret.Name)
	}

	return persistedSecret, err
}

func (o *CreateSecretOptions) BundleSecret() (*kapi.Secret, error) {
	secretData := make(map[string][]byte)

	for _, source := range o.Sources {
		keyName, filePath, err := parseSource(source)
		if err != nil {
			return nil, err
		}

		info, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return nil, fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return nil, fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}

		if info.IsDir() {
			if strings.Contains(source, "=") {
				return nil, errors.New("Cannot give a key name for a directory path.")
			}
			fileList, err := ioutil.ReadDir(filePath)
			if err != nil {
				return nil, fmt.Errorf("error listing files in %s: %v", filePath, err)
			}

			for _, item := range fileList {
				itemPath := path.Join(filePath, item.Name())
				if !item.Mode().IsRegular() {
					if o.Stderr != nil && o.Quiet != true {
						fmt.Fprintf(o.Stderr, "Skipping resource %s\n", itemPath)
					}
				} else {
					keyName = item.Name()
					err = addKeyToSecret(keyName, itemPath, secretData)
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			err = addKeyToSecret(keyName, filePath, secretData)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(secretData) == 0 {
		return nil, errors.New("No files selected")
	}

	// if the secret type isn't specified, attempt to auto-detect likely hit
	secretType := kapi.SecretType(o.SecretTypeName)
	if len(o.SecretTypeName) == 0 {
		secretType = kapi.SecretTypeOpaque

		for _, knownSecretType := range KnownSecretTypes {
			if knownSecretType.Matches(secretData) {
				secretType = knownSecretType.Type
			}
		}
	}

	secret := &kapi.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: o.Name},
		Type:       secretType,
		Data:       secretData,
	}

	return secret, nil
}

func addKeyToSecret(keyName, filePath string, secretData map[string][]byte) error {
	if errors := kvalidation.IsConfigMapKey(keyName); len(errors) > 0 {
		return fmt.Errorf("%v is not a valid key name for a secret: %s", keyName, strings.Join(errors, ", "))
	}
	if _, entryExists := secretData[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s from path %s, another key by that name already exists: %v.", keyName, filePath, secretData)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	secretData[keyName] = data
	return nil
}

// parseSource parses the source given. Acceptable formats include:
// source-name=source-path, where source-name will become the key name and source-path is the path to the key file
// source-path, where source-path is a path to a file or directory, and key names will default to file names
// Key names cannot include '='.
func parseSource(source string) (keyName, filePath string, err error) {
	numSeparators := strings.Count(source, "=")
	switch {
	case numSeparators == 0:
		return path.Base(source), source, nil
	case numSeparators == 1 && strings.HasPrefix(source, "="):
		return "", "", fmt.Errorf("key name for file path %v missing.", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", fmt.Errorf("file path for key name %v missing.", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", errors.New("Key names or file paths cannot contain '='.")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}
