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

package v1alpha2

import (
	"testing"
)

func TestCustomConfigGenerator_getMergedConfigString1(t *testing.T) {
	type fields struct {
		currentConfig      string
		requestedConfig    string
		configBlockDivider string
		keyValueSeparators []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				currentConfig:      "max_connections=300\nshared_buffers=256MB\n\n#*****\nmax_connections=16\nshared_buffers=22MB",
				requestedConfig:    "max_connections =150\nshared_buffers=255MB #shared_buffers=255MB\n#This is a comment = hello",
				configBlockDivider: "#*****",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "max_connections=300\nshared_buffers=256MB\n\n#*****\nmax_connections=150\nshared_buffers=255MB #shared_buffers=255MB\n",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "max_connections=300\nshared_buffers=256MB",
				requestedConfig:    "max_connections =150\nshared_buffers=255MB #shared_buffers=255MB\n#This is a comment = hello",
				configBlockDivider: "#*****",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "max_connections=300\nshared_buffers=256MB\n\n#*****\nmax_connections=150\nshared_buffers=255MB #shared_buffers=255MB\n",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "",
				requestedConfig:    "max_connections =150\nshared_buffers=255MB #shared_buffers=255MB\n#This is a comment = hello",
				configBlockDivider: "#*****",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "\n\n#*****\nmax_connections=150\nshared_buffers=255MB #shared_buffers=255MB\n",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "max_connections=300\nshared_buffers=256MB",
				requestedConfig:    "#This is a comment = hello",
				configBlockDivider: "#*****",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "max_connections=300\nshared_buffers=256MB",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "",
				requestedConfig:    "",
				configBlockDivider: "*****",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "",
				requestedConfig:    "",
				configBlockDivider: "",
				keyValueSeparators: []string{"=", " "},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "",
				requestedConfig:    "",
				configBlockDivider: "********",
				keyValueSeparators: []string{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "",
			fields: fields{
				currentConfig:      "",
				requestedConfig:    "",
				configBlockDivider: "********",
				keyValueSeparators: nil,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := &CustomConfigGenerator{
				CurrentConfig:      tt.fields.currentConfig,
				RequestedConfig:    tt.fields.requestedConfig,
				ConfigBlockDivider: tt.fields.configBlockDivider,
				KeyValueSeparators: tt.fields.keyValueSeparators,
			}
			got, err := generator.GetMergedConfigString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMergedConfigString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMergedConfigString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
