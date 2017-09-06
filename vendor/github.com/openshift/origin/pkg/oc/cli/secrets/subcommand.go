package secrets

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/build/builder/cmd/scmauth"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

const SecretsRecommendedName = "secrets"

const (
	// SourceUsername is the key of the optional username for basic authentication subcommand
	SourceUsername = scmauth.UsernameSecret
	// SourcePassword is the key of the optional password or token for basic authentication subcommand
	SourcePassword = scmauth.PasswordSecret
	// SourceCertificate is the key of the optional certificate authority for basic authentication subcommand
	SourceCertificate = scmauth.CACertName
	// SourcePrivateKey is the key of the required SSH private key for SSH authentication subcommand
	SourcePrivateKey = scmauth.SSHPrivateKeyMethodName
	// SourceGitconfig is the key of the optional gitconfig content for both basic and SSH authentication subcommands
	SourceGitConfig = scmauth.GitConfigName
)

var (
	secretsLong = templates.LongDesc(`
    Manage secrets in your project

    Secrets are used to store confidential information that should not be contained inside of an image.
    They are commonly used to hold things like keys for authentication to other internal systems like
    Docker registries.`)
)

func NewCmdSecrets(name, fullName string, f *clientcmd.Factory, reader io.Reader, out, errOut io.Writer, ocEditFullName string) *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:     name,
		Short:   "Manage secrets",
		Long:    secretsLong,
		Aliases: []string{"secret"},
		Run:     cmdutil.DefaultSubCommandRun(errOut),
	}

	newSecretFullName := fullName + " " + NewSecretRecommendedCommandName
	cmds.AddCommand(NewCmdCreateSecret(NewSecretRecommendedCommandName, newSecretFullName, f, out))
	cmds.AddCommand(NewCmdCreateDockerConfigSecret(CreateDockerConfigSecretRecommendedName, fullName+" "+CreateDockerConfigSecretRecommendedName, f, out, newSecretFullName, ocEditFullName))
	cmds.AddCommand(NewCmdCreateBasicAuthSecret(CreateBasicAuthSecretRecommendedCommandName, fullName+" "+CreateBasicAuthSecretRecommendedCommandName, f, reader, out, newSecretFullName, ocEditFullName))
	cmds.AddCommand(NewCmdCreateSSHAuthSecret(CreateSSHAuthSecretRecommendedCommandName, fullName+" "+CreateSSHAuthSecretRecommendedCommandName, f, out, newSecretFullName, ocEditFullName))
	cmds.AddCommand(NewCmdLinkSecret(LinkSecretRecommendedName, fullName+" "+LinkSecretRecommendedName, f, out))
	cmds.AddCommand(NewCmdUnlinkSecret(UnlinkSecretRecommendedName, fullName+" "+UnlinkSecretRecommendedName, f, out))

	return cmds
}
