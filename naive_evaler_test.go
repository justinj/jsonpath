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
		{"lax 1 + 1", "{}", []string{"2"}},
		{"lax 1 - 1", "{}", []string{"0"}},
		{"lax 2 * 3", "{}", []string{"6"}},
		{"lax 6 / 2", "{}", []string{"3"}},
		{"lax 6 % 4", "{}", []string{"2"}},
		{"lax 2 * 3 + 3", "{}", []string{"9"}},

		{"lax $.foo", `{"foo": 1}`, []string{"1"}},
		{"lax $.foo", `{}`, []string{}},
		{"lax $.foo", `[{"foo": 1}, {"foo": 2}]`, []string{"1", "2"}},
		{"lax $.foo", `[{"foo": 1}, {"bar": 2}]`, []string{"1"}},
		{"lax $.foo.bar", `{"foo": {"bar": 2}}`, []string{"2"}},
		{"strict $.phones[*] ? (exists (@.type)).type",
			`{ "phones": [
			{ "type": "cell", "number": "abc-defg" },
			{                 "number": "pqr-wxyz" },
			{ "type": "home", "number": "hij-klmn" } ] }`,
			[]string{"\"cell\"", "\"home\""}},
		{"lax $[0]", `[1, 2, 3]`, []string{"1"}},
		{"lax $[0, 2]", `[1, 2, 3]`, []string{"1", "3"}},
		{"lax $[last]", `[1, 2, 3]`, []string{"3"}},
		{"lax $[last - 1]", `[1, 2, 3]`, []string{"2"}},
		{"lax $[$[last] - 1]", `[1, 2, 3]`, []string{"3"}},
		{"lax $[0 to 1]", `[1, 2, 3]`, []string{"1", "2"}},
		{"lax $[0 to 0]", `[1, 2, 3]`, []string{"1"}},
		{"lax $[100]", `[1, 2, 3]`, []string{}},
		{"lax $[0 to 100]", `[1, 2, 3]`, []string{"1", "2", "3"}},
		{"lax $[0 to 100]", `"hi"`, []string{"\"hi\""}},

		{"lax $.*", `{"foo": 1, "bar": 2}`, []string{"1", "2"}},
		{"lax $.*", `[{"foo": 1, "bar": 2}]`, []string{"1", "2"}},
		{"lax 'foo'.*", `{}`, []string{}},
		{"lax $[*]", `[1, 2, 3]`, []string{"1", "2", "3"}},
		{"lax $[*].foo", `[{"foo": 1}, {"foo": 2}]`, []string{"1", "2"}},
		{"lax $[*]", `[1, 2, [1, 2, 3]]`, []string{"1", "2", "[1,2,3]"}},
		{"lax $[*][*]", `[1, 2, [1, 2, 3]]`, []string{"1", "1", "2", "2", "3"}},

		// 6.10.5
		{"lax $.*[1 to last]", `{"x":[12,30],"y":[8],"z":["a","b","c"]}`, []string{"30", "\"b\"", "\"c\""}},

		// 6.11.1
		{"lax $.type()", "null", []string{"\"null\""}},
		{"lax $.type()", "true", []string{"\"boolean\""}},
		{"lax $.type()", "3", []string{"\"number\""}},
		{"lax $.type()", "\"hello\"", []string{"\"string\""}},
		{"lax $.type()", "[1,2,3]", []string{"\"array\""}},
		{"lax $.type()", "{\"foo\": 2}", []string{"\"object\""}},

		// 6.11.2
		{"lax $.size()", "null", []string{"1"}},
		{"lax $.size()", "true", []string{"1"}},
		{"lax $.size()", "3", []string{"1"}},
		{"lax $.size()", "\"hello\"", []string{"1"}},
		{"lax $.size()", "[1,2,3]", []string{"3"}},
		{"lax $.size()", "{\"foo\": 2}", []string{"1"}},

		// 6.11.3
		// the spec seems unclear on if .double() should error on non string-or-number values
		{"lax $.double()", "3", []string{"3"}},
		{"lax $.double()", "\"3\"", []string{"3"}},

		{"lax $.ceiling()", "3.3", []string{"4"}},
		{"lax $.floor()", "3.3", []string{"3"}},
		{"lax $.abs()", "-3.3", []string{"3.3"}},

		// .... fair
		// 6.11.4
		// {"$.datetime()", "-3.3", []string{"3.3"}},

		// 6.11.5
		{"lax $.keyvalue()", `{"foo":1, "bar":2}`, []string{`{"id":0,"name":"bar","value":2}`, `{"id":0,"name":"foo","value":1}`}},
		{"lax $[*].keyvalue()", `[{"foo":1, "bar":2},{"baz":3}]`, []string{`{"id":0,"name":"bar","value":2}`, `{"id":0,"name":"foo","value":1}`, `{"id":1,"name":"baz","value":3}`}},
		{"lax $.keyvalue()", `[{"foo":1, "bar":2},{"baz":3}]`, []string{`{"id":0,"name":"bar","value":2}`, `{"id":0,"name":"foo","value":1}`, `{"id":1,"name":"baz","value":3}`}},

		// 6.12.1
		{"lax -$[*]", `[1, 2]`, []string{"-1", "-2"}},
		{"lax +$[*]", `[1, 2]`, []string{"1", "2"}},
		{"lax -$.readings[*].floor()", `{"readings": [15.2, -22.3, 45.9] }`, []string{"-15", "23", "-45"}},
		{"strict -$.readings[*].floor()", `{"readings": [15.2, -22.3, 45.9] }`, []string{"-15", "23", "-45"}},
		{"lax -$.readings.floor()", `{"readings": [15.2, -22.3, 45.9] }`, []string{"-15", "23", "-45"}},

		// 6.13
		{"lax $[*] ? (true == true)", `[1, 2, 3]`, []string{`1`, `2`, `3`}},
		{"lax $[*] ? (@ == 2)", `[1, 2, 3]`, []string{`2`}},
		{"lax $[*] ? (@ == \"foo\")", `["foo", "bar"]`, []string{`"foo"`}},
		{"lax 1 ? ($[*][0] == $[*][1])", `[[1, 2], [2, 3]]`, []string{`1`}},
		{"lax 1 ? (null == null)", `{}`, []string{"1"}},
		{"strict $ ? (@.hours > 9)", `{ "pay": 100, "horas": 10 }`, []string{}},
		{"lax 1 ? ($[*][0] > $[*][1])", `[[1, 2], [2, 3]]`, []string{}},
		{"lax 1 ? ($[*][0] < $[*][1])", `[[1, 2], [2, 3]]`, []string{`1`}},
		{"lax 1 ? ('abc' < 'xyz')", `{}`, []string{`1`}},
		{"lax 1 ? ('abc' > 'xyz')", `{}`, []string{}},
		{"lax 1 ? (true < true)", `{}`, []string{}},
		{"lax 1 ? (false < true)", `{}`, []string{`1`}},
		{"lax 1 ? (true < false)", `{}`, []string{}},
		{"lax 1 ? (false <= false)", `{}`, []string{`1`}},
		{"lax 1 ? (true <= false)", `{}`, []string{}},
		{"lax 1 ? (false <= true)", `{}`, []string{`1`}},
		{"lax 1 ? (false != true)", `{}`, []string{`1`}},

		// 6.13.6
		{"lax $[*] ? (@ like_regex 'foo')", `["foo", "bar", "afoob"]`, []string{"\"foo\"", "\"afoob\""}},

		// 6.13.7
		{"lax $[*] ? (@ starts with 'foo')", `["foo", "bar", "afoob"]`, []string{"\"foo\""}},

		// 6.13.8
		{"lax true ? (exists ($.foo))", `{"foo": 1}`, []string{"true"}},
		{"lax true ? (exists ($.foo))", `{"bar": 1}`, []string{}},

		// 6.13.9
		{"lax true ? ((1 == 'one') is unknown)", `{}`, []string{"true"}},
		{"lax true ? (((1) == 'one') is unknown)", `{}`, []string{"true"}},
		{"lax true ? (((1 == 'one') is unknown))", `{}`, []string{"true"}},
		{"lax true ? ((1 == 1) is unknown)", `{}`, []string{}},

		// ?
		// {"1 ? ($[0] == $[1])", `[[1, 2], [2, 3]]`, []string{`1`}},
		// ?
		// {"$[*] ? (@[*] == 2)", `[[1, 2, 3]]`, []string{`1`, `2`, `3`}},
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
		{"strict $['hello']", `[1, 2, 3]`, "array index must be a number, but found \"hello\""},
		{"lax $['hello']", `[1, 2, 3]`, "array index must be a number, but found \"hello\""},
		{"lax $[1 to 'z']", `[1, 2, 3]`, "array index must be a number, but found \"z\""},
		{"lax $['a' to 1]", `[1, 2, 3]`, "array index must be a number, but found \"a\""},
		{"strict $[0 to 100]", `[1, 2, 3]`, "array index out of bounds"},
		{"strict $[100]", `[1, 2, 3]`, "array index 100 out of bounds"},
		{"strict $[5 to 2]", `[1, 2, 3]`, "the end of a range can't come before the beginning"},
		{"strict $.foo", `{"bar":1}`, "object {\"bar\":1} missing `foo` field"},
		{"strict $.foo", `"wahoo"`, "cannot access field `foo` on non-object \"wahoo\""},
		{"strict $.foo", `[{"foo": 1}, {"foo": 2}]`, "cannot access field `foo` on non-object [{\"foo\":1},{\"foo\":2}]"},
		{"strict 'foo'.*", `{}`, "can't .* non-object \"foo\""},
		{"strict $[0]", `"hi"`, "can't index non-array \"hi\""},

		{"lax $[*] + 2", `[1, 2]`, "binary operators can only operate on single values"},
		{"strict -$.readings.floor()", `{"readings": [15.2, -22.3, 45.9] }`, ".floor() only defined on numbers"},

		// strict mode:
		// {"$[1 to 0]", `[1, 2, 3]`, "the end of a range can't come before the beginning"},
		{"lax -$[*]", `[1, "foo"]`, "unary minus can only accept numbers"},
		{"lax +$[*]", `[1, "foo"]`, "unary plus can only accept numbers"},
	}
	for _, tc := range testCases {
		t.Run(tc.input+"/"+tc.expectedError, func(t *testing.T) {
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
