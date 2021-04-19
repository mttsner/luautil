package beautifier

import (
	"math"
	"strconv"

	"github.com/notnoobmaster/beautifier/ast"
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
	Functions *functions
}

func getString(expr *ast.Expr) (String string, Success bool) {
	switch ex := (*expr).(type) {
	case *ast.StringExpr:
		return ex.Value, true
	case *ast.NumberExpr:
		//return ex.Value, true
	}
	return 
}

func getBool(expr *ast.Expr) (Bool bool, Success bool, expression ast.Expr) {
	switch ex := (*expr).(type) {
	case *ast.TrueExpr:
		return true, true, &ast.TrueExpr{}
	case *ast.FalseExpr:
		return false, true, &ast.FalseExpr{}
	case *ast.StringExpr:
		return true, true, ex
	case *ast.IdentExpr:
		if ex.Value == "getfenv" {
			return true, true, ex
		}
	}
	return false, false, nil
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
		return &ast.NumberExpr{Value: float64(len(ex.Value))}
	case *ast.TableExpr:
		var length float64 = 0
		keys := make(map[float64]bool)
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
							length += float64(len(call.Args))
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
		for length++; keys[length]; length++ {}
		length--
		return &ast.NumberExpr{Value: length}
	}
	return nil
}

func arithmetic(op string, left *ast.Expr, right *ast.Expr) ast.Expr {
	l, okl := (*left).(*ast.NumberExpr)
	r, okr := (*right).(*ast.NumberExpr)

	if okl && okr {
		var result float64
		lv := l.Value
		rv := r.Value
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
		if !math.IsNaN(result) && !math.IsInf(result, 0) {
			return &ast.NumberExpr{Value: result}
		}
	}
	return nil
}

func (s *scope) logical(expr *ast.LogicalOpExpr) ast.Expr {
	l, okl, retl := getBool(&expr.Lhs)
	r, okr, retr := getBool(&expr.Rhs)
	switch expr.Operator {
	case "and":
		if !okl {
			return nil
		}
		if !l {
			return retl
		}
		if !okr {
			return nil
		}
		if r {
			return retr
		}
		return &ast.FalseExpr{}
	case "or":
		if okl && l {
			return retl
		}
		if !okr {
			return nil
		}
		if r {
			return retr
		}
		return &ast.FalseExpr{}
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

func isFoldable(function *ast.FunctionExpr) bool {
	if len(function.Stmts) == 1 {
		switch function.Stmts[0].(type) {
		case *ast.ReturnStmt:
			return true	
		}
	}
	return true
}


func (s *scope) call(expr *ast.FuncCallExpr) ast.Expr {
	if function, ok := expr.Func.(*ast.FunctionExpr); ok {


		if len(function.Stmts) == 1 {
			if ret, ok := function.Stmts[0].(*ast.ReturnStmt); ok && len(ret.Exprs) == 1 {
				return *s.Functions.compileExpr(&ret.Exprs[0])
			}
		}
		for i, name := range function.ParList.Names {
			s.Data[name] = expr.Args[i]
		}
	}
	return nil
}

func getValueOfIndex(index string, table *ast.TableExpr) ast.Expr {
	var key string
	for _, field := range table.Fields {
		switch ex := field.Key.(type) {
		case *ast.StringExpr:
			key = ex.Value
		case *ast.NumberExpr:
			//key = ex.Value
		}
		if key == index {
			return field.Value
		}
	}
	return nil
}

func (s *scope) index(expr *ast.AttrGetExpr) ast.Expr {
	switch ex := expr.Object.(type) {
	case *ast.IdentExpr:
		if data, ok := s.Data[ex.Value]; ok {
			if tbl, ok := data.(*ast.TableExpr); ok {
				switch key := expr.Key.(type) {
				case *ast.StringExpr:
					return getValueOfIndex(key.Value, tbl)
				case *ast.NumberExpr:
					//return getValueOfIndex(key.Value, tbl)
				}
			}
		}
	case *ast.TableExpr:
		switch key := expr.Key.(type) {
		case *ast.StringExpr:
			return getValueOfIndex(key.Value, ex)
		case *ast.NumberExpr:
			//return getValueOfIndex(key.Value, ex)
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
	new := "L_" + strconv.Itoa(s.LocalCount) + "_"
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
		LogicalOpExpr: s.logical,
	}
	s.Functions = f
	f.Traverse(Ast)
}