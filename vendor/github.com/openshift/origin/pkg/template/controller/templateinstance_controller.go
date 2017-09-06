package controller

import (
	"errors"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kerrs "k8s.io/apimachinery/pkg/util/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/authorization"
	kclientsetinternal "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl/resource"

	"github.com/golang/glog"

	"github.com/openshift/origin/pkg/authorization/util"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/config/cmd"
	templateapi "github.com/openshift/origin/pkg/template/apis/template"
	templateapiv1 "github.com/openshift/origin/pkg/template/apis/template/v1"
	"github.com/openshift/origin/pkg/template/generated/informers/internalversion/template/internalversion"
	templateclient "github.com/openshift/origin/pkg/template/generated/internalclientset"
	internalversiontemplate "github.com/openshift/origin/pkg/template/generated/internalclientset/typed/template/internalversion"
	templatelister "github.com/openshift/origin/pkg/template/generated/listers/template/internalversion"
)

const readinessTimeout = time.Hour

// TemplateInstanceController watches for new TemplateInstance objects and
// instantiates the template contained within, using parameters read from a
// linked Secret object.  The TemplateInstanceController instantiates objects
// using its own service account, first verifying that the requester also has
// permissions to instantiate.
type TemplateInstanceController struct {
	restmapper     meta.RESTMapper
	config         *rest.Config
	oc             client.Interface
	kc             kclientsetinternal.Interface
	templateclient internalversiontemplate.TemplateInterface

	lister   templatelister.TemplateInstanceLister
	informer cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	readinessLimiter workqueue.RateLimiter
}

// NewTemplateInstanceController returns a new TemplateInstanceController.
func NewTemplateInstanceController(config *rest.Config, oc client.Interface, kc kclientsetinternal.Interface, templateclient templateclient.Interface, informer internalversion.TemplateInstanceInformer) *TemplateInstanceController {
	c := &TemplateInstanceController{
		restmapper:       client.DefaultMultiRESTMapper(),
		config:           config,
		oc:               oc,
		kc:               kc,
		templateclient:   templateclient.Template(),
		lister:           informer.Lister(),
		informer:         informer.Informer(),
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TemplateInstanceController"),
		readinessLimiter: workqueue.NewItemFastSlowRateLimiter(5*time.Second, 20*time.Second, 200),
	}

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueue(obj.(*templateapi.TemplateInstance))
		},
		UpdateFunc: func(_, obj interface{}) {
			c.enqueue(obj.(*templateapi.TemplateInstance))
		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	return c
}

// getTemplateInstance returns the TemplateInstance from the shared informer,
// given its key (dequeued from c.queue).
func (c *TemplateInstanceController) getTemplateInstance(key string) (*templateapi.TemplateInstance, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, err
	}

	return c.lister.TemplateInstances(namespace).Get(name)
}

// copyTemplateInstance returns a deep copy of a TemplateInstance object.
func (c *TemplateInstanceController) copyTemplateInstance(templateInstance *templateapi.TemplateInstance) (*templateapi.TemplateInstance, error) {
	templateInstanceCopy, err := kapi.Scheme.DeepCopy(templateInstance)
	if err != nil {
		return nil, err
	}

	return templateInstanceCopy.(*templateapi.TemplateInstance), nil
}

// copyTemplate returns a deep copy of a Template object.
func (c *TemplateInstanceController) copyTemplate(template *templateapi.Template) (*templateapi.Template, error) {
	templateCopy, err := kapi.Scheme.DeepCopy(template)
	if err != nil {
		return nil, err
	}

	return templateCopy.(*templateapi.Template), nil
}

