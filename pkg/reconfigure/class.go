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

import "context"

type Class string

const (
	ClassReadOnly Class = "ReadOnly"
	ClassLocked   Class = "PlatformLocked"
	ClassStatic   Class = "Static"
	ClassDynamic  Class = "Dynamic"
)

// ClassStatic is the zero value on purpose: an unclassified parameter must cost a restart
// rather than be applied live and silently diverge from the running server.
func (c Class) OrStatic() Class {
	if c == "" {
		return ClassStatic
	}
	return c
}

// Change is a single differing leaf between the old and the new rendered configuration.
// Path is empty for files diffed opaquely.
type Change struct {
	File string
	Path string
	Old  any
	New  any
}

func (c Change) Key() string {
	if c.Path == "" {
		return c.File
	}
	return c.Path
}

type Classifier interface {
	Classify(ctx context.Context, changes []Change) (map[string]Class, error)
}

// Node identifies one database member a live change is applied to.
type Node struct {
	PodName  string
	NodeType string
}

type LiveApplier interface {
	Apply(ctx context.Context, node Node, changes []Change) error
	Verify(ctx context.Context, node Node, changes []Change) error
}
