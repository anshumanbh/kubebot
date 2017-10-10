package flatmap

import (
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	cases := []struct {
		Input  map[string]interface{}
		Output map[string]string
	}{
		{
			Input: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Output: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},
		{
			Input: map[string]interface{}{
				"foo": []string{
					"one",
					"two",
				},
			},
			Output: map[string]string{
				"foo.0": "one",
				"foo.1": "two",
			},
		},
		{
			Input: map[string]interface{}{
				"foo": []map[interface{}]interface{}{
					map[interface{}]interface{}{
						"name":    "bar",
						"port":    3000,
						"enabled": true,
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.port":    "3000",
				"foo.0.enabled": "true",
			},
		},
		{
			Input: map[string]interface{}{
				"foo": []map[interface{}]interface{}{
					map[interface{}]interface{}{
						"name": "bar",
						"ports": []string{
							"1",
							"2",
						},
					},
				},
			},
			Output: map[string]string{
				"foo.0.name":    "bar",
				"foo.0.ports.0": "1",
				"foo.0.ports.1": "2",
			},
		},
		{
			Input: map[string]interface{}{
				"foo": struct {
					Name string
					Age  int
				}{
					"astaxie",
					30,
				},
			},
			Output: map[string]string{
				"foo.Name": "astaxie",
				"foo.Age":  "30",
			},
		},
	}

	for _, tc := range cases {
		result, err := Flatten(tc.Input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, FlatMap(tc.Output)) {
			t.Fatalf(
				"Input:\n\n%#v\n\nOutput:\n\n%#v\n\nExpected:\n\n%#v\n",
				tc.Input,
				result,
				tc.Output)
		}
	}
}
