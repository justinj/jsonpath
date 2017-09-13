package jsonpath

import "fmt"

func Parse(input string) (jsonPathNode, error) {
	yyErrorVerbose = true
	parser := yyNewParser()
	tok := tokens(input)
	parser.Parse(tok)
	if tok.seenAt {
		return nil, fmt.Errorf("@ only allowed within filter expressions")
	}
	if tok.err != nil {
		return nil, tok.err
	}
	return tok.expr, nil
}
