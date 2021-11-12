package ast

import (
	"strconv"
	"strings"
)

type data struct {
	Precedence int
	Direction  bool // true: right, false: left
	Parent     string
}

type builder struct {
	Str    *strings.Builder
	Indent int
}

// Helper functions
func (s *builder) add(str string)      { s.Str.WriteString(str) }
func (s *builder) addln(str string)    { s.Str.WriteString(str + "\n") }
func (s *builder) addrune(r rune)      { s.Str.WriteRune(r) }
func (s *builder) addpad(str string)   { s.Str.WriteString(" " + str + " ") }
func (s *builder) tab() *builder       { s.Str.WriteString(strings.Repeat("\t", s.Indent)); return s }
func (s *builder) wrap(e Expr, d data) { s.add("("); s.expr(e, d); s.add(")") }

func (s *builder) addcomma(idx int, length int) {
	if idx < length-1 {
		s.Str.WriteString(", ")
	}
}

func (s *builder) expr(ex Expr, d data) {
	switch e := ex.(type) {
	case *NumberExpr:
		s.add(strconv.FormatFloat(e.Value, 'f', -1, 64))
	case *NilExpr:
		s.add("nil")
	case *FalseExpr:
		s.add("false")
	case *TrueExpr:
		s.add("true")
	case *IdentExpr:
		s.add(e.Value)
	case *Comma3Expr:
		s.add("...")
	case *StringExpr:
		s.addrune('"')
		s.add(formatString(e.Value))
		s.addrune('"')
	case *AttrGetExpr:
		switch obj := e.Object.(type) {
		case *IdentExpr, *AttrGetExpr:
			s.expr(e.Object, d)
		case *StringExpr:
			if obj.Value == "" {
				s.add("string")
				break
			}
			s.wrap(e.Object, d)
		default:
			s.wrap(e.Object, d)
		}

		if str, ok := e.Key.(*StringExpr); ok && isValid(str.Value) {
			s.add(".")
			s.add(str.Value)
		} else {
			s.add("[")
			s.expr(e.Key, d)
			s.add("]")
		}
	case *TableExpr:
		s.add("{")
		s.Indent++
		length := len(e.Fields)
		for idx, field := range e.Fields {
			s.addln("")
			s.tab()
			if field.Key != nil {
				if str, ok := field.Key.(*StringExpr); ok && isValid(str.Value) {
					s.add(str.Value)
				} else {
					s.add("[")
					s.expr(field.Key, d)
					s.add("]")
				}
				s.add(" = ")
			}
			s.expr(field.Value, d)
			if idx < length-1 {
				s.Str.WriteRune(',')
				continue
			}
			s.addln("")
			s.Indent--
			s.tab()
			s.Indent++
		}
		s.Indent--
		s.add("}")
	case *LogicalOpExpr:
		if e.Operator == "or" {
			s.wrapIfNeeded(1, false, "or", e.Lhs, e.Rhs, d)
		} else {
			s.wrapIfNeeded(2, false, "and", e.Lhs, e.Rhs, d)
		}
	case *RelationalOpExpr:
		s.wrapIfNeeded(3, false, e.Operator, e.Lhs, e.Rhs, d)
	case *StringConcatOpExpr:
		s.wrapIfNeeded(5, true, "..", e.Lhs, e.Rhs, d)
	case *ArithmeticOpExpr:
		switch e.Operator {
		case "+", "-":
			s.wrapIfNeeded(6, false, e.Operator, e.Lhs, e.Rhs, d)
		case "*", "/", "%":
			s.wrapIfNeeded(7, false, e.Operator, e.Lhs, e.Rhs, d)
		case "^":
			s.wrapIfNeeded(10, false, "^", e.Lhs, e.Rhs, d)
		}
	case *UnaryOpExpr:
		if 8 < d.Precedence || d.Direction {
			s.add("(")
			s.add(e.Operator)
			s.expr(e.Expr, data{Precedence: 8})
			s.add(")")
		} else {
			s.add(e.Operator)
			s.expr(e.Expr, data{Precedence: 8})
		}
	case *FuncCallExpr:
		if e.Func != nil { // hoge.func()
			switch e.Func.(type) {
			case *IdentExpr, *TableExpr, *AttrGetExpr:
				s.expr(e.Func, d)
			default:
				s.wrap(e.Func, d)
			}
		} else { // hoge:method()
			s.expr(e.Receiver, d)
			s.add(":")
			s.add(e.Method)
		}

		s.add("(")
		for i := range e.Args {
			s.expr(e.Args[i], d)
			s.addcomma(i, len(e.Args))
		}
		s.add(")")
	case *FunctionExpr:
		s.add("function(")
		for i, name := range e.ParList.Names {
			s.add(name)
			s.addcomma(i, len(e.ParList.Names))
		}
		if e.ParList.HasVargs {
			if len(e.ParList.Names) > 0 {
				s.add(", ")
			}
			s.add("...")
		}
		s.addln(")")
		s.chunk(e.Chunk)
		s.tab().add("end")
	default:
		panic("Unimplemented expression")
	}
}

