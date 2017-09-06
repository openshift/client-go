package groups

import (
	"errors"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

const (
	AddRecommendedName    = "add-users"
	RemoveRecommendedName = "remove-users"
)

var (
	addLong = templates.LongDesc(`
		Add users to a group.

		This command will append unique users to the list of members for a group.`)

	addExample = templates.Examples(`
		# Add user1 and user2 to my-group
  	%[1]s my-group user1 user2`)

	removeLong = templates.LongDesc(`
		Remove users from a group.

		This command will remove users from the list of members for a group.`)

	removeExample = templates.Examples(`
		# Remove user1 and user2 from my-group
  	%[1]s my-group user1 user2`)
)

type GroupModificationOptions struct {
	GroupClient client.GroupInterface

	Group string
	Users []string
}

func NewCmdAddUsers(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &GroupModificationOptions{}

	cmd := &cobra.Command{
		Use:     name + " GROUP USER [USER ...]",
		Short:   "Add users to a group",
		Long:    addLong,
		Example: fmt.Sprintf(addExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			kcmdutil.CheckErr(options.AddUsers())
		},
	}

	return cmd
}

func NewCmdRemoveUsers(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &GroupModificationOptions{}

	cmd := &cobra.Command{
		Use:     name + " GROUP USER [USER ...]",
		Short:   "Remove users from a group",
		Long:    removeLong,
		Example: fmt.Sprintf(removeExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			kcmdutil.CheckErr(options.RemoveUsers())
		},
	}

	return cmd
}

func (o *GroupModificationOptions) Complete(f *clientcmd.Factory, args []string) error {
	if len(args) < 2 {
		return errors.New("you must specify at least two arguments: GROUP USER [USER ...]")
	}

	o.Group = args[0]
	o.Users = append(o.Users, args[1:]...)

	osClient, _, err := f.Clients()
	if err != nil {
		return err
	}

	o.GroupClient = osClient.Groups()

	return nil
}

func (o *GroupModificationOptions) AddUsers() error {
	group, err := o.GroupClient.Get(o.Group, metav1.GetOptions{})
	if err != nil {
		return err
	}

	existingUsers := sets.NewString(group.Users...)
	for _, user := range o.Users {
		if existingUsers.Has(user) {
			continue
		}

		group.Users = append(group.Users, user)
	}

	_, err = o.GroupClient.Update(group)
	return err
}

func (o *GroupModificationOptions) RemoveUsers() error {
	group, err := o.GroupClient.Get(o.Group, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toDelete := sets.NewString(o.Users...)
	newUsers := []string{}
	for _, user := range group.Users {
		if toDelete.Has(user) {
			continue
		}

		newUsers = append(newUsers, user)
	}
	group.Users = newUsers

	_, err = o.GroupClient.Update(group)
	return err
}
