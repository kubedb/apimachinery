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

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetSharedBufferSizeForPostgres(t *testing.T) {
	type args struct {
		resource *resource.Quantity
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1st",
			args: args{
				// 10GB
				resource: resource.NewQuantity(int64(1024*1024*1024*10), resource.DecimalSI),
			},
			want: "2684354560B",
		},
		{
			name: "2nd",
			args: args{
				// 1GB
				resource: resource.NewQuantity(int64(1024*1024*1024), resource.DecimalSI),
			},
			want: "268435456B",
		},
		{
			name: "3rd",
			args: args{
				resource: resource.NewQuantity(int64(1024*1024), resource.DecimalSI),
			},
			want: "262144B",
		},
		{
			name: "4th",
			args: args{
				resource: nil,
			},
			want: "131072B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSharedBufferSizeForPostgres(tt.args.resource); got != tt.want {
				t.Errorf("GetSharedBufferSizeForPostgres() = %v, want %v", got, tt.want)
			}
		})
	}
}
