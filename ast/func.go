package ast

func isValid(str string) bool {
	for pos, ch := range str {
		if ch == '_' || 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || (('0' <= ch && ch <= '9') && pos > 0) {
			continue
		}
		return false
	}
	return true
}

func (s *builder) wrapIfNeeded(precedence int, associativity bool, op string, lhs Expr, rhs Expr, d data) {
	if precedence < d.Precedence || (precedence == d.Precedence && associativity != d.Direction) {
		s.add("(")
		s.expr(lhs, data{precedence, false, op})
		s.addpad(op)
		s.expr(rhs, data{precedence, true, op})
		s.add(")")
		return
	}
	s.expr(lhs, data{precedence, false, op})
	s.addpad(op)
	s.expr(rhs, data{precedence, true, op})
}
