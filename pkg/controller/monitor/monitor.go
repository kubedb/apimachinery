/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package monitor holds the Prometheus monitoring logic shared by every kubedb operator:
// managing the monitoring agent (ServiceMonitor create/update/delete) and reconciling the
// stats Service that the agent scrapes.
//
// The monitoring-agent-api framework (agents.New) is built on the typed kubernetes.Interface
// and the Prometheus operator client, not controller-runtime's client.Client. Agent management
// therefore keeps those clients (Options) — the documented exception to the KBClient rule. The
// stats Service, which is a plain core/v1 Service, is written through KBClient
// (StatsServiceOptions).
package monitor

import (
	"context"
	"errors"
	"fmt"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

// DB is the minimal contract a database exposes for agent management. Every kubedb DB type
// satisfies it (StatsService is a type-level helper; GetNamespace comes from metav1.Object).
type DB interface {
	StatsService() mona.StatsAccessor
	GetNamespace() string
}

// Options carries the clients the monitoring-agent-api framework needs. Unlike the rest of the
// shared code, the agent framework cannot run on KBClient: agents.New requires the typed
// kubernetes.Interface and the Prometheus operator client. This is the one documented exception
// to the "write through KBClient" rule.
type Options struct {
	Client     kubernetes.Interface
	PromClient pcm.MonitoringV1Interface
}

// newAgent builds the monitoring agent described by spec. Only the Prometheus agent is supported.
func (o Options) newAgent(spec *mona.AgentSpec) (mona.Agent, error) {
	if spec == nil {
		return nil, errors.New("MonitorSpec not found")
	}
	if spec.Prometheus != nil {
		return agents.New(spec.Agent, o.Client, o.PromClient)
	}
	return nil, fmt.Errorf("monitoring controller not found for %v", spec)
}

// addOrUpdate creates or updates the monitoring agent's resources (e.g. the ServiceMonitor).
func (o Options) addOrUpdate(db DB, spec *mona.AgentSpec) (kutil.VerbType, error) {
	agent, err := o.newAgent(spec)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(db.StatsService(), spec)
}

// Delete removes the monitoring agent's resources for db.
func (o Options) Delete(db DB, spec *mona.AgentSpec) error {
	agent, err := o.newAgent(spec)
	if err != nil {
		return err
	}
	_, err = agent.Delete(db.StatsService())
	return err
}

// getOldAgent returns the agent recorded on the stats Service via the mona.KeyAgent annotation,
// or nil if the Service is missing. It lets manageMonitor clean up after an agent-type switch.
func (o Options) getOldAgent(ctx context.Context, db DB) mona.Agent {
	service, err := o.Client.CoreV1().Services(db.GetNamespace()).Get(ctx, db.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	agent, _ := agents.New(mona.AgentType(oldAgentType), o.Client, o.PromClient)
	return agent
}

// setNewAgent records the currently configured agent type on the stats Service annotation so a
// later reconcile can detect an agent-type switch.
func (o Options) setNewAgent(ctx context.Context, db DB, spec *mona.AgentSpec) error {
	service, err := o.Client.CoreV1().Services(db.GetNamespace()).Get(ctx, db.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(ctx, o.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = meta_util.OverwriteKeys(in.Annotations, map[string]string{
			mona.KeyAgent: string(spec.Agent),
		})
		return in
	}, metav1.PatchOptions{})
	return err
}

// Manage reconciles the monitoring agent for db against spec (db.Spec.Monitor):
//   - spec set:  delete a previously-configured agent of a different type, then create/update the
//     current agent and stamp its type on the stats Service.
//   - spec nil:  delete whatever agent was previously configured.
func (o Options) Manage(ctx context.Context, db DB, spec *mona.AgentSpec) error {
	oldAgent := o.getOldAgent(ctx, db)
	if spec != nil {
		if oldAgent != nil && oldAgent.GetType() != spec.Agent {
			if _, err := oldAgent.Delete(db.StatsService()); err != nil {
				klog.Errorf("error in deleting Prometheus agent. Reason: %v", err)
			}
		}
		if _, err := o.addOrUpdate(db, spec); err != nil {
			return err
		}
		return o.setNewAgent(ctx, db, spec)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(db.StatsService()); err != nil {
			klog.Errorf("error in deleting Prometheus agent. Reason: %v", err)
		}
	}
	return nil
}
