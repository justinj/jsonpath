package jsonpath

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// This implementation of eval uses Go's builtin encoding/decoding of json.
type NaiveEvaler struct {
	program jsonPathExpr
}

type evalMode int

const (
	laxMode evalMode = iota
	strictMode
)

type naiveEvalContext struct {
	dollar                 jsonValue
	containingArrayLengths []float64
	atSigns                []jsonValue
	mode                   evalMode
}

type jsonValue interface{}
type jsonSequence []jsonValue

func (n NaiveEvaler) Run(dollar jsonValue) (jsonSequence, error) {
	return n.program.naiveEval(&naiveEvalContext{
		dollar:                 dollar,
		containingArrayLengths: make([]float64, 0, 10),
		mode: laxMode,
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

func comparable(x interface{}, y interface{}) bool {
	if _, ok := x.(map[string]interface{}); ok {
		return false
	}
	if _, ok := y.(map[string]interface{}); ok {
		return false
	}
	if _, ok := x.([]interface{}); ok {
		return false
	}
	if _, ok := y.([]interface{}); ok {
		return false
	}
	if x == nil || y == nil {
		return true
	}
	switch x.(type) {
	case string:
		_, ok := y.(string)
		return ok
	case float64:
		_, ok := y.(float64)
		return ok
	case bool:
		_, ok := y.(bool)
		return ok
	}
	return false
}

func (n BinPred) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	leftVal, err := n.left.naiveEval(ctx)
	if err != nil {
		return sqlJsonUnknown, nil
	}
	rightVal, err := n.right.naiveEval(ctx)
	if err != nil {
		return sqlJsonUnknown, nil
	}
	//TODO strict vs. lax semantics are different here
	switch n.t {
	case eqBinOp:
		for _, l := range leftVal {
			for _, r := range rightVal {
				if !comparable(l, r) {
					return sqlJsonUnknown, nil
				}
				if l == r {
					return sqlJsonTrue, nil
				}
			}
		}
		return sqlJsonFalse, nil
	}
	return 0, fmt.Errorf("unknown op")
}

func (n BinLogic) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	left, err := n.left.naivePredEval(ctx)
	if err != nil {
		return 0, err
	}
	right, err := n.right.naivePredEval(ctx)
	if err != nil {
		return 0, err
	}
	switch n.t {
	case orBinOp:
		if left == sqlJsonTrue || right == sqlJsonTrue {
			return sqlJsonTrue, nil
		}
		return sqlJsonFalse, nil
	case andBinOp:
		if left == sqlJsonTrue && right == sqlJsonTrue {
			return sqlJsonTrue, nil
		}
		return sqlJsonFalse, nil
	}
	return 0, fmt.Errorf("unknown op")
}

func (n BinExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	leftVal, err := n.left.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	if len(leftVal) != 1 {
		return nil, fmt.Errorf("binary operators can only operate on single values")
	}
	left := leftVal[0]
	rightVal, err := n.right.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	if len(rightVal) != 1 {
		return nil, fmt.Errorf("binary operators can only operate on single values")
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
	return jsonSequence{n.val}, nil
}

func (n UnaryExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	expr, err := n.expr.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	switch n.t {
	case uminus:
		result := make(jsonSequence, len(expr))
		for i, e := range expr {
			if num, ok := e.(float64); ok {
				result[i] = -num
			} else {
				return nil, fmt.Errorf("unary minus can only accept numbers")
			}
		}
		return result, nil
	case uplus:
		for _, e := range expr {
			if _, ok := e.(float64); !ok {
				return nil, fmt.Errorf("unary plus can only accept numbers")
			}
		}
		return expr, nil
	}
	return nil, fmt.Errorf("unknown unary op")
}

func (n UnaryNot) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	expr, err := n.expr.naivePredEval(ctx)
	if err != nil {
		return 0, err
	}
	if expr == sqlJsonTrue {
		return sqlJsonFalse, nil
	}
	if expr == sqlJsonFalse {
		return sqlJsonTrue, nil
	}
	return sqlJsonUnknown, nil
}

func (n ParenPred) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	return n.expr.naivePredEval(ctx)
}

func (n ParenExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	return n.expr.naiveEval(ctx)
}

