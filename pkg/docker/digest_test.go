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

package docker

import (
	"testing"
)

func TestImageWithoutDigest(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		want    string
		wantErr bool
	}{
		{
			name:    "kubedb/postgres:v1.2.3",
			image:   "kubedb/postgres:v1.2.3",
			want:    "kubedb/postgres:v1.2.3",
			wantErr: false,
		},
		{
			name:    "ghcr.io/myorg/postgres:v1.2.3",
			image:   "ghcr.io/myorg/postgres:v1.2.3",
			want:    "ghcr.io/myorg/postgres:v1.2.3",
			wantErr: false,
		},
		{
			name:    "ghcr.io/myorg/postgres:v1.2.3@",
			image:   "ghcr.io/myorg/postgres:v1.2.3@",
			want:    "ghcr.io/myorg/postgres:v1.2.3",
			wantErr: false,
		},
		{
			name:    "@ghcr.io/myorg/postgres:v1.2.3",
			image:   "@ghcr.io/myorg/postgres:v1.2.3",
			want:    "",
			wantErr: true,
		},
		{
			name:    "pkbhowmick/redis:latest@sha256:5c7632be083bff6d71dee3716a7e1231086e388ea70907ecb1f18f5f95ad7516",
			image:   "pkbhowmick/redis:latest@sha256:5c7632be083bff6d71dee3716a7e1231086e388ea70907ecb1f18f5f95ad7516",
			want:    "pkbhowmick/redis:latest",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ImageWithoutDigest(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageWithoutDigest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ImageWithoutDigest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
