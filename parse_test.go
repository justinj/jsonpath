package jsonpath

import "testing"

type parseTestCase struct {
	input    string
	expected string
}

func TestParse(t *testing.T) {
	testCases := []parseTestCase{
		{"1", "1"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			tok := tokens(tc.input)
			yyParse(tok)
		})
	}
}
