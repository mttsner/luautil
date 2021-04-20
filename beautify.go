package beautifier

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/notnoobmaster/beautifier/ast"
)

type data struct {
	Precedence int
	Direction  bool // true: right, false: left
	Parent     string
}

type someStruct struct {
	Str  *strings.Builder
	Tabs int
	Data *data
}

func (s *someStruct) add(str string) {
	s.Str.WriteString(str)
}

func (s *someStruct) addln(str string) {
	s.Str.WriteString(str + "\n")
}

func (s *someStruct) addRune(r rune) {
	s.Str.WriteRune(r)
}

func (s *someStruct) addpad(str string) {
	s.Str.WriteString(" " + str + " ")
}

func (s *someStruct) tab() *someStruct {
	s.Str.WriteString(strings.Repeat("\t", s.Tabs))
	return s
}

func (s *someStruct) wrap(expr ast.Expr) {
	s.add("(")
	s.beautifyExpr(expr)
	s.add(")")
}

func (s *someStruct) addComma(idx int, length int) {
	if idx < length-1 {
		s.Str.WriteString(", ")
	}
}

func isValid(str string) bool {
	for pos, ch := range str {
		if ch == '_' || 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || (('0' <= ch && ch <= '9') && pos > 0) {
			continue
		}
		return false
	}
	return true
}

