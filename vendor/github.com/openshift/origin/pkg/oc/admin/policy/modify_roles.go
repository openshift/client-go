package policy

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	uservalidation "github.com/openshift/origin/pkg/user/apis/user/validation"
)

const (
	AddRoleToGroupRecommendedName      = "add-role-to-group"
	AddRoleToUserRecommendedName       = "add-role-to-user"
	RemoveRoleFromGroupRecommendedName = "remove-role-from-group"
	RemoveRoleFromUserRecommendedName  = "remove-role-from-user"

	AddClusterRoleToGroupRecommendedName      = "add-cluster-role-to-group"
	AddClusterRoleToUserRecommendedName       = "add-cluster-role-to-user"
	RemoveClusterRoleFromGroupRecommendedName = "remove-cluster-role-from-group"
	RemoveClusterRoleFromUserRecommendedName  = "remove-cluster-role-from-user"
)

var (
	addRoleToUserExample = templates.Examples(`
		# Add the 'view' role to user1 for the current project
	  %[1]s view user1

	  # Add the 'edit' role to serviceaccount1 for the current project
	  %[1]s edit -z serviceaccount1`)
)

type RoleModificationOptions struct {
	RoleNamespace       string
	RoleName            string
	RoleBindingName     string
	RoleBindingAccessor RoleBindingAccessor

	Targets  []string
	Users    []string
	Groups   []string
	Subjects []kapi.ObjectReference
}

// NewCmdAddRoleToGroup implements the OpenShift cli add-role-to-group command
func NewCmdAddRoleToGroup(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " ROLE GROUP [GROUP ...]",
		Short: "Add a role to groups for the current project",
		Long:  `Add a role to groups for the current project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args, &options.Groups, "group", true); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.AddRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, true, "group", options.Targets, true, out)

		},
	}

	cmd.Flags().StringVar(&options.RoleBindingName, "rolebinding-name", "", "Name of the rolebinding to modify or create. If left empty, appends to the first rolebinding found for the given role")
	cmd.Flags().StringVar(&options.RoleNamespace, "role-namespace", "", "namespace where the role is located: empty means a role defined in cluster policy")

	return cmd
}

// NewCmdAddRoleToUser implements the OpenShift cli add-role-to-user command
func NewCmdAddRoleToUser(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}
	saNames := []string{}

	cmd := &cobra.Command{
		Use:     name + " ROLE (USER | -z SERVICEACCOUNT) [USER ...]",
		Short:   "Add a role to users or serviceaccounts for the current project",
		Long:    `Add a role to users or serviceaccounts for the current project`,
		Example: fmt.Sprintf(addRoleToUserExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.CompleteUserWithSA(f, args, saNames, true); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.AddRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, true, "user", options.Targets, true, out)
		},
	}

	cmd.Flags().StringVar(&options.RoleBindingName, "rolebinding-name", "", "Name of the rolebinding to modify or create. If left empty, appends to the first rolebinding found for the given role")
	cmd.Flags().StringVar(&options.RoleNamespace, "role-namespace", "", "namespace where the role is located: empty means a role defined in cluster policy")
	cmd.Flags().StringSliceVarP(&saNames, "serviceaccount", "z", saNames, "service account in the current namespace to use as a user")

	return cmd
}

// NewCmdRemoveRoleFromGroup implements the OpenShift cli remove-role-from-group command
func NewCmdRemoveRoleFromGroup(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " ROLE GROUP [GROUP ...]",
		Short: "Remove a role from groups for the current project",
		Long:  `Remove a role from groups for the current project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args, &options.Groups, "group", true); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.RemoveRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, false, "group", options.Targets, true, out)
		},
	}

	cmd.Flags().StringVar(&options.RoleNamespace, "role-namespace", "", "namespace where the role is located: empty means a role defined in cluster policy")

	return cmd
}

// NewCmdRemoveRoleFromUser implements the OpenShift cli remove-role-from-user command
func NewCmdRemoveRoleFromUser(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}
	saNames := []string{}

	cmd := &cobra.Command{
		Use:   name + " ROLE USER [USER ...]",
		Short: "Remove a role from users for the current project",
		Long:  `Remove a role from users for the current project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.CompleteUserWithSA(f, args, saNames, true); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.RemoveRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, false, "user", options.Targets, true, out)
		},
	}

	cmd.Flags().StringVar(&options.RoleNamespace, "role-namespace", "", "namespace where the role is located: empty means a role defined in cluster policy")
	cmd.Flags().StringSliceVarP(&saNames, "serviceaccount", "z", saNames, "service account in the current namespace to use as a user")

	return cmd
}

// NewCmdAddClusterRoleToGroup implements the OpenShift cli add-cluster-role-to-group command
func NewCmdAddClusterRoleToGroup(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " <role> <group> [group]...",
		Short: "Add a role to groups for all projects in the cluster",
		Long:  `Add a role to groups for all projects in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args, &options.Groups, "group", false); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.AddRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, true, "group", options.Targets, false, out)
		},
	}

	cmd.Flags().StringVar(&options.RoleBindingName, "rolebinding-name", "", "Name of the rolebinding to modify or create. If left empty, appends to the first rolebinding found for the given role")
	return cmd
}

