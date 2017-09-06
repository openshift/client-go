package storage

import (
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/admin/migrate"
)

var (
	internalMigrateStorageLong = templates.LongDesc(`
		Migrate internal object storage via update

		This command invokes an update operation on every API object reachable by the caller. This forces
		the server to write to the underlying storage if the object representation has changed. Use this
		command to ensure that the most recent storage changes have been applied to all objects (storage
		version, storage encoding, any newer object defaults).

		To operate on a subset of resources, use the --include flag. If you encounter errors during a run
		the command will output a list of resources that received errors, which you can then re-run the
		command on. You may also specify --from-key and --to-key to restrict the set of resource names
		to operate on (key is NAMESPACE/NAME for resources in namespaces or NAME for cluster scoped
		resources). --from-key is inclusive if specified, while --to-key is exclusive.

		By default, events are not migrated since they expire within a very short period of time. If you
		have significantly increased the expiration time of events, run a migration with --include=events

		WARNING: This is a slow command and will put significant load on an API server. It may also
		result in significant intra-cluster traffic.`)

	internalMigrateStorageExample = templates.Examples(`
		# Perform a dry-run of updating all objects
	  %[1]s

	  # To actually perform the update, the confirm flag must be appended
	  %[1]s --confirm

	  # Only migrate pods
	  %[1]s --include=pods --confirm

	  # Only pods that are in namespaces starting with "bar"
	  %[1]s --include=pods --confirm --from-key=bar/ --to-key=bar/\xFF`)
)

type MigrateAPIStorageOptions struct {
	migrate.ResourceOptions
}

// NewCmdMigrateAPIStorage implements a MigrateStorage command
func NewCmdMigrateAPIStorage(name, fullName string, f *clientcmd.Factory, in io.Reader, out, errout io.Writer) *cobra.Command {
	options := &MigrateAPIStorageOptions{
		ResourceOptions: migrate.ResourceOptions{
			In:     in,
			Out:    out,
			ErrOut: errout,

			Include: []string{"*"},
			DefaultExcludes: []schema.GroupResource{
				// openshift resources:
				{Resource: "appliedclusterresourcequotas"},
				{Resource: "imagestreamimages"}, {Resource: "imagestreamtags"}, {Resource: "imagestreammappings"}, {Resource: "imagestreamimports"},
				{Resource: "projectrequests"}, {Resource: "projects"},
				{Resource: "clusterrolebindings"}, {Resource: "rolebindings"},
				{Resource: "clusterroles"}, {Resource: "roles"},
				{Resource: "resourceaccessreviews"}, {Resource: "localresourceaccessreviews"}, {Resource: "subjectaccessreviews"},
				{Resource: "selfsubjectrulesreviews"}, {Resource: "localsubjectaccessreviews"},
				{Resource: "useridentitymappings"},
				{Resource: "podsecuritypolicyreviews"}, {Resource: "podsecuritypolicyselfsubjectreviews"}, {Resource: "podsecuritypolicysubjectreviews"},

				// kubernetes resources:
				{Resource: "bindings"},
				{Resource: "deploymentconfigrollbacks"},
				{Resource: "events"},
				{Resource: "componentstatuses"},
				{Resource: "replicationcontrollerdummies.extensions"},
				{Resource: "podtemplates"},
				{Resource: "selfsubjectaccessreviews", Group: "authorization.k8s.io"}, {Resource: "localsubjectaccessreviews", Group: "authorization.k8s.io"},

				// skip kube RBAC resources for now because no one will have rights to update them yet
				{Resource: "roles", Group: "rbac.authorization.k8s.io"}, {Resource: "rolebindings", Group: "rbac.authorization.k8s.io"},
				{Resource: "clusterroles", Group: "rbac.authorization.k8s.io"}, {Resource: "clusterrolebindings", Group: "rbac.authorization.k8s.io"},
			},
			// Resources known to share the same storage
			OverlappingResources: []sets.String{
				// openshift resources:
				sets.NewString("deploymentconfigs.apps.openshift.io", "deploymentconfigs"),

				sets.NewString("clusterpolicies.authorization.openshift.io", "clusterpolicies"),
				sets.NewString("clusterpolicybindings.authorization.openshift.io", "clusterpolicybindings"),
				sets.NewString("clusterrolebindings.authorization.openshift.io", "clusterrolebindings"),
				sets.NewString("clusterroles.authorization.openshift.io", "clusterroles"),
				sets.NewString("localresourceaccessreviews.authorization.openshift.io", "localresourceaccessreviews"),
				sets.NewString("localsubjectaccessreviews.authorization.openshift.io", "localsubjectaccessreviews"),
				sets.NewString("policies.authorization.openshift.io", "policies"),
				sets.NewString("policybindings.authorization.openshift.io", "policybindings"),
				sets.NewString("resourceaccessreviews.authorization.openshift.io", "resourceaccessreviews"),
				sets.NewString("rolebindingrestrictions.authorization.openshift.io", "rolebindingrestrictions"),
				sets.NewString("rolebindings.authorization.openshift.io", "rolebindings"),
				sets.NewString("roles.authorization.openshift.io", "roles"),
				sets.NewString("selfsubjectrulesreviews.authorization.openshift.io", "selfsubjectrulesreviews"),
				sets.NewString("subjectaccessreviews.authorization.openshift.io", "subjectaccessreviews"),
				sets.NewString("subjectrulesreviews.authorization.openshift.io", "subjectrulesreviews"),

				sets.NewString("builds.build.openshift.io", "builds"),
				sets.NewString("buildconfigs.build.openshift.io", "buildconfigs"),

				sets.NewString("images.image.openshift.io", "images"),
				sets.NewString("imagesignatures.image.openshift.io", "imagesignatures"),
				sets.NewString("imagestreamimages.image.openshift.io", "imagestreamimages"),
				sets.NewString("imagestreamimports.image.openshift.io", "imagestreamimports"),
				sets.NewString("imagestreammappings.image.openshift.io", "imagestreammappings"),
				sets.NewString("imagestreams.image.openshift.io", "imagestreams"),
				sets.NewString("imagestreamtags.image.openshift.io", "imagestreamtags"),

				sets.NewString("clusternetworks.network.openshift.io", "clusternetworks"),
				sets.NewString("egressnetworkpolicies.network.openshift.io", "egressnetworkpolicies"),
				sets.NewString("hostsubnets.network.openshift.io", "hostsubnets"),
				sets.NewString("netnamespaces.network.openshift.io", "netnamespaces"),

				sets.NewString("oauthaccesstokens.oauth.openshift.io", "oauthaccesstokens"),
				sets.NewString("oauthauthorizetokens.oauth.openshift.io", "oauthauthorizetokens"),
				sets.NewString("oauthclientauthorizations.oauth.openshift.io", "oauthclientauthorizations"),
				sets.NewString("oauthclients.oauth.openshift.io", "oauthclients"),

				sets.NewString("projectrequests.project.openshift.io", "projectrequests"),
				sets.NewString("projects.project.openshift.io", "projects"),

				sets.NewString("appliedclusterresourcequotas.quota.openshift.io", "appliedclusterresourcequotas"),
				sets.NewString("clusterresourcequotas.quota.openshift.io", "clusterresourcequotas"),

				sets.NewString("routes.route.openshift.io", "routes"),

				sets.NewString("podsecuritypolicyreviews.security.openshift.io", "podsecuritypolicyreviews"),
				sets.NewString("podsecuritypolicyselfsubjectreviews.security.openshift.io", "podsecuritypolicyselfsubjectreviews"),
				sets.NewString("podsecuritypolicysubjectreviews.security.openshift.io", "podsecuritypolicysubjectreviews"),

				sets.NewString("processedtemplates.template.openshift.io", "processedtemplates"),
				sets.NewString("templates.template.openshift.io", "templates"),

				sets.NewString("groups.user.openshift.io", "groups"),
				sets.NewString("identities.user.openshift.io", "identities"),
				sets.NewString("useridentitymappings.user.openshift.io", "useridentitymappings"),
				sets.NewString("users.user.openshift.io", "users"),

				// kubernetes resources:
				sets.NewString("horizontalpodautoscalers.autoscaling", "horizontalpodautoscalers.extensions"),
				sets.NewString("jobs.batch", "jobs.extensions"),
			},
		},
	}
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s REGISTRY/NAME=REGISTRY/NAME [...]", name),
		Short:   "Update the stored version of API objects",
		Long:    internalMigrateStorageLong,
		Example: fmt.Sprintf(internalMigrateStorageExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete(f, cmd, args))
			kcmdutil.CheckErr(options.Validate())
			kcmdutil.CheckErr(options.Run())
		},
	}
	options.ResourceOptions.Bind(cmd)

	return cmd
}