func (s *someStruct) beautifyExpr(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.NumberExpr:
		s.add(strconv.FormatFloat(ex.Value, 'f', -1, 64))
	case *ast.NilExpr:
		s.add("nil")
	case *ast.FalseExpr:
		s.add("false")
	case *ast.TrueExpr:
		s.add("true")
	case *ast.IdentExpr:
		s.add(ex.Value)
	case *ast.Comma3Expr:
		s.add("...")
	case *ast.StringExpr:
		s.addRune('"')
		for i, ch := range ex.Value {
			switch ch {
			case '\a', '\b', '\f', '\n', '\r', '\t', '\v', '\\', '"':
				s.addRune('\\')
				s.addRune(ch)
			case 65533:
				s.addRune('\\')
				s.Str.WriteString(fmt.Sprint([]byte(ex.Value)[i])) //TODO use strconv
			default:
				s.addRune(ch)
			}
		}
		s.addRune('"')
	case *ast.AttrGetExpr:
		switch obj := ex.Object.(type) {
		case *ast.IdentExpr, *ast.AttrGetExpr:
			s.beautifyExpr(ex.Object)
		case *ast.StringExpr:
			if obj.Value == "" {
				s.add("string")
				break
			}
			s.wrap(ex.Object)
		default:
			s.wrap(ex.Object)
		}

		if str, ok := ex.Key.(*ast.StringExpr); ok && isValid(str.Value) {
			s.add(".")
			s.add(str.Value)
		} else {
			s.add("[")
			s.beautifyExpr(ex.Key)
			s.add("]")
		}
	case *ast.TableExpr:
		s.add("{")
		s.Tabs++
		length := len(ex.Fields)
		for idx, field := range ex.Fields {
			s.addln("")
			s.tab()
			if field.Key != nil {
				if str, ok := field.Key.(*ast.StringExpr); ok && isValid(str.Value) {
					s.add(str.Value)
				} else {
					s.add("[")
					s.beautifyExpr(field.Key)
					s.add("]")
				}
				s.add(" = ")
			}
			s.beautifyExpr(field.Value)
			if idx < length-1 {
				s.Str.WriteRune(',')
				continue
			}
			s.addln("")
			s.Tabs--
			s.tab()
			s.Tabs++
		}
		s.Tabs--
		s.add("}")
	case *ast.ArithmeticOpExpr, *ast.StringConcatOpExpr, *ast.RelationalOpExpr, *ast.LogicalOpExpr:
		var currentPrecedence int
		var operator string
		var associativity bool
		var Lhs ast.Expr
		var Rhs ast.Expr

		switch ex := expr.(type) {
		case *ast.LogicalOpExpr:
			switch ex.Operator {
			case "or":
				currentPrecedence = 1
			case "and":
				currentPrecedence = 2
			}
			operator = ex.Operator
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *ast.RelationalOpExpr:
			currentPrecedence = 3
			operator = ex.Operator
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *ast.StringConcatOpExpr:
			currentPrecedence = 5
			operator = ".."
			associativity = true
			Lhs = ex.Lhs
			Rhs = ex.Rhs
		case *ast.ArithmeticOpExpr:
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
			s.beautifyExpr(Lhs)
			s.addpad(operator)
			s.Data = &data{currentPrecedence, true, operator}
			s.beautifyExpr(Rhs)
			s.add(")")
		} else {
			s.Data = &data{currentPrecedence, false, operator}
			s.beautifyExpr(Lhs)
			s.addpad(operator)
			s.Data = &data{currentPrecedence, true, operator}
			s.beautifyExpr(Rhs)
		}
		s.Data = &data{} // Reset the data
	case *ast.UnaryMinusOpExpr, *ast.UnaryNotOpExpr, *ast.UnaryLenOpExpr, *ast.UnaryBitwiseNotOpExpr:
		var prefix string
		switch ex := expr.(type) {
		case *ast.UnaryMinusOpExpr:
			prefix = "-"
			expr = ex.Expr
		case *ast.UnaryNotOpExpr:
			prefix = "not "
			expr = ex.Expr
		case *ast.UnaryLenOpExpr:
			prefix = "#"
			expr = ex.Expr
		case *ast.UnaryBitwiseNotOpExpr:
			
			prefix = "~"
			expr = ex.Expr
		}
		fmt.Println("Bitwise not")
		// Skidded from luamin.js
		if 8 < s.Data.Precedence && !((s.Data.Parent == "^") && s.Data.Direction == true) {
			s.Data = &data{Precedence: 8}
			s.add("(" + prefix)
			s.beautifyExpr(expr)
			s.add(")")
		} else {
			s.Data = &data{Precedence: 8}
			s.add(prefix)
			s.beautifyExpr(expr)
		}
		s.Data = &data{} // Reset the data
	case *ast.FuncCallExpr:
		if ex.Func != nil { // hoge.func()
			switch ex.Func.(type) {
			case *ast.IdentExpr, *ast.TableExpr, *ast.AttrGetExpr:
				s.beautifyExpr(ex.Func)
			default:
				s.wrap(ex.Func)
			}
		} else { // hoge:method()
			s.beautifyExpr(ex.Receiver)
			s.add(":")
			s.add(ex.Method)
		}

		s.add("(")
		for i := range ex.Args {
			s.beautifyExpr(ex.Args[i])
			s.addComma(i, len(ex.Args))
		}
		s.add(")")
	case *ast.FunctionExpr:
		s.add("function(")
		for i, name := range ex.ParList.Names {
			s.add(name)
			s.addComma(i, len(ex.ParList.Names))
		}
		if ex.ParList.HasVargs {
			if len(ex.ParList.Names) > 0 {
				s.add(", ")
			}
			s.add("...")
		}
		s.addln(")")
		s.beautifyStmt(ex.Stmts)
		s.tab().add("end")
	}
}

func (s *someStruct) elseBody(elseStmt []ast.Stmt) {
	if len(elseStmt) > 0 {
		if elseif, ok := elseStmt[0].(*ast.IfStmt); ok && len(elseStmt) == 1 {
			s.tab().add("elseif ")
			s.beautifyExpr(elseif.Condition)
			s.addln(" then")
			s.beautifyStmt(elseif.Then)
			s.elseBody(elseif.Else)
		} else {
			s.tab().addln("else")
			s.beautifyStmt(elseStmt)
		}
	}
}