// NewCmdAddClusterRoleToUser implements the OpenShift cli add-cluster-role-to-user command
func NewCmdAddClusterRoleToUser(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	saNames := []string{}
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " <role> <user | -z serviceaccount> [user]...",
		Short: "Add a role to users for all projects in the cluster",
		Long:  `Add a role to users for all projects in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.CompleteUserWithSA(f, args, saNames, false); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.AddRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, true, "user", options.Targets, false, out)
		},
	}

	cmd.Flags().StringVar(&options.RoleBindingName, "rolebinding-name", "", "Name of the rolebinding to modify or create. If left empty, appends to the first rolebinding found for the given role")
	cmd.Flags().StringSliceVarP(&saNames, "serviceaccount", "z", saNames, "service account in the current namespace to use as a user")

	return cmd
}

// NewCmdRemoveClusterRoleFromGroup implements the OpenShift cli remove-cluster-role-from-group command
func NewCmdRemoveClusterRoleFromGroup(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " <role> <group> [group]...",
		Short: "Remove a role from groups for all projects in the cluster",
		Long:  `Remove a role from groups for all projects in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, args, &options.Groups, "group", false); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.RemoveRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, false, "group", options.Targets, false, out)
		},
	}

	return cmd
}

// NewCmdRemoveClusterRoleFromUser implements the OpenShift cli remove-cluster-role-from-user command
func NewCmdRemoveClusterRoleFromUser(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	saNames := []string{}
	options := &RoleModificationOptions{}

	cmd := &cobra.Command{
		Use:   name + " <role> <user> [user]...",
		Short: "Remove a role from users for all projects in the cluster",
		Long:  `Remove a role from users for all projects in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.CompleteUserWithSA(f, args, saNames, false); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.RemoveRole(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}
			printSuccessForCommand(options.RoleName, false, "user", options.Targets, false, out)
		},
	}

	cmd.Flags().StringSliceVarP(&saNames, "serviceaccount", "z", saNames, "service account in the current namespace to use as a user")

	return cmd
}

func (o *RoleModificationOptions) CompleteUserWithSA(f *clientcmd.Factory, args []string, saNames []string, isNamespaced bool) error {
	if len(args) < 1 {
		return errors.New("you must specify a role")
	}

	o.RoleName = args[0]
	if len(args) > 1 {
		o.Users = append(o.Users, args[1:]...)
	}

	o.Targets = o.Users

	if (len(o.Users) == 0) && (len(saNames) == 0) {
		return errors.New("you must specify at least one user or service account")
	}

	osClient, _, err := f.Clients()
	if err != nil {
		return err
	}

	roleBindingNamespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	if isNamespaced {
		o.RoleBindingAccessor = NewLocalRoleBindingAccessor(roleBindingNamespace, osClient)
	} else {
		o.RoleBindingAccessor = NewClusterRoleBindingAccessor(osClient)
	}

	for _, sa := range saNames {
		o.Targets = append(o.Targets, sa)
		o.Subjects = append(o.Subjects, kapi.ObjectReference{Namespace: roleBindingNamespace, Name: sa, Kind: "ServiceAccount"})
	}

	return nil
}

func (o *RoleModificationOptions) Complete(f *clientcmd.Factory, args []string, target *[]string, targetName string, isNamespaced bool) error {
	if len(args) < 2 {
		return fmt.Errorf("you must specify at least two arguments: <role> <%s> [%s]...", targetName, targetName)
	}

	o.RoleName = args[0]
	*target = append(*target, args[1:]...)

	o.Targets = *target

	osClient, _, err := f.Clients()
	if err != nil {
		return err
	}

	if isNamespaced {
		roleBindingNamespace, _, err := f.DefaultNamespace()
		if err != nil {
			return err
		}
		o.RoleBindingAccessor = NewLocalRoleBindingAccessor(roleBindingNamespace, osClient)

	} else {
		o.RoleBindingAccessor = NewClusterRoleBindingAccessor(osClient)

	}

	return nil
}

func (o *RoleModificationOptions) getUserSpecifiedBinding() (*authorizationapi.RoleBinding, bool /* isUpdate */, error) {
	// Look for an existing rolebinding by name.
	roleBinding, err := o.RoleBindingAccessor.GetRoleBinding(o.RoleBindingName)
	if err != nil && !kapierrors.IsNotFound(err) {
		return nil, false, err
	}

	if (err != nil && kapierrors.IsNotFound(err)) || roleBinding == nil {
		// Create a new rolebinding with the desired name.
		roleBinding = &authorizationapi.RoleBinding{}
		roleBinding.Name = o.RoleBindingName
		return roleBinding, false, nil
	}

	// Check that we update the rolebinding for the intended role.
	if roleBinding.RoleRef.Name != o.RoleName || roleBinding.RoleRef.Namespace != o.RoleNamespace {
		return nil, false, fmt.Errorf("rolebinding %s found for role %s, not %s", roleBinding.Name, roleBinding.RoleRef.Name, o.RoleName)
	}

	return roleBinding, true, nil
}

func (o *RoleModificationOptions) getUnspecifiedBinding() (*authorizationapi.RoleBinding, bool /* isUpdate */, error) {
	// Look for existing bindings by role.
	roleBindings, err := o.RoleBindingAccessor.GetExistingRoleBindingsForRole(o.RoleNamespace, o.RoleName)
	if err != nil {
		return nil, false, err
	}

	if len(roleBindings) > 0 {
		// only need to add the user or group to a single roleBinding on the role.  Just choose the first one
		return roleBindings[0], true, nil
	}

	// Create a new rolebinding with the default naming.
	roleBinding := &authorizationapi.RoleBinding{}
	roleBindingNames, err := o.RoleBindingAccessor.GetExistingRoleBindingNames()
	if err != nil {
		return nil, false, err
	}

	roleBinding.Name = getUniqueName(o.RoleName, roleBindingNames)

	return roleBinding, false, nil
}

func (o *RoleModificationOptions) AddRole() error {
	var (
		roleBinding *authorizationapi.RoleBinding
		err         error
		isUpdate    bool
	)

	if len(o.RoleBindingName) > 0 {
		roleBinding, isUpdate, err = o.getUserSpecifiedBinding()
	} else {
		roleBinding, isUpdate, err = o.getUnspecifiedBinding()
	}

	if err != nil {
		return err
	}

	roleBinding.RoleRef.Namespace = o.RoleNamespace
	roleBinding.RoleRef.Name = o.RoleName

	newSubjects := authorizationapi.BuildSubjects(o.Users, o.Groups, uservalidation.ValidateUserName, uservalidation.ValidateGroupName)
	newSubjects = append(newSubjects, o.Subjects...)

subjectCheck:
	for _, newSubject := range newSubjects {
		for _, existingSubject := range roleBinding.Subjects {
			if existingSubject.Kind == newSubject.Kind &&
				existingSubject.Name == newSubject.Name &&
				existingSubject.Namespace == newSubject.Namespace {
				continue subjectCheck
			}
		}

		roleBinding.Subjects = append(roleBinding.Subjects, newSubject)
	}

	if isUpdate {
		err = o.RoleBindingAccessor.UpdateRoleBinding(roleBinding)
	} else {
		err = o.RoleBindingAccessor.CreateRoleBinding(roleBinding)
		// If the rolebinding was created in the meantime, rerun
		if kapierrors.IsAlreadyExists(err) {
			return o.AddRole()
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (o *RoleModificationOptions) RemoveRole() error {
	roleBindings, err := o.RoleBindingAccessor.GetExistingRoleBindingsForRole(o.RoleNamespace, o.RoleName)
	if err != nil {
		return err
	}
	if len(roleBindings) == 0 {
		return fmt.Errorf("unable to locate RoleBinding for %v/%v", o.RoleNamespace, o.RoleName)
	}

	subjectsToRemove := authorizationapi.BuildSubjects(o.Users, o.Groups, uservalidation.ValidateUserName, uservalidation.ValidateGroupName)
	subjectsToRemove = append(subjectsToRemove, o.Subjects...)

	for _, roleBinding := range roleBindings {
		roleBinding.Subjects = removeSubjects(roleBinding.Subjects, subjectsToRemove)

		err = o.RoleBindingAccessor.UpdateRoleBinding(roleBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeSubjects(haystack, needles []kapi.ObjectReference) []kapi.ObjectReference {
	newSubjects := []kapi.ObjectReference{}

existingLoop:
	for _, existingSubject := range haystack {
		for _, toRemove := range needles {
			if existingSubject.Kind == toRemove.Kind &&
				existingSubject.Name == toRemove.Name &&
				existingSubject.Namespace == toRemove.Namespace {
				continue existingLoop

			}
		}

		newSubjects = append(newSubjects, existingSubject)
	}

	return newSubjects
}

// prints affirmative output for role modification commands
func printSuccessForCommand(role string, didAdd bool, targetName string, targets []string, isNamespaced bool, out io.Writer) {
	verb := "removed"
	clusterScope := "cluster "
	allTargets := fmt.Sprintf("%q", targets)
	if isNamespaced {
		clusterScope = ""
	}
	if len(targets) > 1 {
		targetName = fmt.Sprintf("%ss", targetName)
	} else if len(targets) == 1 {
		allTargets = fmt.Sprintf("%q", targets[0])
	}
	if didAdd {
		verb = "added"
	}

	fmt.Fprintf(out, "%srole %q %s: %s\n", clusterScope, role, verb, allTargets)
}
