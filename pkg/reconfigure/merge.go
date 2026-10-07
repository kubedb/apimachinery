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

package reconfigure

import (
	"context"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gopkg.in/yaml.v2"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MergeFiles deep-merges override on top of base, per file, the way the provisioner renders a
// config secret: the secret supplies the base document and inline applyConfig overrides it.
// It is used for diffing only; the bytes the provisioner actually writes come from its own path.
func MergeFiles(base map[string][]byte, override map[string]string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for file, content := range override {
		merged, err := mergeDocs(file, out[file], []byte(content))
		if err != nil {
			return nil, err
		}
		out[file] = merged
	}
	return out, nil
}

func mergeDocs(file string, base, override []byte) ([]byte, error) {
	baseTree, err := parseTree(file, base)
	if err != nil {
		return nil, err
	}
	overrideTree, err := parseTree(file, override)
	if err != nil {
		return nil, err
	}
	if len(overrideTree) == 0 && len(baseTree) == 0 {
		// Not a YAML mapping (an executed script, for instance) - the override replaces it whole.
		if len(override) > 0 {
			return override, nil
		}
		return base, nil
	}
	return yaml.Marshal(mergeTrees(baseTree, overrideTree))
}

func mergeTrees(base, override map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		baseChild, baseIsMap := asMap(out[k])
		overrideChild, overrideIsMap := asMap(v)
		if baseIsMap && overrideIsMap {
			out[k] = mergeTrees(baseChild, overrideChild)
			continue
		}
		out[k] = v
	}
	return out
}

// RenderConfiguration resolves a ConfigurationSpec into the rendered config files.
func RenderConfiguration(ctx context.Context, kc client.Client, namespace string, cfg *dbapi.ConfigurationSpec) (map[string][]byte, error) {
	if cfg == nil {
		return map[string][]byte{}, nil
	}
	var base map[string][]byte
	if cfg.SecretName != "" {
		var secret core.Secret
		if err := kc.Get(ctx, client.ObjectKey{Name: cfg.SecretName, Namespace: namespace}, &secret); err != nil {
			return nil, err
		}
		base = secret.Data
	}
	return MergeFiles(base, cfg.Inline)
}

// RenderReconfiguration resolves what a Reconfigure ops request would render on top of the
// database's current configuration, mirroring the operator's own merge of the two.
func RenderReconfiguration(ctx context.Context, kc client.Client, namespace string, cfg *dbapi.ConfigurationSpec, ops *opsapi.ReconfigurationSpec) (map[string][]byte, error) {
	next := &dbapi.ConfigurationSpec{}
	if ops.ConfigSecret != nil {
		next.SecretName = ops.ConfigSecret.Name
	}
	next.Inline = ops.ApplyConfig
	if !ops.RemoveCustomConfig && cfg != nil {
		if next.SecretName == "" {
			next.SecretName = cfg.SecretName
		}
		merged, err := MergeFiles(nil, cfg.Inline)
		if err != nil {
			return nil, err
		}
		inlineBase := make(map[string]string, len(merged))
		for k, v := range merged {
			inlineBase[k] = string(v)
		}
		combined, err := MergeFiles(toBytes(inlineBase), ops.ApplyConfig)
		if err != nil {
			return nil, err
		}
		next.Inline = fromBytes(combined)
	}
	return RenderConfiguration(ctx, kc, namespace, next)
}

func toBytes(in map[string]string) map[string][]byte {
	out := make(map[string][]byte, len(in))
	for k, v := range in {
		out[k] = []byte(v)
	}
	return out
}

func fromBytes(in map[string][]byte) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = string(v)
	}
	return out
}