func (n VariableExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	switch n.name {
	case "$":
		return jsonSequence{ctx.dollar}, nil
	case "@":
		return jsonSequence{ctx.atSigns[len(ctx.atSigns)-1]}, nil
	}
	return nil, fmt.Errorf(":(")
}

func (n LastExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	return jsonSequence{ctx.containingArrayLengths[len(ctx.containingArrayLengths)-1]}, nil
}

func (n BoolExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return jsonSequence{n.val}, nil
}

func (n NullExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return jsonSequence{nil}, nil
}

func (n StringExpr) naiveEval(_ *naiveEvalContext) (jsonSequence, error) {
	return jsonSequence{n.val}, nil
}

func (n AccessExpr) naiveEval(ctx *naiveEvalContext) (jsonSequence, error) {
	left, err := n.left.naiveEval(ctx)
	if err != nil {
		return nil, err
	}
	return n.right.naiveAccess(ctx, left)
}

func (n DotAccessor) naiveAccess(ctx *naiveEvalContext, node jsonSequence) (jsonSequence, error) {
	result := make(jsonSequence, 0, len(node))
	for _, e := range node {
		if obj, ok := e.(map[string]interface{}); ok {
			if v, ok := obj[n.val]; ok {
				result = append(result, v)
			}
		}
	}
	return result, nil
}

func (n MemberWildcardAccessor) naiveAccess(ctx *naiveEvalContext, s jsonSequence) (jsonSequence, error) {
	// TODO: try to estimate size...
	result := make(jsonSequence, 0)
	for _, e := range s {
		if obj, ok := e.(map[string]interface{}); ok {
			for _, v := range obj {
				result = append(result, v)
			}
		} else {
			return nil, fmt.Errorf("arguments to `.*` must be objects")
		}
	}
	return result, nil
}

func (n ArrayAccessor) naiveAccess(ctx *naiveEvalContext, val jsonSequence) (jsonSequence, error) {
	// TODO: try to come up with a sensible estimate of the size of this.
	// if there is no reference to `last`, then we should know the exact size
	result := make(jsonSequence, 0, len(val))
	ctx.containingArrayLengths = append(ctx.containingArrayLengths, 0)
	for _, e := range val {
		if ary, ok := e.([]interface{}); ok {
			ctx.containingArrayLengths[len(ctx.containingArrayLengths)-1] = float64(len(ary) - 1)
			for _, s := range n.subscripts {
				start, err := s.start.naiveEval(ctx)
				if err != nil {
					return nil, err
				}
				if len(start) != 1 {
					//TODO improve error message
					return nil, fmt.Errorf("indexes must return single value")
				}
				i := start[0]
				if idx, ok := i.(float64); ok {
					if s.end == nil {
						if int(idx) < 0 || int(idx) >= len(ary) {
							return nil, fmt.Errorf("array index %d out of bounds", int(idx))
						}
						result = append(result, ary[int(idx)])
					} else {
						end, err := s.end.naiveEval(ctx)
						if err != nil {
							return nil, err
						}
						if len(end) != 1 {
							return nil, fmt.Errorf("indexes must return single value")
						}
						j := end[0]
						if idxEnd, ok := j.(float64); ok {
							if idxEnd < idx && ctx.mode == strictMode {
								return nil, fmt.Errorf("the end of a range can't come before the beginning")
							}
							for i := idx; i <= idxEnd; i++ {
								if int(i) < 0 || int(i) >= len(ary) {
									return nil, fmt.Errorf("array index out of bounds")
								}
								result = append(result, ary[int(i)])
							}
						} else {
							return nil, fmt.Errorf("array index must be a number, but found %#v", j)
						}
					}
				} else {
					//TODO improve error message
					return nil, fmt.Errorf("array index must be a number, but found %#v", i)
				}
			}
		}
	}
	ctx.containingArrayLengths = ctx.containingArrayLengths[:len(ctx.containingArrayLengths)-1]
	return result, nil
}

func (n WildcardArrayAccessor) naiveAccess(ctx *naiveEvalContext, val jsonSequence) (jsonSequence, error) {
	// TODO: handle lax vs. strict mode here
	result := make(jsonSequence, 0, len(val))
	for _, e := range val {
		if ary, ok := e.([]interface{}); ok {
			for _, elem := range ary {
				result = append(result, elem)
			}
		} else {
			// TODO: this is lax mode semantics, strict would error here
			result = append(result, e)
		}
	}
	return result, nil
}

