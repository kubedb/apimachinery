/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"github.com/mikefarah/yq/v3/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

type YqStringsLib struct {
	YqLib yqlib.YqLib
}

func NewYqStringLib() YqStringsLib {
	return YqStringsLib{YqLib: yqlib.NewYqLib()}
}

func (yqs *YqStringsLib) NodeToNodeContexts(node *yaml.Node, arrayMergeStrategy yqlib.ArrayMergeStrategy) ([]*yqlib.NodeContext, error) {
	return yqs.YqLib.GetForMerge(node, "**", arrayMergeStrategy)
}
