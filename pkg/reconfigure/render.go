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
	"fmt"
	"sort"
	"strings"
)

// RenderClasses is the compact per-path listing surfaced on the ops request condition, so
// `kubectl describe` shows why a restart did or did not happen.
func RenderClasses(d Decision) string {
	out := make([]string, 0, len(d.Classes))
	for path, class := range d.Classes {
		out = append(out, fmt.Sprintf("%s=%s", path, class))
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// Explainer supplies the per-parameter reason and remedy shown when a change is denied.
type Explainer interface {
	Explain(c Change, class Class) string
}

func RenderRejection(d Decision, field string, ex Explainer) string {
	var b strings.Builder
	for _, c := range sortedChanges(d.Rejected) {
		class := d.Classes[c.Key()]
		fmt.Fprintf(&b, "%s: Invalid value: %q: %s\n", field, c.Key(), ex.Explain(c, class))
	}
	if len(d.Blocked) > 0 {
		fmt.Fprintf(&b, "%s: Invalid value: %q: ", field, Paths(d.Blocked))
		b.WriteString("these parameters only take effect after a restart, but restart is set to \"false\". ")
		b.WriteString("Set restart to \"auto\" (or \"true\") to apply them with a rolling restart, ")
		b.WriteString("or remove them from this request and change only runtime-settable parameters.\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func sortedChanges(changes []Change) []Change {
	out := append([]Change(nil), changes...)
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

// ParseClasses reads back what RenderClasses wrote. The classification is persisted on the ops
// request so that later reconcile passes, which run after the configuration has already been
// patched and can no longer diff it, still know what each changed parameter was.
func ParseClasses(s string) map[string]Class {
	out := map[string]Class{}
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		path, class, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		out[path] = Class(class)
	}
	return out
}

// ChangesFor rebuilds a change set from a persisted classification and the currently rendered
// configuration, so the new value of each parameter is the one now on disk.
func ChangesFor(files map[string][]byte, classes map[string]Class) ([]Change, error) {
	changes := make([]Change, 0, len(classes))
	for path := range classes {
		if _, ok := files[path]; ok { // an opaque file, keyed by file name
			changes = append(changes, Change{File: path})
			continue
		}
		file, value, _, err := LookupLeaf(files, path)
		if err != nil {
			return nil, err
		}
		// A path that no longer resolves was removed; it keeps its entry with a nil value so the
		// change still counts, and Decide downgrades it to Static.
		changes = append(changes, Change{File: file, Path: path, New: value})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Key() < changes[j].Key() })
	return changes, nil
}

func LookupLeaf(files map[string][]byte, path string) (string, any, bool, error) {
	for _, file := range sortedKeys(files) {
		tree, err := parseTree(file, files[file])
		if err != nil {
			return "", nil, false, err
		}
		var cur any = tree
		found := true
		for _, seg := range strings.Split(path, ".") {
			m, ok := asMap(cur)
			if !ok {
				found = false
				break
			}
			cur, found = m[seg]
			if !found {
				break
			}
		}
		if found {
			return file, cur, true, nil
		}
	}
	return "", nil, false, nil
}
