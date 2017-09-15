package jsonpath

import (
	"bytes"
)

type jsonPathNode interface {
	Format(*bytes.Buffer)
	Walk(visitor)
}

type jsonPathExpr interface {
	Format(*bytes.Buffer)
	Walk(visitor)

	naiveEval(*naiveEvalContext) (jsonSequence, error)
}

type accessor interface {
	Format(*bytes.Buffer)
	Walk(visitor)

	naiveAccess(*naiveEvalContext, jsonSequence) (jsonSequence, error)
}

type binExprType int

const (
	plusBinOp binExprType = iota
	minusBinOp
	timesBinOp
	divBinOp
	modBinOp
	eqBinOp
	neqBinOp
	andBinOp
	orBinOp
	gtBinOp
	gteBinOp
	ltBinOp
	lteBinOp
)

type BinExpr struct {
	t     binExprType
	left  jsonPathExpr
	right jsonPathExpr
}

type unaryExprType int

const (
	uminus unaryExprType = iota
	uplus
	unot
)

type UnaryExpr struct {
	t    unaryExprType
	expr jsonPathExpr
}

type ParenExpr struct {
	expr jsonPathNode
}

type NumberExpr struct {
	val float64
}

type VariableExpr struct {
	name string
}

type LastExpr struct{}

type BoolExpr struct{ val bool }
type NullExpr struct{ val bool }
type StringExpr struct{ val string }

type AccessExpr struct {
	left  jsonPathExpr
	right accessor
}

type DotAccessor struct {
	val    string
	quoted bool
}

type MemberWildcardAccessor struct{}

type RangeSubscriptNode struct {
	start jsonPathExpr
	end   jsonPathExpr
}

type ArrayAccessor struct {
	subscripts []RangeSubscriptNode
}

type WildcardArrayAccessor struct{}

type function int

const (
	typeFunction function = iota
	sizeFunction
	doubleFunction
	ceilingFunction
	floorFunction
	absFunction
	datetimeFunction
	keyvalueFunction
)

type FuncNode struct {
	f   function
	arg jsonPathNode
}

type FilterNode struct {
	pred jsonPathNode
}

type ExistsNode struct {
	expr jsonPathNode
}

type LikeRegexNode struct {
	left    jsonPathNode
	pattern string
	flag    *string
}

type StartsWithNode struct {
	left  jsonPathNode
	right jsonPathNode
}

type IsUnknownNode struct {
	expr jsonPathNode
}
