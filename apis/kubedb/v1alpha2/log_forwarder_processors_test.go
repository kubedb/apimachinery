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

package v1alpha2

import (
	"reflect"
	"testing"

	core "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestLogExtensionCredentialValidation(t *testing.T) {
	selector := core.SecretKeySelector{LocalObjectReference: core.LocalObjectReference{Name: "credentials"}, Key: "token"}
	optional := true
	opt := selector.DeepCopy()
	opt.Optional = &optional
	for _, e := range []*LogExtensions{
		{ExtraConfig: "file_storage/custom: {}"},
		{ExtraConfig: "health_check/custom: {}"},
		{ExtraConfig: "basicauth/client:\n  client_auth:\n    username: " + "${env:MISSING}"},
		{SecretEnv: map[string]core.SecretKeySelector{"EXT_TOKEN": *opt}},
		{SecretEnv: map[string]core.SecretKeySelector{"POD_NAME": selector}},
		{SecretEnv: map[string]core.SecretKeySelector{"NODE_NAME": selector}},
		{SecretEnv: map[string]core.SecretKeySelector{"PATH": selector}},
		{SecretEnv: map[string]core.SecretKeySelector{"bad-name": selector}},
	} {
		lf := validLogForwarder()
		lf.Extensions = e
		if _, err := lf.ExtensionConfiguration(); err == nil {
			t.Fatal("invalid extension credentials accepted")
		}
	}
}

func TestLogInvalidTypedActions(t *testing.T) {
	for _, a := range []LogAttributeAction{
		{Key: "test", Action: "upsert"},
		{Key: "test", Action: "upsert", Value: &apiextensionsv1.JSON{Raw: []byte("{}")}},
		{Key: "test", Action: "delete", FromAttribute: "source"},
		{Key: "test", Action: "extract", Pattern: "(unnamed)"},
		{Key: "test", Action: "convert", ConvertedType: "invalid"},
	} {
		p := &LogProcessors{Attributes: &LogAttributesProcessor{Actions: []LogAttributeAction{a}}}
		if _, _, err := p.UserProcessorConfiguration(); err == nil {
			t.Fatal("invalid action accepted")
		}
	}
}

func TestLogBatchEffectiveDefaultAndNoMutation(t *testing.T) {
	lf := validLogForwarder()
	lf.Processors = &LogProcessors{Batch: &LogBatchProcessor{SendBatchMaxSize: logInt(128)}}
	if lf.ValidateCollection() == nil {
		t.Fatal("batch max below effective default")
	}
	lf.Processors = &LogProcessors{Attributes: &LogAttributesProcessor{Actions: []LogAttributeAction{{Key: "test", Action: "delete"}}}}
	before := lf.DeepCopy()
	if _, _, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, lf) {
		t.Fatal("render helper mutates input")
	}
}