func (n FuncNode) naiveAccess(_ *naiveEvalContext, val jsonSequence) (jsonSequence, error) {
	switch n.f {
	case typeFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			switch e.(type) {
			case nil:
				result[i] = "null"
			case bool:
				result[i] = "boolean"
			case float64:
				result[i] = "number"
			case string:
				result[i] = "string"
			case []interface{}:
				result[i] = "array"
			case map[string]interface{}:
				result[i] = "object"
			default:
				return nil, fmt.Errorf("unknown elem type %T", e)
			}
		}
		return result, nil
	case sizeFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			if ary, ok := e.([]interface{}); ok {
				result[i] = len(ary)
			} else {
				result[i] = 1
			}
		}
		return result, nil
	case doubleFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			switch t := e.(type) {
			case float64:
				result[i] = t
			case string:
				n, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				result[i] = n
			default:
				return nil, fmt.Errorf(".double() only defined on strings and numbers")
			}
		}
		return result, nil
	case ceilingFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			if num, ok := e.(float64); ok {
				result[i] = math.Ceil(num)
			} else {
				return nil, fmt.Errorf(".ceiling() only defined on numbers")
			}
		}
		return result, nil
	case floorFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			if num, ok := e.(float64); ok {
				result[i] = math.Floor(num)
			} else {
				return nil, fmt.Errorf(".floor() only defined on numbers")
			}
		}
		return result, nil
	case absFunction:
		result := make(jsonSequence, len(val))
		for i, e := range val {
			if num, ok := e.(float64); ok {
				result[i] = math.Abs(num)
			} else {
				return nil, fmt.Errorf(".abs() only defined on numbers")
			}
		}
		return result, nil
	case keyvalueFunction:
		result := make(jsonSequence, 0)
		for i, e := range val {
			if obj, ok := e.(map[string]interface{}); ok {
				for k, v := range obj {
					result = append(result, map[string]interface{}{
						"name":  k,
						"value": v,
						"id":    i,
					})
				}
			} else {
				// TODO: lax mode unwraps arrays
				return nil, fmt.Errorf(".keyvalue() only on objects")
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("unimplemented function")
}

func (n FilterNode) naiveAccess(ctx *naiveEvalContext, val jsonSequence) (jsonSequence, error) {
	result := make(jsonSequence, 0, len(val))
	for _, e := range val {
		ctx.atSigns = append(ctx.atSigns, e)
		pass, err := n.pred.naivePredEval(ctx)
		if err != nil {
			return nil, err
		}
		if pass == sqlJsonTrue {
			result = append(result, e)
		}
		ctx.atSigns = ctx.atSigns[:len(ctx.atSigns)-1]
	}
	return result, nil
}

func (n ExistsNode) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	e, err := n.expr.naiveEval(ctx)
	if err != nil {
		return sqlJsonUnknown, nil
	}
	if len(e) > 0 {
		return sqlJsonTrue, nil
	}
	return sqlJsonFalse, nil
}

func (n LikeRegexNode) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	exprs, err := n.left.naiveEval(ctx)
	if err != nil {
		return 0, err
	}
	for _, e := range exprs {
		if s, ok := e.(string); ok {
			if n.pattern.Match([]byte(s)) {
				return sqlJsonTrue, nil
			}
		}
	}
	return sqlJsonFalse, nil
}

func (n StartsWithNode) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	left, err := n.left.naiveEval(ctx)
	if err != nil {
		return 0, err
	}
	right, err := n.right.naiveEval(ctx)
	if err != nil {
		return 0, err
	}
	for _, l := range left {
		for _, r := range right {
			if sl, ok := l.(string); ok {
				if sr, ok := r.(string); ok {
					if strings.HasPrefix(sl, sr) {
						return sqlJsonTrue, nil
					}
				} else {
					return sqlJsonUnknown, nil
				}
			} else {
				return sqlJsonUnknown, nil
			}
		}
	}
	return sqlJsonFalse, nil
}

func (n IsUnknownNode) naivePredEval(ctx *naiveEvalContext) (sqlJsonBool, error) {
	e, err := n.expr.naivePredEval(ctx)
	if err != nil {
		return 0, err
	}
	if e == sqlJsonUnknown {
		return sqlJsonTrue, nil
	}
	return sqlJsonFalse, nil
}