// sync is the actual controller worker function.
func (c *TemplateInstanceController) sync(key string) error {
	templateInstance, err := c.getTemplateInstance(key)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if templateInstance.HasCondition(templateapi.TemplateInstanceReady, kapi.ConditionTrue) ||
		templateInstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue) {
		return nil
	}

	glog.V(4).Infof("TemplateInstance controller: syncing %s", key)

	templateInstance, err = c.copyTemplateInstance(templateInstance)
	if err != nil {
		return err
	}

	if len(templateInstance.Status.Objects) != len(templateInstance.Spec.Template.Objects) {
		err = c.instantiate(templateInstance)
		if err != nil {
			glog.V(4).Infof("TemplateInstance controller: instantiate %s returned %v", key, err)

			templateInstance.SetCondition(templateapi.TemplateInstanceCondition{
				Type:    templateapi.TemplateInstanceInstantiateFailure,
				Status:  kapi.ConditionTrue,
				Reason:  "Failed",
				Message: err.Error(),
			})
		}
	}

	if !templateInstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue) {
		ready, err := c.checkReadiness(templateInstance, time.Now())
		if err != nil {
			glog.V(4).Infof("TemplateInstance controller: checkReadiness %s returned %v", key, err)

			templateInstance.SetCondition(templateapi.TemplateInstanceCondition{
				Type:    templateapi.TemplateInstanceInstantiateFailure,
				Status:  kapi.ConditionTrue,
				Reason:  "Failed",
				Message: err.Error(),
			})
			templateInstance.SetCondition(templateapi.TemplateInstanceCondition{
				Type:    templateapi.TemplateInstanceReady,
				Status:  kapi.ConditionFalse,
				Reason:  "Failed",
				Message: "See InstantiateFailure condition for error message",
			})

		} else if ready {
			templateInstance.SetCondition(templateapi.TemplateInstanceCondition{
				Type:   templateapi.TemplateInstanceReady,
				Status: kapi.ConditionTrue,
				Reason: "Created",
			})

		} else {
			templateInstance.SetCondition(templateapi.TemplateInstanceCondition{
				Type:    templateapi.TemplateInstanceReady,
				Status:  kapi.ConditionFalse,
				Reason:  "Waiting",
				Message: "Waiting for instantiated objects to report ready",
			})
		}
	}

	_, err = c.templateclient.TemplateInstances(templateInstance.Namespace).UpdateStatus(templateInstance)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("TemplateInstance status update failed: %v", err))
		return err
	}

	if !templateInstance.HasCondition(templateapi.TemplateInstanceReady, kapi.ConditionTrue) &&
		!templateInstance.HasCondition(templateapi.TemplateInstanceInstantiateFailure, kapi.ConditionTrue) {
		c.enqueueAfter(templateInstance, c.readinessLimiter.When(key))
	} else {
		c.readinessLimiter.Forget(key)
	}

	return nil
}

func (c *TemplateInstanceController) checkReadiness(templateInstance *templateapi.TemplateInstance, now time.Time) (bool, error) {
	if now.After(templateInstance.CreationTimestamp.Add(readinessTimeout)) {
		return false, fmt.Errorf("Timeout")
	}

	u := &user.DefaultInfo{Name: templateInstance.Spec.Requester.Username}

	for _, object := range templateInstance.Status.Objects {
		if !canCheckReadiness(object.Ref) {
			continue
		}

		mapping, err := c.restmapper.RESTMapping(object.Ref.GroupVersionKind().GroupKind())
		if err != nil {
			return false, err
		}

		if err = util.Authorize(c.kc.Authorization().SubjectAccessReviews(), u, &authorization.ResourceAttributes{
			Namespace: object.Ref.Namespace,
			Verb:      "get",
			Group:     object.Ref.GroupVersionKind().Group,
			Resource:  mapping.Resource,
			Name:      object.Ref.Name,
		}); err != nil {
			return false, err
		}

		cli, err := cmd.ClientMapperFromConfig(c.config).ClientForMapping(mapping)
		if err != nil {
			return false, err
		}

		obj, err := cli.Get().Resource(mapping.Resource).NamespaceIfScoped(object.Ref.Namespace, mapping.Scope.Name() == meta.RESTScopeNameNamespace).Name(object.Ref.Name).Do().Get()
		if err != nil {
			return false, err
		}

		meta, err := meta.Accessor(obj)
		if err != nil {
			return false, err
		}

		if meta.GetUID() != object.Ref.UID {
			return false, kerrors.NewNotFound(schema.GroupResource{Group: mapping.GroupVersionKind.Group, Resource: mapping.Resource}, object.Ref.Name)
		}

		if strings.ToLower(meta.GetAnnotations()[templateapi.WaitForReadyAnnotation]) != "true" {
			continue
		}

		ready, failed, err := checkReadiness(c.oc, object.Ref, obj)
		if err != nil {
			return false, err
		}
		if failed {
			return false, fmt.Errorf("Readiness failed on %s %s/%s", object.Ref.Kind, object.Ref.Namespace, object.Ref.Name)
		}
		if !ready {
			return false, nil
		}
	}

	return true, nil
}

