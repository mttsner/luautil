package ast

import (
	"strings"
	"strconv"
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

// TODO: If it's a invalid char then peek the next char and check if it's a number. 
// In some edge cases like \10\48 it might break.
// However, for now the whole string is going to be escaped if any character is invalid
func formatString(str string) string {
	b := strings.Builder{}
	b.Grow(len(str)) // Min size of the output string.

	for _, ch := range str {
		switch ch {
		case '\a':
			b.WriteString("\\a")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\v':
			b.WriteString("\\v")
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		default:
			if ch < ' ' || ch > '~' {
				return escapeString(str)
			} else {
				b.WriteRune(ch)
			}
		}
	}
	return b.String()
}

// Escapes every character in string
func escapeString(str string) string {
	b := strings.Builder{}
	b.Grow(len(str)) // Min size of the output string.

	for i := range str {
		b.WriteRune('\\')
		b.WriteString(strconv.Itoa(int([]byte(str)[i])))
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
