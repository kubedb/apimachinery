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

package v1alpha1

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"stash.appscode.dev/kubestash/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ RetentionPolicy) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralRetentionPolicy))
}

// Zero returns true if the duration is empty (all values are set to zero).
func (d *Duration) Zero() bool {
	return d.Years == 0 && d.Months == 0 && d.Days == 0 && d.Hours == 0 && d.Minutes == 0
}

func (d *Duration) ToMinutes() int {
	minutes := d.Minutes
	minutes += d.Hours * 60
	minutes += d.Days * 24 * 60
	minutes += d.Months * 30 * 24 * 60
	minutes += d.Years * 365 * 24 * 60
	return minutes
}

var errInvalidDuration = errors.New("invalid duration provided")

// ParseDuration parses a duration from a string. The format is `6y5m234d37h`
func ParseDuration(s string) (Duration, error) {
	var (
		d   Duration
		num int
		err error
	)

	s = strings.TrimSpace(s)

	for s != "" {
		num, s, err = nextNumber(s)
		if err != nil {
			return Duration{}, err
		}

		if len(s) == 0 {
			return Duration{}, errInvalidDuration
		}

		if len(s) > 1 && s[0] == 'm' && s[1] == 'o' {
			d.Months = num
			s = s[2:]
			continue
		}

		switch s[0] {
		case 'y':
			d.Years = num
		case 'd':
			d.Days = num
		case 'h':
			d.Hours = num
		case 'm':
			d.Minutes = num
		default:
			return Duration{}, errInvalidDuration
		}

		s = s[1:]
	}

	return d, nil
}

func nextNumber(input string) (num int, rest string, err error) {
	if len(input) == 0 {
		return 0, "", nil
	}

	var (
		n        string
		negative bool
	)

	if input[0] == '-' {
		negative = true
		input = input[1:]
	}

	for i, s := range input {
		if !unicode.IsNumber(s) {
			rest = input[i:]
			break
		}

		n += string(s)
	}

	if len(n) == 0 {
		return 0, input, errInvalidDuration
	}

	num, err = strconv.Atoi(n)
	if err != nil {
		return 0, input, err
	}

	if negative {
		num = -num
	}

	return num, rest, nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pd, err := ParseDuration(str)
	if err != nil {
		return err
	}

	*d = pd
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d Duration) String() string {
	var s string
	if d.Years != 0 {
		s += fmt.Sprintf("%dy", d.Years)
	}

	if d.Months != 0 {
		s += fmt.Sprintf("%dmo", d.Months)
	}

	if d.Days != 0 {
		s += fmt.Sprintf("%dd", d.Days)
	}

	if d.Hours != 0 {
		s += fmt.Sprintf("%dh", d.Hours)
	}

	if d.Minutes != 0 {
		s += fmt.Sprintf("%dm", d.Minutes)
	}

	return s
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (d Duration) ToUnstructured() interface{} {
	return d.String()
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
//
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ Duration) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (_ Duration) OpenAPISchemaFormat() string { return "" }
