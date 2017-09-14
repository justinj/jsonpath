package jsonpath

import "fmt"

type validationVisitor struct {
	err         error
	filterDepth int
}

func (v *validationVisitor) VisitPre(n jsonPathNode) bool {
	if v.err != nil {
		return false
	}

	switch t := n.(type) {
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
	case FilterNode:
		v.filterDepth--
	}
}
