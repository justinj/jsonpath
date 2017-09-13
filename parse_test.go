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
		{"1", Number{val: 1}},
		{"1+1*1",
			BinExpr{
				t:     plusBinOp,
				left:  Number{val: 1},
				right: BinExpr{t: timesBinOp, left: Number{val: 1}, right: Number{val: 1}},
			}},
		{"1*1+1",
			BinExpr{
				t:     plusBinOp,
				left:  BinExpr{t: timesBinOp, left: Number{val: 1}, right: Number{val: 1}},
				right: Number{val: 1},
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
		"@",
		"last",

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
		"$ ? (\"foo\" is unknown)",

		"$ ? (!\"foo\" is unknown)",
		// ^^^ should this be allowed?
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			parser := yyNewParser()
			tok := tokens(tc)
			parser.Parse(tok)

			if FormatNode(tok.expr) != tc {
				t.Errorf("expected `%s`, got `%s`", tc, FormatNode(tok.expr))
			}
		})
	}
}
