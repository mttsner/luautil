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
	Data   *data
}

// Helper functions
func (s *builder) add(str string)    { s.Str.WriteString(str) }
func (s *builder) addln(str string)  { s.Str.WriteString(str + "\n") }
func (s *builder) addrune(r rune)    { s.Str.WriteRune(r) }
func (s *builder) addpad(str string) { s.Str.WriteString(" " + str + " ") }
func (s *builder) tab() *builder     { s.Str.WriteString(strings.Repeat("\t", s.Indent)); return s }
func (s *builder) wrap(e Expr)       { s.add("("); s.expr(e); s.add(")") }

func (s *builder) addcomma(idx int, length int) {
	if idx < length-1 {
		s.Str.WriteString(", ")
	}
}

func (s *builder) expr(ex Expr) {
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
			s.expr(e.Object)
		case *StringExpr:
			if obj.Value == "" {
				s.add("string")
				break
			}
			s.wrap(e.Object)
		default:
			s.wrap(e.Object)
		}

		if str, ok := e.Key.(*StringExpr); ok && isValid(str.Value) {
			s.add(".")
			s.add(str.Value)
		} else {
			s.add("[")
			s.expr(e.Key)
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
					s.expr(field.Key)
					s.add("]")
				}
				s.add(" = ")
			}
			s.expr(field.Value)
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
	case *ArithmeticOpExpr, *StringConcatOpExpr, *RelationalOpExpr, *LogicalOpExpr:
		var currentPrecedence int
		var operator string
		var associativity bool
		var Lhs Expr
		var Rhs Expr

		switch ex := ex.(type) {
		case *LogicalOpExpr:
			switch ex.Operator {
			case "or":
				currentPrecedence = 1
			case "and":
				currentPrecedence = 2
			}
			operator = ex.Operator
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *RelationalOpExpr:
			currentPrecedence = 3
			operator = ex.Operator
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *StringConcatOpExpr:
			currentPrecedence = 5
			operator = ".."
			associativity = true
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *ArithmeticOpExpr:
			switch ex.Operator {
			case "+", "-":
				currentPrecedence = 6
			case "*", "/", "%":
				currentPrecedence = 7
			case "^":
				currentPrecedence = 10
				associativity = true
			}
			operator = ex.Operator
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		}

		if currentPrecedence < s.Data.Precedence ||
			(currentPrecedence == s.Data.Precedence &&
				associativity != s.Data.Direction &&
				s.Data.Parent != "+" &&
				!(s.Data.Parent == "*" && (operator == "/" || operator == "*"))) {
			s.add("(")
			s.Data = &data{currentPrecedence, false, operator}
			s.expr(Lhs)
			s.addpad(operator)
			s.Data = &data{currentPrecedence, true, operator}
			s.expr(Rhs)
			s.add(")")
		} else {
			s.Data = &data{currentPrecedence, false, operator}
			s.expr(Lhs)
			s.addpad(operator)
			s.Data = &data{currentPrecedence, true, operator}
			s.expr(Rhs)
		}
		s.Data = &data{} // Reset the data
	case *UnaryOpExpr:
		if 8 < s.Data.Precedence && !((s.Data.Parent == "^") && s.Data.Direction) {
			s.Data = &data{Precedence: 8}
			s.add("(" + e.Operator)
			s.expr(e.Expr)
			s.add(")")
		} else {
			s.Data = &data{Precedence: 8}
			s.add(e.Operator)
			s.expr(e.Expr)
		}
		s.Data = &data{} // Reset the data
	case *FuncCallExpr:
		if e.Func != nil { // hoge.func()
			switch e.Func.(type) {
			case *IdentExpr, *TableExpr, *AttrGetExpr:
				s.expr(e.Func)
			default:
				s.wrap(e.Func)
			}
		} else { // hoge:method()
			s.expr(e.Receiver)
			s.add(":")
			s.add(e.Method)
		}

		s.add("(")
		for i := range e.Args {
			s.expr(e.Args[i])
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
	}
}

func (s *builder) elseBody(elseStmt []Stmt) {
	if len(elseStmt) > 0 {
		if elseif, ok := elseStmt[0].(*IfStmt); ok && len(elseStmt) == 1 {
			s.tab().add("elseif ")
			s.expr(elseif.Condition)
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
			s.expr(ex)
			s.addcomma(i, len(stmt.Lhs))
		}
		s.addpad("=")
		for i, ex := range stmt.Rhs {
			s.expr(ex)
			s.addcomma(i, len(stmt.Rhs))
		}
	case *CompoundAssignStmt:
		for i, ex := range stmt.Lhs {
			s.expr(ex)
			s.addcomma(i, len(stmt.Lhs))
		}
		s.addpad(stmt.Operator)
		for i, ex := range stmt.Rhs {
			s.expr(ex)
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
				s.expr(ex)
				s.addcomma(i, len(stmt.Exprs))
			}
		}
	case *FuncCallStmt:
		ex := stmt.Expr.(*FuncCallExpr)
		if ex.Func != nil {
			switch ex.Func.(type) {
			case *IdentExpr, *TableExpr, *AttrGetExpr:
				s.expr(ex.Func)
			default:
				s.wrap(ex.Func)
			}
		} else {
			s.expr(ex.Receiver)
			s.add(":")
			s.add(ex.Method)
		}

		s.add("(")
		for i := range ex.Args {
			s.expr(ex.Args[i])
			s.addcomma(i, len(ex.Args))
		}
		s.add(")")
	case *DoBlockStmt:
		s.addln("do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	case *WhileStmt:
		s.add("while ")
		s.expr(stmt.Condition)
		s.addln(" do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	case *RepeatStmt:
		s.addln("repeat")
		s.chunk(stmt.Chunk)
		s.tab().add("until ")
		s.expr(stmt.Condition)
	case *FunctionStmt:
		s.add("function ")
		if stmt.Name.Func == nil {
			s.expr(stmt.Name.Receiver)
			s.Str.WriteRune(':')
			s.add(stmt.Name.Method)
		} else {
			s.expr(stmt.Name.Func)
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
			s.expr(ex)
			s.addcomma(i, len(stmt.Exprs))
		}
	case *IfStmt:
		s.add("if ")
		s.expr(stmt.Condition)
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
		s.expr(stmt.Init)
		s.add(", ")
		s.expr(stmt.Limit)
		if stmt.Step != nil {
			s.add(", ")
			s.expr(stmt.Step)
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
		for _, ex := range stmt.Exprs {
			s.expr(ex)
		}
		s.addln(" do")
		s.chunk(stmt.Chunk)
		s.tab().add("end")
	}
	s.add(";\n")

}