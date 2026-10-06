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
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/sets"
)

// DiffRendered compares two rendered configuration file sets and returns one Change per
// differing leaf. Files named in opaque are not parsed as YAML; any byte difference in them
// yields a single Change with an empty Path.
func DiffRendered(oldFiles, newFiles map[string][]byte, opaque sets.Set[string]) ([]Change, error) {
	var changes []Change
	for _, file := range sortedKeys(oldFiles, newFiles) {
		oldData, newData := oldFiles[file], newFiles[file]
		if bytes.Equal(oldData, newData) {
			continue
		}
		if opaque.Has(file) {
			changes = append(changes, Change{File: file, Old: len(oldData) > 0, New: len(newData) > 0})
			continue
		}
		oldTree, err := parseTree(file, oldData)
		if err != nil {
			return nil, err
		}
		newTree, err := parseTree(file, newData)
		if err != nil {
			return nil, err
		}
		changes = append(changes, diffTree(file, "", oldTree, newTree)...)
	}
	return changes, nil
}

func parseTree(file string, data []byte) (map[string]any, error) {
	tree := map[string]any{}
	if len(bytes.TrimSpace(data)) == 0 {
		return tree, nil
	}
	if err := yaml.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("failed to parse %s as yaml: %w", file, err)
	}
	return tree, nil
}

func diffTree(file, prefix string, oldTree, newTree map[string]any) []Change {
	var changes []Change
	for _, key := range sortedKeys(oldTree, newTree) {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		oldVal, inOld := oldTree[key]
		newVal, inNew := newTree[key]

		oldChild, oldIsMap := asMap(oldVal)
		newChild, newIsMap := asMap(newVal)
		switch {
		case inOld && inNew && oldIsMap && newIsMap:
			changes = append(changes, diffTree(file, path, oldChild, newChild)...)
		case reflect.DeepEqual(oldVal, newVal) && inOld == inNew:
		default:
			// A subtree replaced by a scalar (or removed outright) is reported at every leaf it
			// held, so classification sees the real parameters rather than the branch name.
			if oldIsMap || newIsMap {
				changes = append(changes, diffTree(file, path, oldChild, newChild)...)
				continue
			}
			changes = append(changes, Change{File: file, Path: path, Old: oldVal, New: newVal})
		}
	}
	return changes
}

func asMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, val := range m {
			out[fmt.Sprintf("%v", k)] = val
		}
		return out, true
	}
	return nil, false
}

func sortedKeys[V any](maps ...map[string]V) []string {
	seen := sets.New[string]()
	for _, m := range maps {
		for k := range m {
			seen.Insert(k)
		}
	}
	keys := seen.UnsortedList()
	sort.Strings(keys)
	return keys
}

// Paths renders the changed paths of a change set, for messages.
func Paths(changes []Change) string {
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		out = append(out, c.Key())
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}
