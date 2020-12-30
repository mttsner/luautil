package beautifier

import (
	"fmt"
	"math"
	"github.com/yuin/gopher-lua/ast"
	"strconv"
)

type local struct {
	Remove func()
	Constant bool
	Referenced bool
	Original string
}

type locals map[string]*local

type scope struct {
	Temp map[string]string
	Locals locals
	Data map[string]ast.Expr
	LocalCount int
}

func getString(expr *ast.Expr) (str string, success bool) {
	if str, ok := (*expr).(*ast.StringExpr); ok {
		return str.Value, ok
	}
	return 
}

func concat(left *ast.Expr, right *ast.Expr) ast.Expr {
	strl, okl := getString(left)
	strr, okr := getString(right)
	if okl && okr {
		return &ast.StringExpr{Value: strl+strr}
	}
	return nil
}

func length(expr *ast.UnaryLenOpExpr) ast.Expr {
	switch ex := expr.Expr.(type) {
	case *ast.StringExpr:
		return &ast.NumberExpr{Value: fmt.Sprint(len(ex.Value))}
	case *ast.TableExpr:
		length := 0
		keys := make(map[string]bool)
		for _, field := range ex.Fields {
			if call, ok := field.Value.(*ast.FuncCallExpr); ok {
				if function, ok := call.Func.(*ast.FunctionExpr); ok {
					if function.ParList.HasVargs && len(function.Stmts) == 1 {
						if ret, ok := function.Stmts[0].(*ast.ReturnStmt); ok {
							for _, v := range ret.Exprs {
								if _, ok := v.(*ast.NumberExpr); ok {
									length++
								}
							}
							length += len(call.Args)
							continue
						}
					}
				}
			}
			if field.Key == nil {
				length++
			} else if num, ok := field.Key.(*ast.NumberExpr); ok {
				keys[num.Value] = true
			}
		}
		for length++; keys[fmt.Sprint(length)]; length++ {}
		length--
		return &ast.NumberExpr{Value: fmt.Sprint(length)}
	}
	return nil
}

func stringify(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func arithmetic(op string, left *ast.Expr, right *ast.Expr) ast.Expr {
	l, okl := (*left).(*ast.NumberExpr)
	r, okr := (*right).(*ast.NumberExpr)

	if okl && okr {
		var result float64
		lv, _ := strconv.ParseFloat(l.Value, 64)
		rv, _ := strconv.ParseFloat(r.Value, 64)
		switch op {
		case "+":
			result = lv+rv
		case "-":
			result = lv-rv
		case "*":
			result = lv*rv
		case "/":
			result = lv/rv
		case "%":
			result = math.Mod(lv,rv)
		case "^":
			result = math.Pow(lv, rv)
		}
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return nil
		} else {
			return &ast.NumberExpr{Value: stringify(result)}
		}
	}
	return nil
}

func (s *scope) newScope() func() {
	old := make(map[string]string)
	for k, v := range s.Temp{
		old[k] = v
	}
	return func() {
		s.Temp = old
	}
}

func (s *scope) call(expr *ast.FuncCallExpr) {
	if ex, ok := expr.Func.(*ast.FunctionExpr); ok {
		for i, name := range ex.ParList.Names {
			s.Data[name] = expr.Args[i]
		}
	}
}

func getValueOfIndex(index string, table *ast.TableExpr) ast.Expr {
	var key string
	for _, field := range table.Fields {
		switch ex := field.Key.(type) {
		case *ast.StringExpr:
			key = ex.Value
		case *ast.NumberExpr:
			key = ex.Value
		}
		if key == index {
			return field.Value
		}
	}
	return nil
}

func (s *scope) index(expr *ast.AttrGetExpr) ast.Expr {
	if ident, ok := expr.Object.(*ast.IdentExpr); ok {
		if data, ok := s.Data[ident.Value]; ok {
			if tbl, ok := data.(*ast.TableExpr); ok {
				switch key := expr.Key.(type) {
				case *ast.StringExpr:
					return getValueOfIndex(key.Value, tbl)
				case *ast.NumberExpr:
					return getValueOfIndex(key.Value, tbl)
				}
			}
		}
	}
	return nil
}

func (s *scope) local(stmt *ast.LocalAssignStmt) {
	for i, name := range stmt.Names {
		if s.Locals[name].Constant {
			if i < len(stmt.Exprs) {
				switch ex := stmt.Exprs[i].(type) {
				case *ast.NumberExpr:
					s.Data[name] = ex
				case *ast.StringExpr:
					s.Data[name] = ex
				}
			}
		}
	}
}

func (s *scope) fold(local string) ast.Expr {
	if s.Locals[local] != nil && s.Locals[local].Constant {
		return s.Data[local]
	}
	return nil
}

func (s *scope) newName(local string) ast.Expr {
	if val, ok := s.Temp[local]; ok {
		s.Locals[val].Referenced = true
		return &ast.IdentExpr{Value: val}
	}
	return nil
}

func (s *scope) genName(str string, remove func()) string {
	s.LocalCount++
	new := "L_" + fmt.Sprint(s.LocalCount) + "_"
	s.Temp[str] = new
	s.Locals[new] = &local{Original: str, Constant: true, Remove: remove}
	return new
}

func (s *scope) assign(stmt *ast.AssignStmt) {
	for _, ex := range stmt.Lhs {
		if ident, ok := ex.(*ast.IdentExpr); ok {
			if s.Locals[ident.Value] == nil {
				continue
			}
			s.Locals[ident.Value].Constant = false
		}
	}
}

// Optimize the Abstract Syntax Tree 
func Optimize(Ast []ast.Stmt) {
	s := &scope{
		Locals: make(locals),
		Data: make(map[string]ast.Expr),
		Temp: make(map[string]string),
	}
	c := &functions{
		IdentExpr: s.newName,
		NewVariable: s.genName,
		NewScope: s.newScope,
		AssignStmt: s.assign,
	}
	c.Traverse(Ast)
	f := &functions{
		StringConcatOpExpr: concat,
		UnaryLenOpExpr: length,
		ArithmeticOpExpr: arithmetic,
		IdentExpr: s.fold,
		FuncCallExpr: s.call,
		AttrGetExpr: s.index,
		LocalAssignStmt: s.local,
	}
	f.Traverse(Ast)
}