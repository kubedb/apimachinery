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

package logforwarder

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/yaml"
)

// ParseComponents reads one native component mapping without aliases or duplicate keys.
// Errors deliberately omit input values, which may contain credential references.
func ParseComponents(raw string, reserved map[string]bool) (map[string]map[string]interface{}, error) {
	out := map[string]map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	d := yamlv3.NewDecoder(strings.NewReader(raw))
	var root yamlv3.Node
	if err := d.Decode(&root); err != nil {
		return nil, fmt.Errorf("invalid native component YAML")
	}
	var extra yamlv3.Node
	if err := d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("native components require exactly one YAML document")
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yamlv3.MappingNode {
		return nil, fmt.Errorf("native components require a mapping")
	}
	if err := validateNode(root.Content[0]); err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err := yaml.UnmarshalStrict([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("invalid native component mapping")
	}
	for id, value := range values {
		if !regexp.MustCompile(`^[a-z][a-z0-9_]*(/[A-Za-z0-9][A-Za-z0-9_.-]*)?$`).MatchString(id) {
			return nil, fmt.Errorf("invalid component ID")
		}
		typ := strings.SplitN(id, "/", 2)[0]
		if reserved[id] || reserved[typ+"/*"] {
			return nil, fmt.Errorf("component %q is managed by KubeDB", id)
		}
		cfg, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("component %q must contain an object", id)
		}
		out[id] = cfg
	}
	return out, nil
}

func validateNode(n *yamlv3.Node) error {
	if n.Kind == yamlv3.AliasNode {
		return fmt.Errorf("native configuration aliases are not supported")
	}
	if n.Kind == yamlv3.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yamlv3.ScalarNode || seen[k.Value] || k.Value == "<<" {
				return fmt.Errorf("native configuration keys must be unique scalars")
			}
			seen[k.Value] = true
		}
	}
	for _, c := range n.Content {
		if err := validateNode(c); err != nil {
			return err
		}
	}
	return nil
}

// OrderedStages requires each configured user component exactly once.
func OrderedStages(components map[string]map[string]interface{}, order []string) ([]string, error) {
	if len(order) == 0 {
		if len(components) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("explicit processor order is required")
	}
	seen := map[string]bool{}
	for _, id := range order {
		if _, ok := components[id]; !ok || seen[id] {
			return nil, fmt.Errorf("processor order contains unknown or repeated stage %q", id)
		}
		seen[id] = true
	}
	if len(seen) != len(components) {
		return nil, fmt.Errorf("processor order must list every user stage exactly once")
	}
	return append([]string(nil), order...), nil
}

// SortedIDs provides deterministic extension activation order.
func SortedIDs(components map[string]map[string]interface{}) []string {
	ids := make([]string, 0, len(components))
	for id := range components {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
