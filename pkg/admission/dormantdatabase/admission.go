package dormantdatabase

import (
	"sync"

	hookapi "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	plugin "github.com/kubedb/apimachinery/pkg/admission"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/reference"
)

type DormantDatabaseValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &DormantDatabaseValidator{}

func (a *DormantDatabaseValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "dormantdatabases",
		},
		"dormantdatabase"
}

func (a *DormantDatabaseValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if a.extClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (a *DormantDatabaseValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// No validation on CREATE
	if (req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindDormantDatabase {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha1().DormantDatabases(req.Namespace).Get(req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if kerr.IsNotFound(err) {
				break
			}
			if err := a.handleOwnerReferences(obj); err != nil {
				return hookapi.StatusInternalServerError(err)
			}
		}
	case admission.Update:
		// validate the operation made by User
		obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		OldObj, err := meta_util.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		if err := plugin.ValidateUpdate(obj, OldObj, req.Kind.Kind); err != nil {
			return hookapi.StatusBadRequest(err)
		}
	}

	status.Allowed = true
	return status
}

func (a *DormantDatabaseValidator) handleOwnerReferences(dormantDatabase *api.DormantDatabase) error {
	if dormantDatabase.Spec.WipeOut {
		if err := a.setOwnerReferenceToObjects(dormantDatabase); err != nil {
			return err
		}
	} else {
		if err := a.removeOwnerReferenceFromObjects(dormantDatabase); err != nil {
			return err
		}
	}
	return nil
}

func (a *DormantDatabaseValidator) setOwnerReferenceToObjects(dormantDatabase *api.DormantDatabase) error {
	// Get LabelSelector for Other Components first
	dbKind, err := meta_util.GetStringValue(dormantDatabase.ObjectMeta.Labels, api.LabelDatabaseKind)
	if err != nil {
		return err
	}
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDatabase.Name,
		api.LabelDatabaseKind: dbKind,
	}
	labelSelector := labels.SelectorFromSet(labelMap)

	// Get object reference of dormant database
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, dormantDatabase)
	if rerr != nil {
		return rerr
	}
	if err := util.SetOwnerReferenceToSnapshots(a.extClient.KubedbV1alpha1(), dormantDatabase.ObjectMeta,
		labelSelector, ref); err != nil {
		return nil
	}
	if err := core_util.SetOwnerReferenceToPVCs(a.client, dormantDatabase.ObjectMeta,
		labelSelector, ref); err != nil {
		return nil
	}

	// secretList := dormantDatabase.GetDatabaseSecrets
	//for _, sec := range secretList {
	// Add ownership to secrets
	//}

	return nil
}

func (a *DormantDatabaseValidator) removeOwnerReferenceFromObjects(dormantDatabase *api.DormantDatabase) error {
	// First, Get LabelSelector for Other Components
	dbKind, err := meta_util.GetStringValue(dormantDatabase.ObjectMeta.Labels, api.LabelDatabaseKind)
	if err != nil {
		return err
	}
	labelMap := map[string]string{
		api.LabelDatabaseName: dormantDatabase.Name,
		api.LabelDatabaseKind: dbKind,
	}
	labelSelector := labels.SelectorFromSet(labelMap)

	// Get object reference of dormant database
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, dormantDatabase)
	if rerr != nil {
		return rerr
	}
	if err := util.RemoveOwnerReferenceFromSnapshots(a.extClient.KubedbV1alpha1(), dormantDatabase.ObjectMeta,
		labelSelector, ref); err != nil {
		return nil
	}
	if err := core_util.RemoveOwnerReferenceFromPVCs(a.client, dormantDatabase.ObjectMeta,
		labelSelector, ref); err != nil {
		return nil
	}

	//secretVolList := dormantDatabase.GetDatabaseSecrets()
	// Remove ownership from secrets

	return nil
}
