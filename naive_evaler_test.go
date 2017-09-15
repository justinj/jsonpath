package jsonpath

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNaiveEval(t *testing.T) {
	testCases := []struct {
		input    string
		context  string
		expected []string
	}{
		{"1 + 1", "{}", []string{"2"}},
		{"1 - 1", "{}", []string{"0"}},
		{"2 * 3", "{}", []string{"6"}},
		{"6 / 2", "{}", []string{"3"}},
		{"6 % 4", "{}", []string{"2"}},
		{"2 * 3 + 3", "{}", []string{"9"}},

		{"$.foo", `{"foo": 1}`, []string{"1"}},
		{"$.foo.bar", `{"foo": {"bar": 2}}`, []string{"2"}},
		{"$[0]", `[1, 2, 3]`, []string{"1"}},
		{"$[0, 2]", `[1, 2, 3]`, []string{"1", "3"}},
		{"$[last]", `[1, 2, 3]`, []string{"3"}},
		{"$[last - 1]", `[1, 2, 3]`, []string{"2"}},
		{"$[$[last] - 1]", `[1, 2, 3]`, []string{"3"}},
		{"$[0 to 1]", `[1, 2, 3]`, []string{"1", "2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			evaler, err := NewNaiveEvaler(tc.input)
			if err != nil {
				t.Fatalf("%s", err)
			}

			var dollar interface{}
			err = json.Unmarshal([]byte(tc.context), &dollar)
			if err != nil {
				t.Fatalf("couldn't decode %s: %s", tc.context, err)
			}

			result, err := evaler.Run(dollar)
			if err != nil {
				t.Fatal(err.Error())
			}
			if len(result) != len(tc.expected) {
				t.Fatalf("expected %#v, got %#v", tc.expected, result)
			}

			for i, r := range result {
				var expected interface{}
				err = json.Unmarshal([]byte(tc.expected[i]), &expected)
				if err != nil {
					t.Fatalf("couldn't decode %s: %s", tc.expected, err)
				}

				if !reflect.DeepEqual(r, expected) {
					t.Fatalf("expected %#v, got %#v", tc.expected, result)
				}
			}
		})
	}
}

// error test cases

// {"$[100]", `[1, 2, 3]`, []string{"1", "2"}},
