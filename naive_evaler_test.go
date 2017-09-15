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

		// 6.10.5
		{"$.*[1 to last]", `{"x":[12,30],"y":[8],"z":["a","b","c"]}`, []string{"30", "\"b\"", "\"c\""}},

		// 6.11.1
		{"$.type()", "null", []string{"\"null\""}},
		{"$.type()", "true", []string{"\"boolean\""}},
		{"$.type()", "3", []string{"\"number\""}},
		{"$.type()", "\"hello\"", []string{"\"string\""}},
		{"$.type()", "[1,2,3]", []string{"\"array\""}},
		{"$.type()", "{\"foo\": 2}", []string{"\"object\""}},

		// 6.11.2
		{"$.size()", "null", []string{"1"}},
		{"$.size()", "true", []string{"1"}},
		{"$.size()", "3", []string{"1"}},
		{"$.size()", "\"hello\"", []string{"1"}},
		{"$.size()", "[1,2,3]", []string{"3"}},
		{"$.size()", "{\"foo\": 2}", []string{"1"}},

		// 6.11.3
		// the spec seems unclear on if .double() should error on non string-or-number values
		{"$.double()", "3", []string{"3"}},
		{"$.double()", "\"3\"", []string{"3"}},

		{"$.ceiling()", "3.3", []string{"4"}},
		{"$.floor()", "3.3", []string{"3"}},
		{"$.abs()", "-3.3", []string{"3.3"}},

		// .... fair
		// 6.11.4
		// {"$.datetime()", "-3.3", []string{"3.3"}},

		// 6.11.5
		{"$.keyvalue()", `{"foo":1, "bar":2}`, []string{`{"id":0,"name":"bar","value":2}`, `{"id":0,"name":"foo","value":1}`}},
		{"$[*].keyvalue()", `[{"foo":1, "bar":2},{"baz":3}]`, []string{`{"id":0,"name":"bar","value":2}`, `{"id":0,"name":"foo","value":1}`, `{"id":1,"name":"baz","value":3}`}},

		// 6.12.1
		{"-$[*]", `[1, 2]`, []string{"-1", "-2"}},
		{"+$[*]", `[1, 2]`, []string{"1", "2"}},
		{"-$.readings[*].floor()", `{"readings": [15.2, -22.3, 45.9] }`, []string{"-15", "23", "-45"}},
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

		{"$[*] + 2", `[1, 2]`, "binary operators can only operate on single values"},

		// strict mode:
		// {"$[1 to 0]", `[1, 2, 3]`, "the end of a range can't come before the beginning"},
		{"-$[*]", `[1, "foo"]`, "unary minus can only accept numbers"},
		{"+$[*]", `[1, "foo"]`, "unary plus can only accept numbers"},
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