func (o *MigrateAPIStorageOptions) Complete(f *clientcmd.Factory, c *cobra.Command, args []string) error {
	o.ResourceOptions.SaveFn = o.save
	if err := o.ResourceOptions.Complete(f, c); err != nil {
		return err
	}
	return nil
}

func (o MigrateAPIStorageOptions) Validate() error {
	return o.ResourceOptions.Validate()
}

func (o MigrateAPIStorageOptions) Run() error {
	return o.ResourceOptions.Visitor().Visit(migrate.AlwaysRequiresMigration)
}

// save invokes the API to alter an object. The reporter passed to this method is the same returned by
// the migration visitor method (for this type, transformAPIStorage). It should return an error
// if the input type cannot be saved. It returns migrate.ErrRecalculate if migration should be re-run
// on the provided object.
func (o *MigrateAPIStorageOptions) save(info *resource.Info, reporter migrate.Reporter) error {
	switch info.Object.(type) {
	// TODO: add any custom mutations necessary
	default:
		// load the body and save it back, without transformation to avoid losing fields
		get := info.Client.Get().
			Resource(info.Mapping.Resource).
			NamespaceIfScoped(info.Namespace, info.Mapping.Scope.Name() == meta.RESTScopeNameNamespace).
			Name(info.Name).Do()
		data, err := get.Raw()
		if err != nil {
			return migrate.DefaultRetriable(info, err)
		}
		update := info.Client.Put().
			Resource(info.Mapping.Resource).
			NamespaceIfScoped(info.Namespace, info.Mapping.Scope.Name() == meta.RESTScopeNameNamespace).
			Name(info.Name).Body(data).
			Do()
		if err := update.Error(); err != nil {
			return migrate.DefaultRetriable(info, err)
		}

		if oldObject, err := get.Get(); err == nil {
			info.Refresh(oldObject, true)
			oldVersion := info.ResourceVersion
			if object, err := update.Get(); err == nil {
				info.Refresh(object, true)
				if info.ResourceVersion == oldVersion {
					return migrate.ErrUnchanged
				}
			} else {
				glog.V(4).Infof("unable to calculate resource version: %v", err)
			}
		} else {
			glog.V(4).Infof("unable to calculate resource version: %v", err)
		}
	}
	return nil
}
