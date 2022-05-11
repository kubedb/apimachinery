package health

import (
	"context"
	"sync"
	"time"

	"k8s.io/klog/v2"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

type HealthChecker struct {
	ctxCancels map[string]context.CancelFunc
	mux        sync.Mutex
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		ctxCancels: make(map[string]context.CancelFunc),
		mux:        sync.Mutex{},
	}
}

// Start creates a health check go routine.
// Call this method after successful creation of all the replicas of a database.
func (hc *HealthChecker) Start(key string, healthCheckSpec dbapi.HealthCheckSpec, fn func(context.Context, string, *HealthCard)) {
	if !hc.keyExists(key) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		hc.setCancel(key, cancel)
		ticker := time.NewTicker(time.Duration(*healthCheckSpec.PeriodSeconds) * time.Second)
		healthCheckStore := newHealthCard(*healthCheckSpec.FailureThreshold)
		go func() {
			for {
				select {
				case <-ctx.Done():
					hc.deleteCancel(key)
					cancel()
					ticker.Stop()
					klog.Infoln("Health check stopped for key " + key)
					return
				case <-ticker.C:
					ctx, cancel := context.WithTimeout(ctx, time.Duration(*healthCheckSpec.TimeoutSeconds)*time.Second)
					klog.V(5).Infoln("Health check running for key " + key)
					fn(ctx, key, healthCheckStore)
					klog.V(5).Infof("Debug client count = %d\n", healthCheckStore.GetClientCount())
					cancel()
				}
			}
		}()
	}
}

// Stop stops a health check go routine.
// Call this method when the database is deleted or halted.
func (hc *HealthChecker) Stop(key string) {
	if hc.keyExists(key) {
		hc.getCancel(key)()
		hc.deleteCancel(key)
	}
}

func (hc *HealthChecker) keyExists(key string) bool {
	hc.mux.Lock()
	defer hc.mux.Unlock()
	_, ok := hc.ctxCancels[key]
	return ok
}

func (hc *HealthChecker) getCancel(key string) context.CancelFunc {
	hc.mux.Lock()
	defer hc.mux.Unlock()
	return hc.ctxCancels[key]
}

func (hc *HealthChecker) setCancel(key string, cancel context.CancelFunc) {
	hc.mux.Lock()
	defer hc.mux.Unlock()
	hc.ctxCancels[key] = cancel
}

func (hc *HealthChecker) deleteCancel(key string) {
	hc.mux.Lock()
	defer hc.mux.Unlock()
	delete(hc.ctxCancels, key)
}
