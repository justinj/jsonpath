package jsonpath

import "testing"

func TestEval(t *testing.T) {
	testCases := []struct {
		input  string
		result jsonPathNode
	}{}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
		})
	}
}
