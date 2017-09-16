package jsonpath

import (
	"bytes"
	"regexp"
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

type sqlJsonBool int

const (
	sqlJsonFalse sqlJsonBool = iota
	sqlJsonTrue
	sqlJsonUnknown
)

type jsonPathPred interface {
	Format(*bytes.Buffer)
	Walk(visitor)

	naivePredEval(*naiveEvalContext) (sqlJsonBool, error)
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
)

type BinExpr struct {
	t     binExprType
	left  jsonPathExpr
	right jsonPathExpr
}

type binPredType int

const (
	eqBinOp binPredType = iota
	neqBinOp
	gtBinOp
	gteBinOp
	ltBinOp
	lteBinOp
)

type BinPred struct {
	t     binPredType
	left  jsonPathExpr
	right jsonPathExpr
}

type binLogicType int

const (
	andBinOp binLogicType = iota
	orBinOp
)

type BinLogic struct {
	t     binLogicType
	left  jsonPathPred
	right jsonPathPred
}

type unaryExprType int

const (
	uminus unaryExprType = iota
	uplus
)

type UnaryExpr struct {
	t    unaryExprType
	expr jsonPathExpr
}

type UnaryNot struct {
	expr jsonPathPred
}

type ParenExpr struct {
	expr jsonPathExpr
}

type ParenPred struct {
	expr jsonPathPred
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
	pred jsonPathPred
}

type ExistsNode struct {
	expr jsonPathExpr
}

type LikeRegexNode struct {
	left       jsonPathExpr
	rawPattern string
	pattern    *regexp.Regexp
	flag       *string
}

type StartsWithNode struct {
	left  jsonPathExpr
	right jsonPathExpr
}

type IsUnknownNode struct {
	expr jsonPathPred
}
