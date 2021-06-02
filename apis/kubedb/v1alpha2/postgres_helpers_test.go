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

func TestGetSharedBufferSizeForPostgres(t *testing.T) {
	type args struct {
		val int64
	}
	var tests = []struct {
		name string
		args args
		want string
	}{
		{
			name: "initial",
			args: args{
				val: 0,
			},
			want: "128MB",
		},
		{
			name: "1GB",
			args: args{
				val: 1024 * 1024 * 1024,
			},
			want: "256MB",
		},
		{
			name: "10GB",
			args: args{
				val: 1024 * 1024 * 1024 * 10,
			},
			want: "2.5GB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSharedBufferSizeForPostgres(tt.args.val); got != tt.want {
				t.Errorf("GetSharedBufferSizeForPostgres() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRound(t *testing.T) {
	type args struct {
		val     float64
		roundOn float64
		places  int
	}
	tests := []struct {
		name       string
		args       args
		wantNewVal float64
	}{
		{
			name: "1st test",
			args: args{
				val:     1.666,
				roundOn: .4,
				places:  2,
			},
			wantNewVal: 1.67,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNewVal := Round(tt.args.val, tt.args.roundOn, tt.args.places); gotNewVal != tt.wantNewVal {
				t.Errorf("Round() = %v, want %v", gotNewVal, tt.wantNewVal)
			}
		})
	}
}

func TestConvertBytesInMB(t *testing.T) {
	type args struct {
		value int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1st",
			args: args{
				value: 0,
			},
			want: "0B",
		},
		{
			name: "2nd",
			args: args{
				value: 1,
			},
			want: "1B",
		},
		{
			name: "3rd",
			args: args{
				value: 10245,
			},
			want: "10KB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesInMB(tt.args.value); got != tt.want {
				t.Errorf("ConvertBytesInMB() = %v, want %v", got, tt.want)
			}
		})
	}
}
