package ast
import (
	"fmt"
	"strings"
)

func isValid(str string) bool {
	for pos, ch := range str {
		if ch == '_' || 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || (('0' <= ch && ch <= '9') && pos > 0) {
			continue
		}
		return false
	}
	return true
}

func formatString(str string) string {
	b := strings.Builder{}
	for i, ch := range str {
		switch ch {
		case '\a', '\b', '\f', '\n', '\r', '\t', '\v', '\\', '"':
			b.WriteRune('\\')
			b.WriteRune(ch)
		case 65533:
			b.WriteRune('\\')
			b.WriteString(fmt.Sprint([]byte(str)[i]))
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (s *builder) wrapIfNeeded(precedence int, associativity bool, op string, lhs Expr, rhs Expr, d data) {
	if precedence < d.Precedence || associativity != d.Direction {
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