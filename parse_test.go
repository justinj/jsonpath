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
		{"1", NumberExpr{val: 1}},
		{"1+1*1",
			BinExpr{
				t:     plusBinOp,
				left:  NumberExpr{val: 1},
				right: BinExpr{t: timesBinOp, left: NumberExpr{val: 1}, right: NumberExpr{val: 1}},
			}},
		{"1*1+1",
			BinExpr{
				t:     plusBinOp,
				left:  BinExpr{t: timesBinOp, left: NumberExpr{val: 1}, right: NumberExpr{val: 1}},
				right: NumberExpr{val: 1},
			}},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			parser := yyNewParser()
			tok := tokens(tc.input)
			parser.Parse(tok)

			if !reflect.DeepEqual(tok.expr, tc.result) {
				t.Errorf("expected `%#v`, got `%#v`", tc.result, tok.expr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	testCases := []string{
		"1",
		"2",
		"1 + 1",
		"1 + 2 + 3",
		"1 - 1",
		"1 - (1 + 2)",
		"1 - ((((1 + 2))))",
		"1 * 2 / 3 * 4",
		"1 + -1",
		"1 + +1",
		"5 % 2",
		"true",
		"false",
		"null",
		"\"hello\"",
		"\"he\\\"llo\"",

		"$a + 1",
		"$foobar",
		"$",
		"$[last]",

		"$.foo",
		"$.\"$foo\"",
		"$.\"$f\\\"oo\"",
		"$.\"foo bar\"",
		"$.foo.bar",
		"$.*",
		"$[1]",
		"$[1, 2]",
		"$[1, 2 to 4]",
		"$[last]",
		"$[0, last - 1 to last, 5]",
		"$[\"hello\"]",
		"$[*]",
		"$.type()",
		"$.size()",
		"$.double()",
		"$.ceiling()",
		"$.floor()",
		"$.abs()",
		"$.datetime(\"foobar\")",
		"$.keyvalue()",

		"$ ? (exists (@.foobar))",
		"$ ? (1 == 1)",
		"$ ? (1 > 1)",
		"$ ? (1 < 1)",
		"$ ? (1 <= 1)",
		"$ ? (1 >= 1)",
		"$ ? (1 != 1)",
		// "$ ? (1 <> 1)", <- need Parse2 to test this
		"$ ? (1 != 1 && 1 == 1)",
		"$ ? (1 != 1 || 1 == 1)",
		"$ ? ((1 != 1) || 1 == 1)",
		"$ ? ((1 == 1))",
		"$ ? (!(1 == 1))",

		"$ ? (\"foo\" like_regex \"bar\")",
		"$ ? (\"foo\" like_regex \"bar\" flag \"i\")",
		"$ ? (\"foo\" starts with \"fo\")",
		"$ ? ((1 == 1) is unknown)",

		"$ ? (!(1 == 1) is unknown)",
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
		{"(", "syntax error: unexpected $end"},
		{"@.foo", "@ only allowed within filter expressions"},
		{"$ ? ((@.foo == 1) is unknown)[*] + @.foo", "@ only allowed within filter expressions"},
		{"@.foo + $ ? ((@.foo == 1) is unknown)[*]", "@ only allowed within filter expressions"},
		{"$ ? (@.foo)", "filter expressions cannot be raw json values - if you expect `@.foo` to be boolean true, write `@.foo == true`"},
		{"last", "`last` can only appear inside an array subscript"},
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
