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

package gitops

const (
	// GroupName is the group name use in this package
	GroupName = "gitops.kubedb.com"
	// MutatorGroupName is the group name used to implement mutating webhooks for types in this package
	MutatorGroupName = "mutators." + GroupName
	// ValidatorGroupName is the group name used to implement validating webhooks for types in this package
	ValidatorGroupName = "validators." + GroupName
)
