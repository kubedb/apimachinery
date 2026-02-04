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

package v1alpha1

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	kclient "kmodules.xyz/client-go/client"
	cutil "kmodules.xyz/client-go/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"
)

const (
	MigratorJobTriggered = "MigratorJobTriggered"

	DestroySignalSend = "DestroySignalSend"

	// Migration status conditions
	MigrationRunning       = "MigrationRunning"
	ReasonMigrationRunning = "MigrationInProgress"

	MigrationSucceeded       = "MigrationSucceeded"
	ReasonMigrationSucceeded = "MigrationCompleted"

	MigrationFailed       = "MigrationFailed"
	ReasonMigrationFailed = "MigrationError"
)

func SetMigratorJobTriggeredConditionToTrue(migrator *Migrator) {
	newCond := kmapi.Condition{
		Type:    MigratorJobTriggered,
		Status:  metav1.ConditionTrue,
		Message: "Migrator job has been triggered.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)
}

func SetMigratorJobTriggeredConditionToFalse(migrator *Migrator, err error) {
	newCond := kmapi.Condition{
		Type:    MigratorJobTriggered,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)
}

func (m *Migrator) CalculatePhase() MigratorPhase {
	if cutil.IsConditionTrue(m.Status.Conditions, MigrationSucceeded) {
		return MigratorPhaseSucceeded
	}
	if cutil.IsConditionTrue(m.Status.Conditions, MigrationFailed) {
		return MigratorPhaseFailed
	}
	if cutil.IsConditionTrue(m.Status.Conditions, MigrationRunning) {
		return MigratorPhaseRunning
	}
	if cutil.IsConditionTrue(m.Status.Conditions, MigratorJobTriggered) {
		return MigratorPhaseRunning
	}
	return MigratorPhasePending
}

func SetDestroySignalSentConditionToTrue(migrator *Migrator) {
	newCond := kmapi.Condition{
		Type:    DestroySignalSend,
		Status:  metav1.ConditionTrue,
		Message: "Destroy signal has been sent.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)
}

func SetDestroySignalSentConditionToFalse(migrator *Migrator, err error) {
	newCond := kmapi.Condition{
		Type:    DestroySignalSend,
		Status:  metav1.ConditionFalse,
		Message: err.Error(),
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)
}

// SetMigrationRunningCondition sets the condition indicating migration is in progress
func SetMigrationRunningCondition(migrator *Migrator) {
	newCond := kmapi.Condition{
		Type:    MigrationRunning,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonMigrationRunning,
		Message: "Migration is currently in progress.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)
}

// SetMigrationSucceededCondition sets the condition indicating migration completed successfully
func SetMigrationSucceededCondition(migrator *Migrator) {
	newCond := kmapi.Condition{
		Type:    MigrationSucceeded,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonMigrationSucceeded,
		Message: "Migration completed successfully.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)

	// Clear running condition
	clearCond := kmapi.Condition{
		Type:    MigrationRunning,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonMigrationSucceeded,
		Message: "Migration completed.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, clearCond)
}

// SetMigrationFailedCondition sets the condition indicating migration failed
func SetMigrationFailedCondition(migrator *Migrator, err error) {
	newCond := kmapi.Condition{
		Type:    MigrationFailed,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonMigrationFailed,
		Message: err.Error(),
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, newCond)

	// Clear running condition
	clearCond := kmapi.Condition{
		Type:    MigrationRunning,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonMigrationFailed,
		Message: "Migration failed.",
	}
	migrator.Status.Conditions = cutil.SetCondition(migrator.Status.Conditions, clearCond)
}

func CreateOrUpdateMigratorAddon(ctx context.Context, mgr manager.Manager, path string) (*MigratorAddon, error) {
	rawYAML, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read addon file from %s: %w", path, err)
	}
	addonObj := &MigratorAddon{}
	if err := yaml.Unmarshal(rawYAML, addonObj); err != nil {
		return nil, fmt.Errorf("failed to parse addon.yaml: %w", err)
	}
	uncachedClient, err := client.New(mgr.GetConfig(), client.Options{
		Scheme: mgr.GetScheme(),
		Mapper: mgr.GetRESTMapper(),
		Cache:  nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create uncache client: %w", err)
	}
	_, err = kclient.CreateOrPatch(
		ctx,
		uncachedClient,
		addonObj,
		func(obj client.Object, createOp bool) client.Object {
			in := obj.(*MigratorAddon)
			in.Spec = addonObj.Spec
			return in
		},
	)
	return addonObj, err
}
