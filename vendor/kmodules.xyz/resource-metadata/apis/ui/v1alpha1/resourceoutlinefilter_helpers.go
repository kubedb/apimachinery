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

func (a ResourceOutlineFilterSpec) GetPage(name string) ResourcePageOutlineFilter {
	for _, page := range a.Pages {
		if page.Name == name {
			return page
		}
	}
	return ResourcePageOutlineFilter{
		Name:     name,
		Show:     false,
		Sections: nil,
	}
}

func (a ResourceOutlineFilterSpec) GetAction(name string) ActionTemplateGroupFilter {
	for _, action := range a.Actions {
		if action.Name == name {
			return action
		}
	}
	return ActionTemplateGroupFilter{
		Name:  name,
		Show:  false,
		Items: nil,
	}
}

func (a ResourcePageOutlineFilter) GetSection(name string) SectionOutlineFilter {
	for _, section := range a.Sections {
		if section.Name == name {
			return section
		}
	}
	return SectionOutlineFilter{
		Name:   name,
		Show:   false,
		Blocks: nil,
	}
}
