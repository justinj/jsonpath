package jsonpath

import (
	"bytes"
	"fmt"
)

func FormatNode(n jsonPathNode) string {
	b := bytes.NewBuffer(nil)
	n.Format(b)
	return b.String()
}

func (s Program) Format(b *bytes.Buffer) {
	if s.mode == modeLax {
		b.WriteString("lax ")
	}
	if s.mode == modeStrict {
		b.WriteString("strict ")
	}
	s.root.Format(b)
}

func (s NumberExpr) Format(b *bytes.Buffer) {
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
	}
	b.WriteByte(' ')
	s.right.Format(b)
}

func (s BinPred) Format(b *bytes.Buffer) {
	s.left.Format(b)
	b.WriteByte(' ')
	switch s.t {
	case eqBinOp:
		b.WriteString("==")
	case neqBinOp:
		b.WriteString("!=")
	case ltBinOp:
		b.WriteByte('<')
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

func (s BinLogic) Format(b *bytes.Buffer) {
	s.left.Format(b)
	b.WriteByte(' ')
	switch s.t {
	case orBinOp:
		b.WriteString("||")
	case andBinOp:
		b.WriteString("&&")
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
	}
	s.expr.Format(b)
}

func (s UnaryNot) Format(b *bytes.Buffer) {
	b.WriteByte('!')
	s.expr.Format(b)
}

func (s ParenPred) Format(b *bytes.Buffer) {
	b.WriteByte('(')
	s.expr.Format(b)
	b.WriteByte(')')
}

func (s ParenExpr) Format(b *bytes.Buffer) {
	b.WriteByte('(')
	s.expr.Format(b)
	b.WriteByte(')')
}

func (s VariableExpr) Format(b *bytes.Buffer) {
	b.WriteString(s.name)
}

func (s LastExpr) Format(b *bytes.Buffer) {
	b.WriteString("last")
}

func (s BoolExpr) Format(b *bytes.Buffer) {
	if s.val {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
}

func (s NullExpr) Format(b *bytes.Buffer) {
	b.WriteString("null")
}

func (s StringExpr) Format(b *bytes.Buffer) {
	b.WriteString(fmt.Sprintf("%#v", s.val))
}

func (s AccessExpr) Format(b *bytes.Buffer) {
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

func (s RangeSubscriptNode) Format(b *bytes.Buffer) {
	s.start.Format(b)
	if s.end != nil {
		b.WriteString(" to ")
		s.end.Format(b)
	}
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
	b.WriteString(fmt.Sprintf("%#v", s.rawPattern))
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
	b.WriteByte('(')
	s.expr.Format(b)
	b.WriteByte(')')
	b.WriteString(" is unknown")
}
