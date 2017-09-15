package jsonpath

import "fmt"

// This implementation of eval uses Go's builtin encoding/decoding of json.
type NaiveEvaler struct {
	program jsonPathExpr
}

type naiveEvalContext struct {
	dollar jsonValue
}

type jsonValue interface{}
type jsonSequence []jsonValue

func (n NaiveEvaler) Run(dollar jsonValue) (jsonSequence, error) {
	return n.program.naiveEval(&naiveEvalContext{
		dollar: dollar,
	})
}

func NewNaiveEvaler(program string) (*NaiveEvaler, error) {
	p, err := Parse(program)
	if err != nil {
		return nil, err
	}
	return &NaiveEvaler{
		program: p,
	}, nil
}

func (n BinExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	leftVal, err := n.left.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	left := leftVal[0]
	rightVal, err := n.right.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	right := rightVal[0]
	if l, ok := left.(float64); ok {
		if r, ok := right.(float64); ok {
			switch n.t {
			case plusBinOp:
				return jsonSequence{l + r}, nil
			case minusBinOp:
				return jsonSequence{l - r}, nil
			case timesBinOp:
				return jsonSequence{l * r}, nil
			case divBinOp:
				return jsonSequence{l / r}, nil
			case modBinOp:
				return jsonSequence{float64(int(l) % int(r))}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown op")
}

func (n NumberExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return []jsonValue{n.val}, nil
}

func (n UnaryExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n ParenExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n VariableExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	return jsonSequence{ctx.dollar}, nil
}

func (n LastExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n BoolExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n NullExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n StringExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n AccessExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	left, err := n.left.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	return n.right.naiveAccess(ctx, left)
}

func (n DotAccessor) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n DotAccessor) naiveAccess(ctx *naiveEvalContext, node jsonSequence) (jsonSequence, error) {
	result := make(jsonSequence, len(node), len(node))
	for i, e := range node {
		if obj, ok := e.(map[string]interface{}); ok {
			result[i] = obj[n.val]
		}
	}
	return result, nil
}

func (n MemberWildcardAccessor) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n MemberWildcardAccessor) naiveAccess(_ *naiveEvalContext, _ jsonSequence) (jsonSequence, error) {
	return nil, nil
}

func (n ArrayAccessor) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n ArrayAccessor) naiveAccess(ctx *naiveEvalContext, val jsonSequence) (jsonSequence, error) {
	// TODO: try to come up with a sensible estimate of the size of this.
	// if there is no reference to `last`, then we should know the exact size
	result := make(jsonSequence, 0, len(val))
	for _, e := range val {
		if ary, ok := e.([]interface{}); ok {
			for _, s := range n.subscripts {
				index, err := s.start.naiveEval(ctx)
				if err != nil {
					return nil, err
				}
				if len(index) != 1 {
					//TODO improve error message
					return nil, fmt.Errorf("indexes must return single value")
				}
				i := index[0]
				if idx, ok := i.(float64); ok {
					result = append(result, ary[int(idx)])
				} else {
					//TODO improve error message
					return nil, fmt.Errorf("index must be number")
				}
			}
		}
	}
	return result, nil
}

func (n RangeSubscriptNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n WildcardArrayAccessor) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n WildcardArrayAccessor) naiveAccess(_ *naiveEvalContext, _ jsonSequence) (jsonSequence, error) {
	return nil, nil
}

func (n FuncNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n FuncNode) naiveAccess(_ *naiveEvalContext, _ jsonSequence) (jsonSequence, error) {
	return nil, nil
}

func (n FilterNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n FilterNode) naiveAccess(_ *naiveEvalContext, _ jsonSequence) (jsonSequence, error) {
	return nil, nil
}

func (n ExistsNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n LikeRegexNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n StartsWithNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}

func (n IsUnknownNode) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return nil, nil
}
