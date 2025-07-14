/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifestwork

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"kubedb.dev/apimachinery/apis/kubedb"
	db_cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	apiworkv1 "open-cluster-management.io/api/work/v1"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	apiworkv1 "kubeops.dev/ManifestWork/apis/apps/v1"
	ManifestWorkcs "kubeops.dev/ManifestWork/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	*amc.Config

	Queue *queue.Worker[any]
	// Kubebuilder client
	KBClient client.Client
	// Kubernetes client
	Client kubernetes.Interface
	// KubeDB client
	DBClient db_cs.Interface
	// ManifestWork client
	PSClient ManifestWorkcs.Interface
	// Dynamic client
	DynamicClient dynamic.Interface
}

type Lister interface {
	List()
}

func NewController(
	config *amc.Config,
	client kubernetes.Interface,
	dbClient db_cs.Interface,
	dmClient dynamic.Interface,
	psClient ManifestWorkcs.Interface,
	q *queue.Worker[any],
) *Controller {
	return &Controller{
		Config:        config,
		Client:        client,
		DBClient:      dbClient,
		DynamicClient: dmClient,
		PSClient:      psClient,
		Queue:         q,
	}
}

func (c *Controller) WithCacheClient(kbClient client.Client) *Controller {
	c.KBClient = kbClient
	return c
}

func (c *Controller) InitManifestWorkWatcher() {
	klog.Infoln("Initializing ManifestWork watcher.....")
	// Initialize ManifestWork Watcher
	c.MWInformer = c.ManifestInformerFactory.Work().V1().ManifestWorks().Informer()
	c.MWLister = c.ManifestInformerFactory.Work().V1().ManifestWorks().Lister()
	_, _ = c.MWInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mw, ok := obj.(*apiworkv1.ManifestWork)
			if !ok {
				return
			}
			if !c.isOwnedByKubeDB(mw) {
				return
			}

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if ps, ok := newObj.(*apiworkv1.ManifestWork); ok {
				if c.RestrictToNamespace != core.NamespaceAll {
					if ps.GetNamespace() != c.RestrictToNamespace {
						klog.Infof("Skipping ManifestWork %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", ps.GetNamespace(), ps.GetName(), c.RestrictToNamespace)
						return
					}
				}

				c.enqueueConditionally(ps)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var ps *apiworkv1.ManifestWork
			var ok bool
			if ps, ok = obj.(*apiworkv1.ManifestWork); !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.V(5).Info("error decoding object, invalid type")
					return
				}
				ps, ok = tombstone.Obj.(*apiworkv1.ManifestWork)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
					return
				}
				klog.V(5).Infof("Recovered deleted object '%v' from tombstone", tombstone.Obj.(metav1.Object).GetName())
			}

			if c.RestrictToNamespace != core.NamespaceAll {
				if ps.GetNamespace() != c.RestrictToNamespace {
					klog.Infof("Skipping ManifestWork %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", ps.GetNamespace(), ps.GetName(), c.RestrictToNamespace)
					return
				}
			}

			ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(ps), kubedb.GroupName)
			if err != nil || !ok {
				klog.Warningln(err)
				return
			}
			dbInfo, err := c.extractDatabaseInfo(ps)
			if err != nil {
				if !kerr.IsNotFound(err) {
					klog.Warningf("failed to extract database info from ManifestWork: %s/%s. Reason: %v", ps.Namespace, ps.Name, err)
				}
				return
			}
			if dbInfo == nil {
				return
			}
			err = c.ensureReadyReplicasCond(dbInfo)
			if err != nil {
				klog.Warningf("failed to update ReadyReplicas condition. Reason: %v", err)
				return
			}
		},
	})
}

func (c *Controller) enqueueConditionally(ps *apiworkv1.ManifestWork) {
	if c.isOwnedByKubeDB(ps) {
		queue.Enqueue(c.PSQueue.GetQueue(), cache.ExplicitKey(ps.Namespace+"/"+ps.Name))
	}
}

