/*
Copyright 2023.
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

package v1alpha2

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var mssqlLog = logf.Log.WithName("mssql-resource")

var _ webhook.Defaulter = &MsSQL{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (m *MsSQL) Default() {
	if m == nil {
		return
	}
	mssqlLog.Info("default", "name", m.Name)

	// Implement defaulting logic here
}

var _ webhook.Validator = &MsSQL{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (m *MsSQL) ValidateCreate() (admission.Warnings, error) {
	mssqlLog.Info("validate create", "name", m.Name)

	allErr := m.validateMsSQL()
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "example.com", Kind: "MsSQL"}, m.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (m *MsSQL) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	mssqlLog.Info("validate update", "name", m.Name)

	oldMsSQL := old.(*MsSQL)
	allErr := m.validateMsSQL()

	// Add specific update validation logic here

	if len(allErr) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "example.com", Kind: "MsSQL"}, m.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (m *MsSQL) ValidateDelete() (admission.Warnings, error) {
	mssqlLog.Info("validate delete", "name", m.Name)

	// Add delete validation logic here if needed

	return nil, nil
}

func (m *MsSQL) validateMsSQL() field.ErrorList {
	var allErr field.ErrorList

	// Implement validation logic for MsSQL resource

	return allErr
}