func (s *builder) elseBody(elseStmt []Stmt) {
	if len(elseStmt) > 0 {
		if elseif, ok := elseStmt[0].(*IfStmt); ok && len(elseStmt) == 1 {
			s.tab().add("elseif ")
			s.expr(elseif.Condition, data{})
			s.addln(" then")
			s.chunk(elseif.Then)
			s.elseBody(elseif.Else)
		} else {
			s.tab().addln("else")
			s.chunk(elseStmt)
		}
	}
}

func (b *builder) chunk(c Chunk) {
	b.Indent++
	for _, s := range c {
		b.stmt(s)
	}
	b.Indent--
}

func (s *builder) stmt(st Stmt) {
	s.tab()
	switch stmt := st.(type) {
	case *AssignStmt:
		for i, ex := range stmt.Lhs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Lhs))
		}
		s.addpad("=")
		for i, ex := range stmt.Rhs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Rhs))
		}
	case *CompoundAssignStmt:
		for i, ex := range stmt.Lhs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Lhs))
		}
		s.addpad(stmt.Operator)
		for i, ex := range stmt.Rhs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Rhs))
		}
	case *LocalAssignStmt:
		s.add("local ")
		for i, name := range stmt.Names {
			s.add(name)
			s.addcomma(i, len(stmt.Names))
		}
		if len(stmt.Exprs) > 0 {
			s.add(" = ")
			for i, ex := range stmt.Exprs {
				s.expr(ex, data{})
				s.addcomma(i, len(stmt.Exprs))
			}
		}
	case *FuncCallStmt:
		ex := stmt.Expr.(*FuncCallExpr)
		if ex.Func != nil {
			switch ex.Func.(type) {
			case *IdentExpr, *TableExpr, *AttrGetExpr:
				s.expr(ex.Func, data{})
			default:
				s.wrap(ex.Func, data{})
			}
		} else {
			s.expr(ex.Receiver, data{})
			s.add(":")
			s.add(ex.Method)
		}

		s.add("(")
		for i := range ex.Args {
			s.expr(ex.Args[i], data{})
			s.addcomma(i, len(ex.Args))
		}
		s.add(")")
	case *DoBlockStmt:
		s.addln("do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	case *WhileStmt:
		s.add("while ")
		s.expr(stmt.Condition, data{})
		s.addln(" do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	case *RepeatStmt:
		s.addln("repeat")
		s.chunk(stmt.Chunk)
		s.tab().add("until ")
		s.expr(stmt.Condition, data{})
	case *LocalFunctionStmt:
		s.add("local function ")
		s.add(stmt.Name)
		s.expr(stmt.Func, data{})
	case *FunctionStmt:
		s.add("function ")
		if stmt.Name.Func == nil {
			s.expr(stmt.Name.Receiver, data{})
			s.Str.WriteRune(':')
			s.add(stmt.Name.Method)
		} else {
			s.expr(stmt.Name.Func, data{})
		}
		s.addrune('(')
		for i, name := range stmt.Func.ParList.Names {
			s.add(name)
			s.addcomma(i, len(stmt.Func.ParList.Names))
		}
		s.addln(")")
		s.chunk(stmt.Func.Chunk)
		s.tab().add("end")
	case *ReturnStmt:
		s.add("return ")
		for i, ex := range stmt.Exprs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Exprs))
		}
	case *IfStmt:
		s.add("if ")
		s.expr(stmt.Condition, data{})
		s.addln(" then")
		s.chunk(stmt.Then)
		s.elseBody(stmt.Else)
		s.tab().add("end")
	case *BreakStmt:
		s.add("break")
	case *ContinueStmt:
		s.add("continue")
	case *NumberForStmt:
		s.add("for ")
		s.add(stmt.Name)
		s.add(" = ")
		s.expr(stmt.Init, data{})
		s.add(", ")
		s.expr(stmt.Limit, data{})
		if stmt.Step != nil {
			s.add(", ")
			s.expr(stmt.Step, data{})
		}
		s.addln(" do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	case *GenericForStmt:
		s.add("for ")
		for i, name := range stmt.Names {
			s.add(name)
			s.addcomma(i, len(stmt.Names))
		}
		s.add(" in ")
		for i, ex := range stmt.Exprs {
			s.expr(ex, data{})
			s.addcomma(i, len(stmt.Exprs))
		}
		s.addln(" do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	default:
		panic("Unimplemented statement")
	}
	s.add(";\n")
}
