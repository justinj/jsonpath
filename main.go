package jsonpath

func Parse(input string) (jsonPathNode, error) {
	yyErrorVerbose = true
	parser := yyNewParser()
	tok := tokens(input)
	parser.Parse(tok)

	if tok.err != nil {
		return nil, tok.err
	}

	validator := &validationVisitor{}
	tok.expr.Walk(validator)
	if validator.err != nil {
		return nil, validator.err
	}

	return tok.expr, nil
}
