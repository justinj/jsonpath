package jsonpath

import (
	"reflect"
	"testing"
)

func iterToArray(iter func() (jsonpathSym, bool)) []string {
	result := make([]string, 0)
	for n, ok := iter(); ok; n, ok = iter() {
		result = append(result, n.Lexeme())
	}
	return result
}

type testCase struct {
	input    string
	expected []string
}

func TestLex(t *testing.T) {
	testCases := []testCase{
		{"", []string{}},
		{"$", []string{"$"}},
		{".", []string{"."}},
		{"$.", []string{"$", "."}},
		{"@*$.", []string{"@", "*", "$", "."}},
		{"a", []string{"a"}},
		{"foo", []string{"foo"}},
		{"_fo_o", []string{"_fo_o"}},
		{"foo1", []string{"foo1"}},
		{"1foo", []string{"1", "foo"}},
		{"foo.bar", []string{"foo", ".", "bar"}},
		{"foo bar", []string{"foo", "bar"}},
		{"foo   .   bar", []string{"foo", ".", "bar"}},
		{"$.foo.bar", []string{"$", ".", "foo", ".", "bar"}},
		{"[foo]", []string{"[", "foo", "]"}},
		{"[()]", []string{"[", "(", ")", "]"}},
		{"1", []string{"1"}},
		{"1.1", []string{"1.1"}},
		{"123.123", []string{"123.123"}},
		{"1 < 2", []string{"1", "<", "2"}},
		{"1 <= 2", []string{"1", "<=", "2"}},
		{"1 > 2", []string{"1", ">", "2"}},
		{"1 >= 2", []string{"1", ">=", "2"}},
		{"1 && 2", []string{"1", "&&", "2"}},
		{"1 || 2", []string{"1", "||", "2"}},
		{"1 + 2", []string{"1", "+", "2"}},
		{"1 * 2", []string{"1", "*", "2"}},
		{"'hello world'", []string{"'hello world'"}},
		{"\"hello world\"", []string{"'hello world'"}},
		{"'hi \\'foo\\''", []string{"'hi 'foo''"}},
		{"'hi\\nthere'", []string{"'hi\nthere'"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			c := lex(tc.input)
			result := make([]string, 0)
			for elem := range c {
				result = append(result, elem.Lexeme())
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("expected %v, was %v", tc.expected, result)
			}
		})
	}
}
