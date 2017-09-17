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

func TestLex(t *testing.T) {
	testCases := []struct {
		input          string
		expected       []string
		expectedIdents []int
	}{
		{"", []string{}, []int{}},
		{"$", []string{"$"}, []int{IDENT}},
		{".", []string{"."}, []int{'.'}},
		{"$.", []string{"$", "."}, []int{IDENT, '.'}},
		{"@*$.", []string{"@", "*", "$", "."}, []int{'@', '*', IDENT, '.'}},
		{"$foo", []string{"$foo"}, []int{IDENT}},
		{"$foo.bar", []string{"$foo", ".", "bar"}, []int{IDENT, '.', IDENT}},
		{"$foo   .   bar", []string{"$foo", ".", "bar"}, []int{IDENT, '.', IDENT}},
		{"$.foo.bar", []string{"$", ".", "foo", ".", "bar"}, []int{IDENT, '.', IDENT, '.', IDENT}},
		{"[$foo]", []string{"[", "$foo", "]"}, []int{'[', IDENT, ']'}},
		{"[()]", []string{"[", "(", ")", "]"}, []int{'[', '(', ')', ']'}},
		{"1", []string{"1"}, []int{NUMBER}},
		{"1.1", []string{"1.1"}, []int{NUMBER}},
		{"123.123", []string{"123.123"}, []int{NUMBER}},
		{"12.3e0", []string{"12.3"}, []int{NUMBER}},
		{"true false null", []string{"true", "false", "null"}, []int{TRUE, FALSE, NULL}},
		{"1 == 2", []string{"1", "==", "2"}, []int{NUMBER, EQ, NUMBER}},
		{"1 < 2", []string{"1", "<", "2"}, []int{NUMBER, '<', NUMBER}},
		{"1 <= 2", []string{"1", "<=", "2"}, []int{NUMBER, LTE, NUMBER}},
		{"1 > 2", []string{"1", ">", "2"}, []int{NUMBER, '>', NUMBER}},
		{"1 >= 2", []string{"1", ">=", "2"}, []int{NUMBER, GTE, NUMBER}},
		{"1 && 2", []string{"1", "&&", "2"}, []int{NUMBER, AND, NUMBER}},
		{"1 || 2", []string{"1", "||", "2"}, []int{NUMBER, OR, NUMBER}},
		{"1 + 2", []string{"1", "+", "2"}, []int{NUMBER, '+', NUMBER}},
		{"1 * 2", []string{"1", "*", "2"}, []int{NUMBER, '*', NUMBER}},
		{"1 / 2", []string{"1", "/", "2"}, []int{NUMBER, '/', NUMBER}},
		{"1 % 2", []string{"1", "%", "2"}, []int{NUMBER, '%', NUMBER}},
		{"''", []string{"''"}, []int{STR}},
		{"'hello world'", []string{"'hello world'"}, []int{STR}},
		{"\"hello world\"", []string{"'hello world'"}, []int{STR}},
		{"'hi \\'foo\\''", []string{"'hi 'foo''"}, []int{STR}},
		{"'hi\\nthere'", []string{"'hi\nthere'"}, []int{STR}},
		{"strict lax", []string{"strict", "lax"}, []int{STRICT, LAX}},
		{"to", []string{"to"}, []int{TO}},
		{"[1, 2]", []string{"[", "1", ",", "2", "]"}, []int{'[', NUMBER, ',', NUMBER, ']'}},

		{"$.type()", []string{"$", ".", "type", "(", ")"}, []int{IDENT, '.', FUNC_TYPE, '(', ')'}},
		{"$.type    ()", []string{"$", ".", "type", "(", ")"}, []int{IDENT, '.', FUNC_TYPE, '(', ')'}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, c := lex(tc.input)
			result := make([]string, 0)
			resultIdents := make([]int, 0)

			for elem := range c {
				result = append(result, elem.Lexeme())
				resultIdents = append(resultIdents, elem.identifier())
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("expected %v, was %v", tc.expected, result)
			}

			if !reflect.DeepEqual(resultIdents, tc.expectedIdents) {
				t.Fatalf("expected %v, was %v", tc.expectedIdents, resultIdents)
			}
		})
	}
}

func TestLexErrors(t *testing.T) {
	testCases := []struct {
		input         string
		expectedError string
		expectedBegin int
		expectedEnd   int
	}{
		{"1 = 1", "use == instead of =", 2, 2},
		{"1 | 1", "| must be followed by |", 2, 2},
		{"foo", "unrecognized keyword \"foo\"", 0, 2},
		{"\"hello", "unterminated string", 0, 5},
		{"\"\\y\"", "invalid escape sequence \"\\y\"", 2, 2},
		{"$.bar()", "invalid function \"bar\"", 2, 4},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, c := lex(tc.input)

			for elem := range c {
				if err, ok := elem.(errSym); ok {
					if err.msg != tc.expectedError {
						t.Fatalf("expected \"%s\", got \"%s\"", tc.expectedError, err.msg)
					}
					if err.begin != tc.expectedBegin && err.end != tc.expectedEnd {
						t.Fatalf(
							"expected error to encompass [%d, %d], but was [%d, %d]",
							tc.expectedBegin, tc.expectedEnd,
							err.begin, err.end,
						)
					}
					return
				}
			}
			t.Fatalf("expected \"%s\" to error, but no error occurred", tc.input)
		})
	}
}
