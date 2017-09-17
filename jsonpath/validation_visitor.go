package jsonpath

import "fmt"

type validationVisitor struct {
	err                error
	filterDepth        int
	arrayAccessorDepth int
}

func (v *validationVisitor) VisitPre(n jsonPathNode) bool {
	if v.err != nil {
		return false
	}

	switch t := n.(type) {
	case ArrayAccessor:
		v.arrayAccessorDepth++
	case LastExpr:
		if v.arrayAccessorDepth == 0 {
			v.err = fmt.Errorf("`last` can only appear inside an array subscript")
		}
	case FilterNode:
		v.filterDepth++
	case VariableExpr:
		if t.name == "@" && v.filterDepth == 0 {
			v.err = fmt.Errorf("@ only allowed within filter expressions")
		}
	}
	return true
}

func (v *validationVisitor) VisitPost(n jsonPathNode) {
	switch n.(type) {
	case ArrayAccessor:
		v.arrayAccessorDepth--
	case FilterNode:
		v.filterDepth--
	}
}
