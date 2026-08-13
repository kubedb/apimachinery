/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secret

import (
	"strings"
	"testing"

	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
)

func classesOf(s string) (upper, lower, digit, symbol bool) {
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			upper = true
		case r >= 'a' && r <= 'z':
			lower = true
		case r >= '0' && r <= '9':
			digit = true
		default:
			symbol = true
		}
	}
	return
}

// The default generator must keep its three-class guarantee and must keep
// excluding symbols: several databases render the password into
// delimiter-based config formats that punctuation would corrupt.
func TestGeneratePasswordDefault(t *testing.T) {
	for i := 0; i < 200; i++ {
		pw := generatePassword(kubedb.DefaultPasswordLength)
		if len(pw) != kubedb.DefaultPasswordLength {
			t.Fatalf("length = %d, want %d", len(pw), kubedb.DefaultPasswordLength)
		}
		upper, lower, digit, symbol := classesOf(pw)
		if !upper || !lower || !digit {
			t.Fatalf("missing a required class in %q: upper=%v lower=%v digit=%v", pw, upper, lower, digit)
		}
		if symbol {
			t.Fatalf("default generator emitted a symbol in %q; symbols are deliberately excluded", pw)
		}
	}
}

// A nil PasswordGenerator must leave existing behaviour untouched. This is the
// path every database other than the one opting in takes.
func TestGeneratedDataNilGeneratorUsesDefault(t *testing.T) {
	o := Options{DefaultUsername: "root"}
	data := o.generatedData()

	if got := string(data[core.BasicAuthUsernameKey]); got != "root" {
		t.Fatalf("username = %q, want %q", got, "root")
	}
	pw := string(data[core.BasicAuthPasswordKey])
	if len(pw) != kubedb.DefaultPasswordLength {
		t.Fatalf("length = %d, want %d", len(pw), kubedb.DefaultPasswordLength)
	}
	if _, _, _, symbol := classesOf(pw); symbol {
		t.Fatalf("nil generator produced a symbol in %q", pw)
	}
}

// A non-nil PasswordGenerator must be used verbatim, and must receive the
// configured password length.
func TestGeneratedDataHonoursPasswordGenerator(t *testing.T) {
	var gotN int
	o := Options{
		DefaultUsername: "postgres",
		PasswordGenerator: func(n int) string {
			gotN = n
			return strings.Repeat("x", n)
		},
	}
	data := o.generatedData()

	if gotN != kubedb.DefaultPasswordLength {
		t.Fatalf("generator called with n = %d, want %d", gotN, kubedb.DefaultPasswordLength)
	}
	if got, want := string(data[core.BasicAuthPasswordKey]), strings.Repeat("x", kubedb.DefaultPasswordLength); got != want {
		t.Fatalf("password = %q, want %q", got, want)
	}
	if got := string(data[core.BasicAuthUsernameKey]); got != "postgres" {
		t.Fatalf("username = %q, want %q", got, "postgres")
	}
}
