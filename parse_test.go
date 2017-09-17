package jsonpath

import (
	"reflect"
	"testing"
)

type parseTestCase struct {
	input  string
	result jsonPathNode
}

func TestParseComplete(t *testing.T) {
	testCases := []parseTestCase{
		{"lax 1", Program{root: NumberExpr{val: 1}, mode: modeLax}},
		{"lax 1+1*1",
			Program{
				mode: modeLax,
				root: BinExpr{
					t:     plusBinOp,
					left:  NumberExpr{val: 1},
					right: BinExpr{t: timesBinOp, left: NumberExpr{val: 1}, right: NumberExpr{val: 1}},
				}}},
		{"lax 1*1+1",
			Program{
				mode: modeLax,
				root: BinExpr{
					t:     plusBinOp,
					left:  BinExpr{t: timesBinOp, left: NumberExpr{val: 1}, right: NumberExpr{val: 1}},
					right: NumberExpr{val: 1},
				}}},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			parser := yyNewParser()
			tok := tokens(tc.input)
			parser.Parse(tok)

			if !reflect.DeepEqual(tok.root, tc.result) {
				t.Errorf("expected `%#v`, got `%#v`", tc.result, tok.root)
			}
		})
	}
}

func TestParse(t *testing.T) {
	testCases := []string{
		"lax 1",
		"lax 2",
		"lax 1 + 1",
		"lax 1 + 2 + 3",
		"lax 1 - 1",
		"lax 1 - (1 + 2)",
		"lax 1 - ((((1 + 2))))",
		"lax 1 * 2 / 3 * 4",
		"lax 1 + -1",
		"lax 1 + +1",
		"lax 5 % 2",
		"lax true",
		"lax false",
		"lax null",
		"lax \"hello\"",
		"lax \"he\\\"llo\"",

		"lax $a + 1",
		"lax $foobar",
		"lax $",
		"lax $[last]",

		"lax $.foo",
		"lax $.\"$foo\"",
		"lax $.\"$f\\\"oo\"",
		"lax $.\"foo bar\"",
		"lax $.foo.bar",
		"lax $.*",
		"lax $[1]",
		"lax $[1, 2]",
		"lax $[1, 2 to 4]",
		"lax $[last]",
		"lax $[0, last - 1 to last, 5]",
		"lax $[\"hello\"]",
		"lax $[*]",
		"lax $.type()",
		"lax $.size()",
		"lax $.double()",
		"lax $.ceiling()",
		"lax $.floor()",
		"lax $.abs()",
		"lax $.datetime(\"foobar\")",
		"lax $.keyvalue()",

		"lax $ ? (exists (@.foobar))",
		"lax $ ? (1 == 1)",
		"lax $ ? (1 > 1)",
		"lax $ ? (1 < 1)",
		"lax $ ? (1 <= 1)",
		"lax $ ? (1 >= 1)",
		"lax $ ? (1 != 1)",
		// "$ ? (1 <> 1)", <- need Parse2 to test this
		"lax $ ? (1 != 1 && 1 == 1)",
		"lax $ ? (1 != 1 || 1 == 1)",
		"lax $ ? ((1 != 1) || 1 == 1)",
		"lax $ ? ((1 == 1))",
		"lax $ ? (!(1 == 1))",

		"lax $ ? (\"foo\" like_regex \"bar\")",
		"lax $ ? (\"foo\" like_regex \"bar\" flag \"i\")",
		"lax $ ? (\"foo\" starts with \"fo\")",
		"lax $ ? ((1 == 1) is unknown)",

		"lax $ ? (!(1 == 1) is unknown)",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			res, err := Parse(tc)
			if err != nil {
				t.Fatal(err)
			}
			if FormatNode(res) != tc {
				t.Errorf("expected `%s`, got `%s`", tc, FormatNode(res))
			}
		})
	}
}

func TestParseError(t *testing.T) {
	testCases := []struct {
		input  string
		errMsg string
	}{
		{"lax (", "syntax error: unexpected $end"},
		{"lax @.foo", "@ only allowed within filter expressions"},
		{"lax $ ? ((@.foo == 1) is unknown)[*] + @.foo", "@ only allowed within filter expressions"},
		{"lax @.foo + $ ? ((@.foo == 1) is unknown)[*]", "@ only allowed within filter expressions"},
		{"lax $ ? (@.foo)", "filter expressions cannot be raw json values - if you expect `@.foo` to be boolean true, write `@.foo == true`"},
		{"lax last", "`last` can only appear inside an array subscript"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, err := Parse(tc.input)
			if err == nil {
				t.Fatalf("expected \"%s\" to error with \"%s\", but no error occurred", tc.input, tc.errMsg)
			}
			if err.Error() != tc.errMsg {
				t.Fatalf("expected \"%s\" to error with \"%s\", but error was \"%s\"", tc.input, tc.errMsg, err.Error())
			}
		})
	}
}
