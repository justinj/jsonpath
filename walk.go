package jsonpath

type visitor interface {
	VisitPre(jsonPathNode) (recurse bool)
	VisitPost(jsonPathNode)
}

func (n BinExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.left.Walk(v)
		n.right.Walk(v)
		v.VisitPost(n)
	}
}

func (n UnaryExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.expr.Walk(v)
		v.VisitPost(n)
	}
}

func (n ParenExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.expr.Walk(v)
		v.VisitPost(n)
	}
}

func (n NumberExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n VariableExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n LastExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n BoolExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n NullExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n StringExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n AccessExpr) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.left.Walk(v)
		n.right.Walk(v)
		v.VisitPost(n)
	}
}

func (n DotAccessor) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n MemberWildcardAccessor) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n ArrayAccessor) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		for _, e := range n.subscripts {
			e.Walk(v)
		}
		v.VisitPost(n)
	}
}

func (n RangeSubscriptNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.start.Walk(v)
		if n.end != nil {
			n.end.Walk(v)
		}
		v.VisitPost(n)
	}
}

func (n WildcardArrayAccessor) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		v.VisitPost(n)
	}
}

func (n FuncNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		if n.arg != nil {
			n.arg.Walk(v)
		}
		v.VisitPost(n)
	}
}

func (n FilterNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.pred.Walk(v)
		v.VisitPost(n)
	}
}

func (n ExistsNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.expr.Walk(v)
		v.VisitPost(n)
	}
}

func (n LikeRegexNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.left.Walk(v)
		v.VisitPost(n)
	}
}

func (n StartsWithNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.left.Walk(v)
		n.right.Walk(v)
		v.VisitPost(n)
	}
}

func (n IsUnknownNode) Walk(v visitor) {
	if rec := v.VisitPre(n); rec {
		n.expr.Walk(v)
		v.VisitPost(n)
	}
}