func (c *Controller) isOwnedByKubeDB(ps *apiworkv1.ManifestWork) bool {
	// only enqueue if the controlling owner is a KubeDB resource
	ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(ps), kubedb.GroupName)
	if err != nil {
		klog.Warningf("failed to enqueue ManifestWork: %s/%s. Reason: %v", ps.Namespace, ps.Name, err)
		return false
	}
	return ok
}

// addManifest adds the ManifestWork for the manifestwork to the sync queue
func (c *Controller) addManifestWork(mw *apiworkv1.ManifestWork) {
	klog.Infoln("adding manifest work in the queue", mw.Name)
	if mw.DeletionTimestamp != nil {
		klog.Infoln("need to delete manifestwork", mw.Name)
		// on a restart of the controller manager, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		c.deleteManifestWork(mw)
		return
	}

	dbs := c.(mw)
	if len(sets) == 0 {
		return
	}
	logger.V(4).Info("Orphan manifeswork created with labels", "manifeswork", klog.KObj(mw), "labels", mw.Labels)
	for _, set := range sets {
		klog.Infoln("2222222222222222222222222222222222222222222222222222, ", set.Name)
		c.enqueueManifestWork(set)
	}
}

// updateManifestWork adds the ManifestWork for the current and ManifestWork pods to the sync queue.
func (c *Controller) updateManifestWork(logger klog.Logger, old, cur interface{}) {
	klog.Infoln("updating manifest work in the queue", old.(*apiworkv1.ManifestWork).Name)
	curMW := cur.(*apiworkv1.ManifestWork)
	oldMW := old.(*apiworkv1.ManifestWork)
	if curMW.ResourceVersion == oldMW.ResourceVersion {
		// In the event of a re-list we may receive update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}
	klog.Infof("updateManifest first check")

	//labelChanged := !reflect.DeepEqual(curMW.Labels, oldMW.Labels)
	//
	//curControllerRef := metav1.GetControllerOf(curMW)
	//oldControllerRef := metav1.GetControllerOf(oldMW)
	//controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	//if controllerRefChanged && oldControllerRef != nil {
	//	// The ControllerRef was changed. Sync the old controller, if any.
	//	if set := ssc.resolveControllerRef(oldMW.Namespace, oldControllerRef); set != nil {
	//		klog.Infoln("3333333333333333333333333333333333333", set.Name)
	//		ssc.enqueueManifestWork(set)
	//	}
	//}

	//// If it has a ControllerRef, that's all that matters.
	//if curControllerRef != nil {
	//	set := ssc.resolveControllerRef(curMW.Namespace, curControllerRef)
	//	if set == nil {
	//		return
	//	}
	//	logger.V(4).Info("manifesWork objectMeta updated", "manifeswork", klog.KObj(curMW), "oldObjectMeta", oldMW.ObjectMeta, "newObjectMeta", curMW.ObjectMeta)
	//	klog.Infoln("4444444444444444444444444444444444444444444", set.Name)
	//	ssc.enqueueManifestWork(set)
	//	// TODO: MinReadySeconds in the Pod will generate an Available condition to be added in
	//	// the Pod status which in turn will trigger a requeue of the owning replica set thus
	//	// having its status updated with the newly available replica.
	//	// TODO: read from feedback
	//	//if !podutil.IsPodReady(oldMW) && podutil.IsPodReady(curMW) && set.Spec.MinReadySeconds > 0 {
	//	//	logger.V(2).Info("ManifestWork will be enqueued after minReadySeconds for availability check", "statefulSet", klog.KObj(set), "minReadySeconds", set.Spec.MinReadySeconds)
	//	//	// Add a second to avoid milliseconds skew in AddAfter.
	//	//	// See https://github.com/kubernetes/kubernetes/issues/39785#issuecomment-279959133 for more info.
	//	//	ssc.enqueueSSAfter(set, (time.Duration(set.Spec.MinReadySeconds)*time.Second)+time.Second)
	//	//}
	//	return
	//}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	// if labelChanged || controllerRefChanged {
	sets := c.getManifestWorksForManifestWorks(curMW)
	klog.Infoln("len(sets)", len(sets))
	if len(sets) == 0 {
		return
	}
	logger.V(4).Info("Orphan ManifestWork objectMeta updated", "manifest", klog.KObj(curMW), "oldObjectMeta", oldMW.ObjectMeta, "newObjectMeta", curMW.ObjectMeta)
	for _, set := range sets {
		klog.Infoln("55555555555555555555555555555555555555555555555555555", set.Name)
		c.enqueueManifestWork(set)
	}
	//}
}

// deleteManifestWork enqueues the ManifestWork for the ManifestWork accounting for deletion tombstones.
func (c *Controller) deleteManifestWork(mw *apiworkv1.ManifestWork) {
	mw, ok := obj.(*apiworkv1.ManifestWork)

	// When a delete is dropped, the relist will notice an object in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
			return
		}
		mw, ok = tombstone.Obj.(*apiworkv1.ManifestWork)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a manifestwork %+v", obj))
			return
		}
	}

	klog.Infof("deleting manifest work in the queue: %s/%s", mw.Namespace, mw.Name)

	sets := c.enqueueKubedbDatabaseOnly(mw)
	if len(sets) == 0 {
		return
	}

	for _, set := range sets {
		logger.V(4).Info("ManifestWork deleted, enqueuing owner ManifestWork", "manifestwork", klog.KObj(mw), "ManifestWork", klog.KObj(set))
		c.enqueueManifestWork(set)
	}
}

