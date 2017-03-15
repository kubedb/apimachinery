package controller

import (
	"time"

	"github.com/appscode/go/types"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func (w *Controller) EnsureDatabaseSnapshot() {
	resourceName := "database-snapshot" + "." + tapi.V1beta1SchemeGroupVersion.Group

	if _, err := w.Client.Extensions().ThirdPartyResources().Get(resourceName); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	thirdPartyResource := &extensions.ThirdPartyResource{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "ThirdPartyResource",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name: resourceName,
		},
		Versions: []extensions.APIVersion{
			{
				Name: tapi.V1beta1SchemeGroupVersion.Version,
			},
		},
	}

	if _, err := w.Client.Extensions().ThirdPartyResources().Create(thirdPartyResource); err != nil {
		log.Fatalln(err)
	}
}

func (w *Controller) EnsureDeletedDatabase() {
	resourceName := "deleted-database" + "." + tapi.V1beta1SchemeGroupVersion.Group

	if _, err := w.Client.Extensions().ThirdPartyResources().Get(resourceName); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		return
	}

	thirdPartyResource := &extensions.ThirdPartyResource{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "ThirdPartyResource",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name: resourceName,
		},
		Versions: []extensions.APIVersion{
			{
				Name: tapi.V1beta1SchemeGroupVersion.Version,
			},
		},
	}

	if _, err := w.Client.Extensions().ThirdPartyResources().Create(thirdPartyResource); err != nil {
		log.Fatalln(err)
	}
}

func (w *Controller) CheckDatabaseSnapshotJob(snapshot *tapi.DatabaseSnapshot, jobName string, checkTime float64) {

	jobSuccess := false

	then := time.Now()
	now := time.Now()
	for now.Sub(then).Minutes() < checkTime {
		log.Debugln("Checking for Job ", jobName)
		job, err := w.Client.Batch().Jobs(snapshot.Namespace).Get(jobName)
		if err != nil {
			break
		}
		log.Debugf("Pods Statuses:	%d Running / %d Succeeded / %d Failed", job.Status.Active, job.Status.Succeeded, job.Status.Failed)
		// If job is success
		if job.Status.Succeeded > 0 {
			jobSuccess = true
			break
		} else if job.Status.Failed > 0 {
			break
		}

		time.Sleep(time.Minute)
		now = time.Now()
	}

	if err := w.Client.Batch().Jobs(snapshot.Namespace).Delete(jobName, nil); err != nil {
		log.Errorln(err)
	}

	_snapshot, err := w.ExtClient.DatabaseSnapshot(snapshot.Namespace).Get(snapshot.Name)
	if err != nil {
		log.Errorln(err)
		return
	}

	unversionedNow := unversioned.Now()
	_snapshot.Status.CompletionTime = &unversionedNow
	_snapshot.Status.Active = types.BoolP(false)

	if jobSuccess {
		_snapshot.Status.Message = "Successful"
		_snapshot.Status.Succeeded = types.BoolP(true)
	} else {
		_snapshot.Status.Message = "Unsuccessful"
		_snapshot.Status.Failed = types.BoolP(true)
	}

	if _, err := w.ExtClient.DatabaseSnapshot(_snapshot.Namespace).Update(_snapshot); err != nil {
		log.Errorln(err)
	}
}
