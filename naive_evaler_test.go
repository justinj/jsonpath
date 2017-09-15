package jsonpath

import (
	"encoding/json"
	"reflect"
	"sort"
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
		{"$[0 to 0]", `[1, 2, 3]`, []string{"1"}},

		{"$.*", `{"foo": 1, "bar": 2}`, []string{"1", "2"}},
		{"$[*]", `[1, 2, 3]`, []string{"1", "2", "3"}},
		{"$[*]", `[1, 2, [1, 2, 3]]`, []string{"1", "2", "[1,2,3]"}},
		{"$[*][*]", `[1, 2, [1, 2, 3]]`, []string{"1", "1", "2", "2", "3"}},
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

			// Sort the results because we don't care about order
			sort.Strings(tc.expected)
			stringResult := make([]string, len(result))
			for i, v := range result {
				s, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				stringResult[i] = string(s)
			}
			sort.Strings(stringResult)

			if !reflect.DeepEqual(tc.expected, stringResult) {
				t.Fatalf("expected %#v, got %#v", tc.expected, stringResult)
			}
		})
	}
}

// error test cases
func TestNaiveEvalErrors(t *testing.T) {
	testCases := []struct {
		input         string
		context       string
		expectedError string
	}{
		// TODO: include the object in the error
		{"$['hello']", `[1, 2, 3]`, "array index must be a number, but found \"hello\""},
		{"$[1 to 'z']", `[1, 2, 3]`, "array index must be a number, but found \"z\""},
		{"$['a' to 1]", `[1, 2, 3]`, "array index must be a number, but found \"a\""},
		{"$[100]", `[1, 2, 3]`, "array index 100 out of bounds"},
		{"$[1 to 0]", `[1, 2, 3]`, "the end of a range must be greater than the beginning"},
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

			_, err = evaler.Run(dollar)
			if err == nil {
				t.Fatal("expected error, got <nil>")
			}
			if err.Error() != tc.expectedError {
				t.Fatalf("expected %#v, got %#v", tc.expectedError, err.Error())
			}
		})
	}
}
