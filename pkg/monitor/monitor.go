package monitor

import (
	"context"

	tapi "github.com/k8sdb/apimachinery/api"
	kapi "k8s.io/kubernetes/pkg/api"
)

type Monitor interface {
	AddMonitor(context.Context, *kapi.ObjectMeta, *tapi.MonitoringSpec) error
	UpdateMonitor(context.Context, *kapi.ObjectMeta, *tapi.MonitoringSpec) error
	DeleteMonitor(context.Context, *kapi.ObjectMeta, *tapi.MonitoringSpec) error
}
