package luautil

import (
	"math"
	"strconv"

	"github.com/notnoobmaster/luautil/ast"
)

func getString(expr ast.Expr) (String string, Success bool) {
	switch ex := expr.(type) {
	case *ast.StringExpr:
		return ex.Value, true
	case *ast.NumberExpr:
		return strconv.FormatFloat(ex.Value, 'f', -1, 64), true
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

func concat(e *ast.StringConcatOpExpr) ast.Expr {
	strl, okl := getString(e.Lhs)
	strr, okr := getString(e.Rhs)
	if okl && okr {
		return &ast.StringExpr{Value: strl+strr}
	}
	return nil
}

func arithmetic(e *ast.ArithmeticOpExpr) ast.Expr {
	l, okl := e.Lhs.(*ast.NumberExpr)
	r, okr := e.Rhs.(*ast.NumberExpr)

	if okl && okr {
		var result float64
		lv := l.Value
		rv := r.Value
		switch e.Operator {
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

func logical(expr *ast.LogicalOpExpr) ast.Expr {
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

// Optimize the Abstract Syntax Tree 
func Optimize(Ast []ast.Stmt) {
	f := &Handlers{
		StringConcatOpExpr: concat,
		ArithmeticOpExpr: arithmetic,
		LogicalOpExpr: logical,
	}
	f.Traverse(Ast)
}