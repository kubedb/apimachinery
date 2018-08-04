package v1alpha1

import (
	"reflect"
	"testing"
)

func TestFilterTags(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]string
		out  map[string]string
	}{
		{
			"IndexRune < 0",
			map[string]string{
				"k": "v",
			},
			map[string]string{
				"k": "v",
			},
		},
		{
			"IndexRune == 0",
			map[string]string{
				"/k": "v",
			},
			map[string]string{
				"/k": "v",
			},
		},
		{
			"IndexRune < n - xyz.abc/w1",
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
		},
		{
			"IndexRune < n - .abc/w1",
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
		},
		{
			"IndexRune == n - matching_domain",
			map[string]string{
				GenericKey + "/w1": "v1",
				"w2":               "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain",
			map[string]string{
				"xyz." + GenericKey + "/w1": "v1",
				"w2": "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain-2",
			map[string]string{
				"." + GenericKey + "/w1": "v1",
				"w2": "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune == n - unmatched_domain",
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain",
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain-2",
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := filterTags(nil, c.in)
			if !reflect.DeepEqual(c.out, result) {
				t.Errorf("Failed filterTag test for '%v': expected %+v, got %+v", c.in, c.out, result)
			}
		})
	}
}