func (s *someStruct) beautifyStmt(chunk []ast.Stmt) {
	s.Tabs++
	for _, st := range chunk {
		s.tab()
		switch stmt := st.(type) {
		case *ast.AssignStmt:
			for i, ex := range stmt.Lhs {
				s.beautifyExpr(ex)
				s.addComma(i, len(stmt.Lhs))
			}
			s.addpad("=")
			for i, ex := range stmt.Rhs {
				s.beautifyExpr(ex)
				s.addComma(i, len(stmt.Rhs))
			}
		case *ast.CompoundAssignStmt:
			for i, ex := range stmt.Lhs {
				s.beautifyExpr(ex)
				s.addComma(i, len(stmt.Lhs))
			}
			s.addpad(stmt.Operator)
			for i, ex := range stmt.Rhs {
				s.beautifyExpr(ex)
				s.addComma(i, len(stmt.Rhs))
			}
		case *ast.LocalAssignStmt:
			s.add("local ")
			for i, name := range stmt.Names {
				s.add(name)
				s.addComma(i, len(stmt.Names))
			}
			if len(stmt.Exprs) > 0 {
				s.add(" = ")
				for i, ex := range stmt.Exprs {
					s.beautifyExpr(ex)
					s.addComma(i, len(stmt.Exprs))
				}
			}
		case *ast.FuncCallStmt:
			ex := stmt.Expr.(*ast.FuncCallExpr)
			if ex.Func != nil {
				switch ex.Func.(type) {
				case *ast.IdentExpr, *ast.TableExpr, *ast.AttrGetExpr:
					s.beautifyExpr(ex.Func)
				default:
					s.wrap(ex.Func)
				}
			} else {
				s.beautifyExpr(ex.Receiver)
				s.add(":")
				s.add(ex.Method)
			}

			s.add("(")
			for i := range ex.Args {
				s.beautifyExpr(ex.Args[i])
				s.addComma(i, len(ex.Args))
			}
			s.add(")")
		case *ast.DoBlockStmt:
			s.addln("do")
			s.beautifyStmt(stmt.Stmts)
			s.tab().add("end")
		case *ast.WhileStmt:
			s.add("while ")
			s.beautifyExpr(stmt.Condition)
			s.addln(" do")
			s.beautifyStmt(stmt.Stmts)
			s.tab().add("end")
		case *ast.RepeatStmt:
			s.addln("repeat")
			s.beautifyStmt(stmt.Stmts)
			s.tab().add("until ")
			s.beautifyExpr(stmt.Condition)
		case *ast.FuncDefStmt:
			s.add("function ")
			if stmt.Name.Func == nil {
				s.beautifyExpr(stmt.Name.Receiver)
				s.Str.WriteRune(':')
				s.add(stmt.Name.Method)
			} else {
				s.beautifyExpr(stmt.Name.Func)
			}
			s.addRune('(')
			for i, name := range stmt.Func.ParList.Names {
				s.add(name)
				s.addComma(i, len(stmt.Func.ParList.Names))
			}
			s.addln(")")
			s.beautifyStmt(stmt.Func.Stmts)
			s.tab().add("end")
		case *ast.ReturnStmt:
			s.add("return ")
			for i, ex := range stmt.Exprs {
				s.beautifyExpr(ex)
				s.addComma(i, len(stmt.Exprs))
			}
		case *ast.IfStmt:
			s.add("if ")
			s.beautifyExpr(stmt.Condition)
			s.addln(" then")
			s.beautifyStmt(stmt.Then)
			s.elseBody(stmt.Else)
			s.tab().add("end")
		case *ast.BreakStmt:
			s.add("break")
		case *ast.ContinueStmt:
			s.add("continue")
		case *ast.NumberForStmt:
			s.add("for ")
			s.add(stmt.Name)
			s.add(" = ")
			s.beautifyExpr(stmt.Init)
			s.add(", ")
			s.beautifyExpr(stmt.Limit)
			if stmt.Step != nil {
				s.add(", ")
				s.beautifyExpr(stmt.Step)
			}
			s.addln(" do")
			s.beautifyStmt(stmt.Stmts)
			s.tab().add("end")
		case *ast.GenericForStmt:
			s.add("for ")
			for i, name := range stmt.Names {
				s.add(name)
				s.addComma(i, len(stmt.Names))
			}
			s.add(" in ")
			for _, ex := range stmt.Exprs {
				s.beautifyExpr(ex)
			}
			s.addln(" do")
			s.beautifyStmt(stmt.Stmts)
			s.tab().add("end")
		}
		s.add(";\n")
	}
	s.Tabs--
}

// Beautify the Abstract Syntax Tree
func Beautify(ast []ast.Stmt) string {
	s := &someStruct{
		Str:  &strings.Builder{},
		Data: &data{},
		Tabs: -1, // Accounting for the fact that each beautifyStmt call increments Tabs by one
	}
	s.beautifyStmt(ast)
	return s.Str.String()
}