// Run runs the controller until stopCh is closed, with as many workers as
// specified.
func (c *TemplateInstanceController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return
	}

	glog.V(2).Infof("Starting TemplateInstance controller")

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

// runWorker repeatedly calls processNextWorkItem until the latter wants to
// exit.
func (c *TemplateInstanceController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem reads from the queue and calls the sync worker function.
// It returns false only when the queue is closed.
func (c *TemplateInstanceController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.sync(key.(string))
	if err == nil { // for example, success, or the TemplateInstance has gone away
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("TemplateInstance %v failed with: %v", key, err))
	c.queue.AddRateLimited(key) // avoid hot looping

	return true
}

// enqueue adds a TemplateInstance to c.queue.  This function is called on the
// shared informer goroutine.
func (c *TemplateInstanceController) enqueue(templateInstance *templateapi.TemplateInstance) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(templateInstance)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", templateInstance, err))
		return
	}

	c.queue.Add(key)
}

// enqueueAfter adds a TemplateInstance to c.queue after a duration.
func (c *TemplateInstanceController) enqueueAfter(templateInstance *templateapi.TemplateInstance, duration time.Duration) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(templateInstance)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", templateInstance, err))
		return
	}

	c.queue.AddAfter(key, duration)
}

// instantiate instantiates the objects contained in a TemplateInstance.  Any
// parameters for instantiation are contained in the Secret linked to the
// TemplateInstance.
func (c *TemplateInstanceController) instantiate(templateInstance *templateapi.TemplateInstance) error {
	if templateInstance.Spec.Requester == nil || templateInstance.Spec.Requester.Username == "" {
		return fmt.Errorf("spec.requester.username not set")
	}

	extra := map[string][]string{}
	for k, v := range templateInstance.Spec.Requester.Extra {
		extra[k] = []string(v)
	}

	u := &user.DefaultInfo{
		Name:   templateInstance.Spec.Requester.Username,
		UID:    templateInstance.Spec.Requester.UID,
		Groups: templateInstance.Spec.Requester.Groups,
		Extra:  extra,
	}

	var secret *kapi.Secret
	if templateInstance.Spec.Secret != nil {
		if err := util.Authorize(c.kc.Authorization().SubjectAccessReviews(), u, &authorization.ResourceAttributes{
			Namespace: templateInstance.Namespace,
			Verb:      "get",
			Group:     kapi.GroupName,
			Resource:  "secrets",
			Name:      templateInstance.Spec.Secret.Name,
		}); err != nil {
			return err
		}

		s, err := c.kc.Core().Secrets(templateInstance.Namespace).Get(templateInstance.Spec.Secret.Name, metav1.GetOptions{})
		secret = s
		if err != nil {
			return err
		}
	}

	template, err := c.copyTemplate(&templateInstance.Spec.Template)
	if err != nil {
		return err
	}

	// We label all objects we create - this is needed by the template service
	// broker.
	if template.ObjectLabels == nil {
		template.ObjectLabels = make(map[string]string)
	}
	template.ObjectLabels[templateapi.TemplateInstanceLabel] = templateInstance.Name

	if secret != nil {
		for i, param := range template.Parameters {
			if value, ok := secret.Data[param.Name]; ok {
				template.Parameters[i].Value = string(value)
				template.Parameters[i].Generate = ""
			}
		}
	}

	if err := util.Authorize(c.kc.Authorization().SubjectAccessReviews(), u, &authorization.ResourceAttributes{
		Namespace: templateInstance.Namespace,
		Verb:      "create",
		Group:     templateapi.GroupName,
		Resource:  "templateconfigs",
		Name:      template.Name,
	}); err != nil {
		return err
	}

	glog.V(4).Infof("TemplateInstance controller: creating TemplateConfig for %s/%s", templateInstance.Namespace, templateInstance.Name)

	template, err = c.oc.TemplateConfigs(templateInstance.Namespace).Create(template)
	if err != nil {
		return err
	}

	errs := runtime.DecodeList(template.Objects, kapi.Codecs.UniversalDecoder())
	if len(errs) > 0 {
		return kerrs.NewAggregate(errs)
	}

	// We add an OwnerReference to all objects we create - this is also needed
	// by the template service broker for cleanup.
	for _, obj := range template.Objects {
		meta, _ := meta.Accessor(obj)
		ref := meta.GetOwnerReferences()
		ref = append(ref, metav1.OwnerReference{
			APIVersion: templateapiv1.SchemeGroupVersion.String(),
			Kind:       "TemplateInstance",
			Name:       templateInstance.Name,
			UID:        templateInstance.UID,
		})
		meta.SetOwnerReferences(ref)
	}

	bulk := cmd.Bulk{
		Mapper: &resource.Mapper{
			RESTMapper:   c.restmapper,
			ObjectTyper:  kapi.Scheme,
			ClientMapper: cmd.ClientMapperFromConfig(c.config),
		},
		Op: func(info *resource.Info, namespace string, obj runtime.Object) (runtime.Object, error) {
			if len(info.Namespace) > 0 {
				namespace = info.Namespace
			}
			if namespace == "" {
				return nil, errors.New("namespace was empty")
			}
			if info.Mapping.Resource == "" {
				return nil, errors.New("resource was empty")
			}
			if err := util.Authorize(c.kc.Authorization().SubjectAccessReviews(), u, &authorization.ResourceAttributes{
				Namespace: namespace,
				Verb:      "create",
				Group:     info.Mapping.GroupVersionKind.Group,
				Resource:  info.Mapping.Resource,
				Name:      info.Name,
			}); err != nil {
				return nil, err
			}
			return obj, nil
		},
	}

	// First, do all the SARs to ensure the requester actually has permissions
	// to create.
	glog.V(4).Infof("TemplateInstance controller: running SARs for %s/%s", templateInstance.Namespace, templateInstance.Name)

	errs = bulk.Run(&kapi.List{Items: template.Objects}, templateInstance.Namespace)
	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}

	bulk.Op = func(info *resource.Info, namespace string, obj runtime.Object) (runtime.Object, error) {
		// as cmd.Create, but be tolerant to the existence of objects that we
		// created before.
		helper := resource.NewHelper(info.Client, info.Mapping)
		if len(info.Namespace) > 0 {
			namespace = info.Namespace
		}
		createObj, createErr := helper.Create(namespace, false, obj)
		if kerrors.IsAlreadyExists(createErr) {
			obj, err := helper.Get(namespace, info.Name, false)
			if err != nil {
				return nil, err
			}

			meta, err := meta.Accessor(obj)
			if err != nil {
				return nil, err
			}

			if meta.GetLabels()[templateapi.TemplateInstanceLabel] == templateInstance.Name {
				createObj, createErr = obj, nil
			}
		}

		if createErr != nil {
			return createObj, createErr
		}

		meta, err := meta.Accessor(createObj)
		if err != nil {
			return nil, err
		}

		templateInstance.Status.Objects = append(templateInstance.Status.Objects,
			templateapi.TemplateInstanceObject{
				Ref: kapi.ObjectReference{
					Kind:       info.Mapping.GroupVersionKind.Kind,
					Namespace:  namespace,
					Name:       info.Name,
					UID:        meta.GetUID(),
					APIVersion: info.Mapping.GroupVersionKind.GroupVersion().String(),
				},
			},
		)

		return createObj, nil
	}

	// Second, create the objects, being tolerant if they already exist and are
	// labelled as having previously been created by us.
	glog.V(4).Infof("TemplateInstance controller: creating objects for %s/%s", templateInstance.Namespace, templateInstance.Name)

	templateInstance.Status.Objects = nil

	errs = bulk.Run(&kapi.List{Items: template.Objects}, templateInstance.Namespace)
	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}

	return nil
}
