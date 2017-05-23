package monitor

import (
	"context"

	tapi "github.com/k8sdb/apimachinery/api"
	kapi "k8s.io/kubernetes/pkg/api"
)

type promuKey struct{}

func NewPrometheusContext(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, promuKey{}, params)
}

func FromPrometheusContext(ctx context.Context) (params map[string]string, ok bool) {
	params, ok = ctx.Value(promuKey{}).(map[string]string)
	return
}

type PrometheusController struct{}

func (c *PrometheusController) AddMonitor(meta *kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	err := c.ensureExporter(meta)
	if err != nil {
		return err
	}
	if ok, err := c.supportPrometheusOperator(); err != nil {
		return err
	} else if !ok {
		return nil
	}
	return c.ensureMonitor(meta, spec)
}

func (c *PrometheusController) UpdateMonitor(meta *kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	err := c.ensureExporter(meta)
	if err != nil {
		return err
	}
	if ok, err := c.supportPrometheusOperator(); err != nil {
		return err
	} else if !ok {
		return nil
	}
	return c.ensureMonitor(meta, spec)
}

func (c *PrometheusController) DeleteMonitor(meta *kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	if ok, err := c.supportPrometheusOperator(); err != nil {
		return err
	} else if !ok {
		return nil
	}
	// Delete a service monitor for this DB TPR, if does not exist
	return nil
}

func (c *PrometheusController) ensureExporter(meta *kapi.ObjectMeta) error {
	// check if the global exporter is running or not
	// if not running, create a deployment of exporter pod
	return nil
}

func (c *PrometheusController) supportPrometheusOperator() (bool, error) {
	// check if the prometheus TPR group exists
	return false, nil
}

func (c *PrometheusController) ensureMonitor(meta *kapi.ObjectMeta, spec *tapi.MonitorSpec) error {
	// Check if a service monitor exists,
	// If does not exist, create one.
	// If exists, then update it only if update is needed.
	return nil
}
