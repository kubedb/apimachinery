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

package configgenerator

import "testing"

func TestBuildSetGlobal(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		want    string
		wantErr bool
	}{
		{"numbers and flags", map[string]string{"a.cnf": "[mysqld]\nmax_connections=300\nread_only = ON"}, "SET GLOBAL max_connections=300, read_only=ON", false},
		{"dashes, case and comments", map[string]string{"a.cnf": "# c\nWait-Timeout=100"}, "SET GLOBAL wait_timeout=100", false},
		{"size suffix", map[string]string{"a.cnf": "sort_buffer_size=1M"}, "SET GLOBAL sort_buffer_size=1048576", false},
		{"quoted number", map[string]string{"a.cnf": "max_connections=\"710\""}, "SET GLOBAL max_connections=710", false},
		{"string", map[string]string{"a.cnf": "sql_mode=\"STRICT_TRANS_TABLES\""}, "SET GLOBAL sql_mode='STRICT_TRANS_TABLES'", false},
		{"string escaping", map[string]string{"a.cnf": "x=a'b\\c"}, "SET GLOBAL x='a\\'b\\\\c'", false},
		{"multiple files sorted", map[string]string{"b.cnf": "b=2", "a.cnf": "a=1"}, "SET GLOBAL a=1, b=2", false},
		{"no variables", map[string]string{"a.cnf": "[mysqld]\n# nothing"}, "", false},
		{"bare flag", map[string]string{"a.cnf": "skip-name-resolve"}, "", true},
		{"bare flag among vars", map[string]string{"a.cnf": "[mysqld]\na=1\nskip-name-resolve"}, "", true},
		{"bad name", map[string]string{"a.cnf": "a;drop=1"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildSetGlobal(tt.files)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("BuildSetGlobal() = %q, %v; want %q, err=%v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestBuildSetGlobalSourceOrder(t *testing.T) {
	got, err := BuildSetGlobal(map[string]string{"a.cnf": "x=1"}, map[string]string{"a.cnf": "x=2"})
	if want := "SET GLOBAL x=1, x=2"; err != nil || got != want {
		t.Errorf("got %q, %v; want %q", got, err, want)
	}
}
