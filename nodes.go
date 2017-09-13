package jsonpath

import (
	"bytes"
)

type visitor interface {
	VisitPre(jsonPathNode)
	VisitPost(jsonPathNode)
}

type jsonPathNode interface {
	Format(*bytes.Buffer)
	Walk(*visitor)
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
	left  jsonPathNode
	right jsonPathNode
}

type unaryExprType int

const (
	uminus unaryExprType = iota
	uplus
	unot
)

type UnaryExpr struct {
	t    unaryExprType
	expr jsonPathNode
}

type ParenExpr struct {
	val jsonPathNode
}

type NumberExpr struct {
	val float64
}

type VariableExpr struct {
	name string
}

type AtExpr struct{}
type LastExpr struct{}

type BoolExpr struct{ val bool }
type NullExpr struct{ val bool }
type StringExpr struct{ val string }

type AccessExpr struct {
	left  jsonPathNode
	right jsonPathNode
}

type DotAccessor struct {
	val    string
	quoted bool
}

type MemberWildcardAccessor struct{}

type ArrayAccessor struct {
	subscripts []jsonPathNode
}

type RangeNode struct {
	start jsonPathNode
	end   jsonPathNode
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