// getPetSetsForPod returns a list of PetSets that potentially match
// a given pod.
func (c *Controller) enqueueKubedbDatabaseOnly(mw *apiworkv1.ManifestWork) {
	sets, err := c.setLister.GetManifestWorksPetSets(mw)
	if err != nil {
		return nil
	}
	// More than one set is selecting the same Pod
	if len(sets) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		setNames := []string{}
		for _, s := range sets {
			setNames = append(setNames, s.Name)
		}
		utilruntime.HandleError(
			fmt.Errorf(
				"user error: more than one PetSet is selecting manifeswork with labels: %+v. Sets: %v",
				mw.Labels, setNames))
	}
	return sets
}

// GetManifestWorksPetSets returns a list of PetSets that potentially match a manifestwork.
// It lists PetSets across all namespaces and matches them based on labels.
func (c *Controller) GetManifestWorksPetSets(mw *apiworkv1.ManifestWork) ([]*dbapi.Postgres, error) {
	if len(mw.Labels) == 0 {
		return nil, fmt.Errorf("no Database found for manifestwork %s because it has no labels", mw.Name)
	}

	list, err := c.DBClient.KubedbV1().Postgreses().List()
	if err != nil {
		return nil, err
	}

	var psList []*api.PetSet
	for _, ps := range list {
		selector, err := metav1.LabelSelectorAsSelector(ps.Spec.Selector)
		if err != nil {
			klog.Warningf("PetSet %s/%s has an invalid selector: %v", ps.Namespace, ps.Name, err)
			continue
		}

		if selector.Empty() || !selector.Matches(labels.Set(mw.Labels)) {
			continue
		}
		psList = append(psList, ps)
	}

	if len(psList) == 0 {
		return nil, fmt.Errorf("could not find any PetSet for manifestwork %s in any namespace with labels: %v", mw.Name, mw.Labels)
	}

	if len(psList) > 1 {
		setNames := []string{}
		for _, s := range psList {
			setNames = append(setNames, s.Name)
		}
		utilruntime.HandleError(
			fmt.Errorf(
				"user error: more than one PetSet is selecting manifestwork with labels: %+v. Sets: %v",
				mw.Labels, setNames))
	}

	return psList, nil
}
