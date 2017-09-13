package jsonpath

import (
	"bytes"
	"fmt"
)

type jsonPathNode interface {
	Format(*bytes.Buffer)
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

type Paren struct {
	val jsonPathNode
}

type Number struct {
	val float64
}

type Variable struct {
	name string
}

type AtSign struct{}
type Last struct{}

type BoolNode struct{ val bool }
type NullNode struct{ val bool }
type StringNode struct{ val string }

type AccessNode struct {
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
	typeFunction = iota
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

func FormatNode(n jsonPathNode) string {
	b := bytes.NewBuffer(nil)
	n.Format(b)
	return b.String()
}

func (s Number) Format(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf("%v", s.val))
}

func (s BinExpr) Format(b *bytes.Buffer) {
	s.left.Format(b)
	b.WriteByte(' ')
	switch s.t {
	case plusBinOp:
		b.WriteByte('+')
	case minusBinOp:
		b.WriteByte('-')
	case timesBinOp:
		b.WriteByte('*')
	case divBinOp:
		b.WriteByte('/')
	case modBinOp:
		b.WriteByte('%')
	case eqBinOp:
		b.WriteString("==")
	case neqBinOp:
		b.WriteString("!=")
	case ltBinOp:
		b.WriteByte('<')
	case andBinOp:
		b.WriteString("&&")
	case orBinOp:
		b.WriteString("||")
	case gtBinOp:
		b.WriteByte('>')
	case lteBinOp:
		b.WriteString("<=")
	case gteBinOp:
		b.WriteString(">=")
	}
	b.WriteByte(' ')
	s.right.Format(b)
}

func (s UnaryExpr) Format(b *bytes.Buffer) {
	switch s.t {
	case uminus:
		b.WriteByte('-')
	case uplus:
		b.WriteByte('+')
	case unot:
		b.WriteByte('!')
	}
	s.expr.Format(b)
}

func (s Paren) Format(b *bytes.Buffer) {
	b.WriteByte('(')
	s.val.Format(b)
	b.WriteByte(')')
}

func (s Variable) Format(b *bytes.Buffer) {
	b.WriteString(s.name)
}

func (s AtSign) Format(b *bytes.Buffer) {
	b.WriteByte('@')
}

func (s Last) Format(b *bytes.Buffer) {
	b.WriteString("last")
}

func (s BoolNode) Format(b *bytes.Buffer) {
	if s.val {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
}

func (s NullNode) Format(b *bytes.Buffer) {
	b.WriteString("null")
}

func (s StringNode) Format(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf("%#v", s.val))
}

func (s AccessNode) Format(b *bytes.Buffer) {
	s.left.Format(b)
	s.right.Format(b)
}

func (s DotAccessor) Format(b *bytes.Buffer) {
	b.WriteByte('.')
	if s.quoted {
		b.WriteString(fmt.Sprintf("%#v", s.val))
	} else {
		b.WriteString(s.val)
	}
}

func (s MemberWildcardAccessor) Format(b *bytes.Buffer) {
	b.WriteString(".*")
}

func (s ArrayAccessor) Format(b *bytes.Buffer) {
	b.WriteByte('[')
	for i, elem := range s.subscripts {
		if i != 0 {
			b.WriteString(", ")
		}
		elem.Format(b)
	}
	b.WriteByte(']')
}

func (s RangeNode) Format(b *bytes.Buffer) {
	s.start.Format(b)
	b.WriteString(" to ")
	s.end.Format(b)
}

func (s WildcardArrayAccessor) Format(b *bytes.Buffer) {
	b.WriteString("[*]")
}

func (s FuncNode) Format(b *bytes.Buffer) {
	switch s.f {
	case typeFunction:
		b.WriteString(".type()")
	case sizeFunction:
		b.WriteString(".size()")
	case doubleFunction:
		b.WriteString(".double()")
	case ceilingFunction:
		b.WriteString(".ceiling()")
	case floorFunction:
		b.WriteString(".floor()")
	case absFunction:
		b.WriteString(".abs()")
	case datetimeFunction:
		b.WriteString(".datetime(")
		s.arg.Format(b)
		b.WriteByte(')')
	case keyvalueFunction:
		b.WriteString(".keyvalue()")
	}
}

func (s FilterNode) Format(b *bytes.Buffer) {
	b.WriteString(" ? (")
	s.pred.Format(b)
	b.WriteByte(')')
}

func (s ExistsNode) Format(b *bytes.Buffer) {
	b.WriteString("exists (")
	s.expr.Format(b)
	b.WriteByte(')')
}

func (s LikeRegexNode) Format(b *bytes.Buffer) {
	s.left.Format(b)
	b.WriteString(" like_regex ")
	b.WriteString(fmt.Sprintf("%#v", s.pattern))
	if s.flag != nil {
		b.WriteString(" flag ")
		b.WriteString(fmt.Sprintf("%#v", *s.flag))
	}
}

func (s StartsWithNode) Format(b *bytes.Buffer) {
	s.left.Format(b)
	b.WriteString(" starts with ")
	s.right.Format(b)
}

func (s IsUnknownNode) Format(b *bytes.Buffer) {
	s.expr.Format(b)
	b.WriteString(" is unknown")
}
