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

func hasSymbol(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		default:
			return true
		}
	}
	return false
}

// The default generator must keep excluding symbols: several databases render
// the password into delimiter-based config formats that punctuation would
// corrupt. It must also not repeat.
func TestGeneratePasswordDefault(t *testing.T) {
	seen := make(map[string]struct{}, 500)
	for i := 0; i < 500; i++ {
		pw := generatePassword(kubedb.DefaultPasswordLength)
		if len(pw) != kubedb.DefaultPasswordLength {
			t.Fatalf("length = %d, want %d", len(pw), kubedb.DefaultPasswordLength)
		}
		if hasSymbol(pw) {
			t.Fatalf("default generator emitted a symbol in %q; symbols are deliberately excluded", pw)
		}
		if _, dup := seen[pw]; dup {
			t.Fatalf("duplicate password %q after %d draws", pw, i)
		}
		seen[pw] = struct{}{}
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
	if hasSymbol(pw) {
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
