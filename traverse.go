package beautifier

import (
	"github.com/notnoobmaster/beautifier/ast"
	"github.com/yuin/gopher-lua"
)

type constLValueExpr struct {
	ast.ExprBase

	Value lua.LValue
}

type functions struct {
	NewScope func() func()
	NewVariable func(variable string, remove func()) string
	IdentExpr func(variable string) ast.Expr
	//
	UnaryMinusOpExpr func(expr *ast.UnaryMinusOpExpr) ast.Expr
	UnaryNotOpExpr func(expr *ast.UnaryNotOpExpr) ast.Expr
	UnaryLenOpExpr func(expr *ast.UnaryLenOpExpr) ast.Expr
	//
	FuncCallExpr func(expr *ast.FuncCallExpr) ast.Expr
	AttrGetExpr func(expr *ast.AttrGetExpr) ast.Expr
	//
	StringConcatOpExpr func(left *ast.Expr, right *ast.Expr) ast.Expr
	RelationalOpExpr func(expr *ast.RelationalOpExpr) *ast.Expr
	ArithmeticOpExpr func(operator string, left *ast.Expr, right *ast.Expr) ast.Expr
	LogicalOpExpr func(*ast.LogicalOpExpr) ast.Expr
	//
	WhileStmt func(stmt *ast.WhileStmt)
	RepeatStmt func(stmt *ast.RepeatStmt)
	DoBlockStmt func(stmt *ast.DoBlockStmt)
	LocalAssignStmt func(stmt *ast.LocalAssignStmt)
	FuncDefStmt func(stmt *ast.FuncDefStmt)
	AssignStmt func(stmt *ast.AssignStmt)
	ReturnStmt func(stmt *ast.ReturnStmt)
	IfStmt func(stmt *ast.IfStmt)
	NumberForStmt func(stmt *ast.NumberForStmt)
	GenericForStmt func(stmt *ast.GenericForStmt)
	FuncCallStmt func(expr *ast.FuncCallExpr)
	BreakStmt func()
}

func (Functions *functions) compileTableExpr(expr *ast.TableExpr, ret *ast.Expr) *ast.Expr {
	for _, field := range expr.Fields {
		if field.Key != nil {
			field.Key = *Functions.compileExpr(&field.Key)
		}
		field.Value = *Functions.compileExpr(&field.Value)
	}
	return ret
}

func (Functions *functions) compileUnaryOpExpr(exprs *ast.Expr) *ast.Expr {
	switch ex := (*exprs).(type) {
	case *ast.UnaryMinusOpExpr:
		ex.Expr = *Functions.compileExpr(&ex.Expr)
		if Functions.UnaryMinusOpExpr != nil {
			if ret := Functions.UnaryMinusOpExpr(ex); ret != nil {
				return &ret
			}
		}
	case *ast.UnaryNotOpExpr:
		ex.Expr = *Functions.compileExpr(&ex.Expr)
		if Functions.UnaryNotOpExpr != nil {
			if ret := Functions.UnaryNotOpExpr(ex); ret != nil {
				return &ret
			}
		}
	case *ast.UnaryLenOpExpr:
		ex.Expr = *Functions.compileExpr(&ex.Expr)
		if Functions.UnaryLenOpExpr != nil {
			if ret := Functions.UnaryLenOpExpr(ex); ret != nil {
				return &ret
			}
		}
	}
	return exprs
}

func (Functions *functions) compileRelationalOpExpr(expr *ast.RelationalOpExpr, ret *ast.Expr) *ast.Expr {
	expr.Lhs = *Functions.compileExpr(&expr.Lhs)
	expr.Rhs = *Functions.compileExpr(&expr.Rhs)
	if Functions.RelationalOpExpr != nil {
		if ex := Functions.RelationalOpExpr(expr); ex != nil {
			return ex
		}
	}
	return ret
}

func (Functions *functions) compileArithmeticOpExpr(expr *ast.ArithmeticOpExpr, ret *ast.Expr) *ast.Expr {
	expr.Lhs = *Functions.compileExpr(&expr.Lhs)
	expr.Rhs = *Functions.compileExpr(&expr.Rhs)
	if Functions.ArithmeticOpExpr != nil {
		if ex := Functions.ArithmeticOpExpr(expr.Operator, &expr.Lhs, &expr.Rhs); ex != nil {
			return &ex
		}
	}
	return ret
}

func (Functions *functions) compileStringConcatOpExpr(expr *ast.StringConcatOpExpr, ret *ast.Expr) *ast.Expr {
	expr.Lhs = *Functions.compileExpr(&expr.Lhs)
	expr.Rhs = *Functions.compileExpr(&expr.Rhs)
	if Functions.StringConcatOpExpr != nil {
		if ex := Functions.StringConcatOpExpr(&expr.Lhs, &expr.Rhs); ex != nil {
			return &ex
		}
	}
	return ret
}

func (Functions *functions) compileLogicalOpExpr(expr *ast.LogicalOpExpr, ret *ast.Expr) *ast.Expr {
	expr.Lhs = *Functions.compileExpr(&expr.Lhs)
	expr.Rhs = *Functions.compileExpr(&expr.Rhs)
	if Functions.LogicalOpExpr != nil {
		if ex := Functions.LogicalOpExpr(expr); ex != nil {
			return &ex
		}
	}
	return ret
}

func (Functions *functions) compileFunctionExpr(expr *ast.FunctionExpr) {
	endScope := Functions.startScope()
	if Functions.NewVariable != nil {
		for i, name := range expr.ParList.Names {
			expr.ParList.Names[i] = Functions.NewVariable(name, func(){})
		}
	}
	if Functions.FuncCallStmt != nil {
		return 
	}
	Functions.Traverse(expr.Stmts)
	endScope()
}

func (Functions *functions) compileFuncCallExpr(expr *ast.FuncCallExpr, ret *ast.Expr) *ast.Expr {
	for i, arg := range expr.Args {
		expr.Args[i] = *Functions.compileExpr(&arg)
	}
	if Functions.FuncCallExpr != nil {
		if ex := Functions.FuncCallExpr(expr); ex != nil {
			return &ex
		}
	}
	if expr.Func != nil { // hoge.func()
		expr.Func = *Functions.compileExpr(&expr.Func)
	} else { // hoge:method()
		expr.Receiver = *Functions.compileExpr(&expr.Receiver)
	}
	return ret
}

func (Functions *functions) compileIdentExpr(expr *ast.IdentExpr, ret *ast.Expr) *ast.Expr {
	if Functions.IdentExpr != nil {
		if ex := Functions.IdentExpr(expr.Value); ex != nil {
			return &ex
		}
	}
	return ret
}

func (Functions *functions) compileAttrGetExpr(expr *ast.AttrGetExpr, ret *ast.Expr) *ast.Expr {
	expr.Key = *Functions.compileExpr(&expr.Key)
	if Functions.AttrGetExpr != nil {
		if ex := Functions.AttrGetExpr(expr); ex != nil {
			return &ex
		} 
	}
	expr.Object = *Functions.compileExpr(&expr.Object)
	return ret 
}


func (Functions *functions) compileExpr(expr *ast.Expr) *ast.Expr {
	switch ex := (*expr).(type) {
	case *ast.StringExpr:

	case *ast.NumberExpr:
		
	case *ast.NilExpr:

	case *ast.FalseExpr:

	case *ast.TrueExpr:

	case *ast.IdentExpr:
		return Functions.compileIdentExpr(ex, expr)
	case *ast.Comma3Expr:

	case *ast.AttrGetExpr:
		return Functions.compileAttrGetExpr(ex, expr)
	case *ast.TableExpr:
		return Functions.compileTableExpr(ex, expr)
	case *ast.ArithmeticOpExpr:
		return Functions.compileArithmeticOpExpr(ex, expr)
	case *ast.StringConcatOpExpr:
		return Functions.compileStringConcatOpExpr(ex, expr)
	case *ast.UnaryMinusOpExpr, *ast.UnaryNotOpExpr, *ast.UnaryLenOpExpr:
		return Functions.compileUnaryOpExpr(expr)
	case *ast.RelationalOpExpr:
		return Functions.compileRelationalOpExpr(ex, expr)
	case *ast.LogicalOpExpr:
		return Functions.compileLogicalOpExpr(ex, expr)
	case *ast.FuncCallExpr:
		return Functions.compileFuncCallExpr(ex, expr)
	case *ast.FunctionExpr:
		Functions.compileFunctionExpr(ex)
	}
	return expr
}

func (Functions *functions) compileAssignStmt(stmt *ast.AssignStmt) {
	for i := range stmt.Lhs {
		if ex := Functions.compileExpr(&stmt.Lhs[i]); ex != nil {
			stmt.Lhs[i] = *ex
		}
	}
	for i := range stmt.Rhs {
		if ex := Functions.compileExpr(&stmt.Rhs[i]); ex != nil {
			stmt.Rhs[i] = *ex
		}
	}
	if Functions.AssignStmt != nil {
		Functions.AssignStmt(stmt)
	}
}


func (Functions *functions) compileLocalAssignStmt(stmt *ast.LocalAssignStmt) {
	if len(stmt.Exprs) > 0 {
		for i, expr := range stmt.Exprs {
			stmt.Exprs[i] = *Functions.compileExpr(&expr)
		}
	}
	if Functions.NewVariable != nil {
		for i, name := range stmt.Names {
			stmt.Names[i] = Functions.NewVariable(name, func(){
				stmt.Names = append(stmt.Names[:i], stmt.Names[i+1:]...)
				if len(stmt.Exprs) > i {
					stmt.Exprs = append(stmt.Exprs[:i], stmt.Exprs[i+1:]...)
				}
			})
		}
	}
	if Functions.LocalAssignStmt != nil {
		Functions.LocalAssignStmt(stmt)
	}
}

func (Functions *functions) compileReturnStmt(stmt *ast.ReturnStmt) {
	for i := range stmt.Exprs {
		if ex := Functions.compileExpr(&stmt.Exprs[i]); ex != nil {
			stmt.Exprs[i] = *ex
		}
	}
	if Functions.ReturnStmt != nil {
		Functions.ReturnStmt(stmt)
	}
}
// USELESS
func (Functions *functions) compileBranchCondition(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.UnaryNotOpExpr:
		Functions.compileBranchCondition(ex.Expr)
	case *ast.LogicalOpExpr:
		Functions.compileBranchCondition(ex.Lhs)
		Functions.compileBranchCondition(ex.Rhs)
	case *ast.RelationalOpExpr:
		Functions.compileExpr(&ex.Lhs)
		Functions.compileExpr(&ex.Rhs)
	default:
		Functions.compileExpr(&expr)
	}
}


func (Functions *functions) compileIfStmt(stmt *ast.IfStmt) {
	if ex := Functions.compileExpr(&stmt.Condition); ex != nil {
		stmt.Condition = *Functions.compileExpr(&stmt.Condition)
	}
	if Functions.IfStmt != nil {
		if len(stmt.Else) > 0 {
			Functions.IfStmt(stmt)
		} else {
			Functions.IfStmt(stmt)
		}
	} else{
		endScope := Functions.startScope()
		Functions.Traverse(stmt.Then)
		endScope()
		if len(stmt.Else) > 0 {
			endScope = Functions.startScope()
			Functions.Traverse(stmt.Else)
			endScope()
		}
	}
}

func (Functions *functions) compileNumberForStmt(stmt *ast.NumberForStmt) {
	if ex := Functions.compileExpr(&stmt.Init); ex != nil {
		stmt.Init = *ex
	}
	if ex := Functions.compileExpr(&stmt.Limit); ex != nil {
		stmt.Limit = *ex
	}
	if stmt.Step != nil {
		if ex := Functions.compileExpr(&stmt.Step); ex != nil {
			stmt.Step = *ex
		}
	}
	endScope := Functions.startScope()
	if Functions.NewVariable != nil {
		stmt.Name = Functions.NewVariable(stmt.Name, func(){})
	}
	if Functions.NumberForStmt != nil {
		Functions.NumberForStmt(stmt)
	} else {
		Functions.Traverse(stmt.Stmts)
		endScope()
	}
}

func (Functions *functions) compileGenericForStmt(stmt *ast.GenericForStmt) {
	for i := range stmt.Exprs {
		if ex := Functions.compileExpr(&stmt.Exprs[i]); ex != nil {
			stmt.Exprs[i] = *ex
		}
	}
	endScope := Functions.startScope()
	if Functions.NewVariable != nil {
		for i, name := range stmt.Names {
			stmt.Names[i] = Functions.NewVariable(name, func(){})
		}
	}
	if Functions.GenericForStmt != nil {
		Functions.GenericForStmt(stmt)
	} else {
		Functions.Traverse(stmt.Stmts)
		endScope()
	}
}

func (Functions *functions) compileFuncDefStmt(stmt *ast.FuncDefStmt) {
	if stmt.Name.Func == nil {
		if ex := Functions.compileExpr(&stmt.Name.Receiver); ex != nil {
			stmt.Name.Receiver = *ex
		}
		if ex := Functions.compileExpr(&[]ast.Expr{stmt.Func}[0]); ex != nil { // This is scuffed af
			//stmt.Func = *ex
			panic("figure this shit out")
		}
	} else {
		astmt := &ast.AssignStmt{Lhs: []ast.Expr{stmt.Name.Func}, Rhs: []ast.Expr{stmt.Func}}
		Functions.compileAssignStmt(astmt)
	}
	if Functions.FuncDefStmt != nil {
		Functions.FuncDefStmt(stmt)
	}
}

func (Functions *functions) compileRepeatStmt(stmt *ast.RepeatStmt) {
	if ex := Functions.compileExpr(&stmt.Condition); ex != nil {
		stmt.Condition = *Functions.compileExpr(&stmt.Condition)
	}
	if Functions.RepeatStmt != nil {
		Functions.RepeatStmt(stmt)
	} else {
		endScope := Functions.startScope()
		Functions.Traverse(stmt.Stmts)
		endScope()
	}
}

func (Functions *functions) compileWhileStmt(stmt *ast.WhileStmt) {
	if ex := Functions.compileExpr(&stmt.Condition); ex != nil {
		stmt.Condition = *Functions.compileExpr(&stmt.Condition)
	}
	if Functions.WhileStmt != nil {
		Functions.WhileStmt(stmt)
	} else {
		endScope := Functions.startScope()
		Functions.Traverse(stmt.Stmts)
		endScope()
	}
}

func (Functions *functions) startScope() func() {
	if Functions.NewScope != nil {
		return Functions.NewScope()
	}
	return func(){}
}

func (Functions *functions) compileDoBlockStmt(stmt *ast.DoBlockStmt) {
	if Functions.DoBlockStmt != nil {
		Functions.DoBlockStmt(stmt)
	} else {
		endScope := Functions.startScope()
		Functions.Traverse(stmt.Stmts)
		endScope()
	}
}

func (Functions *functions) compileBreakStmt() {
	if Functions.BreakStmt != nil {
		Functions.BreakStmt()
	}
}

func (Functions *functions) compileFuncCallStmt(stmt *ast.FuncCallStmt) {
	Functions.compileFuncCallExpr(stmt.Expr.(*ast.FuncCallExpr), nil)
	if Functions.FuncCallStmt != nil {
		Functions.FuncCallStmt(stmt.Expr.(*ast.FuncCallExpr))
	}
}

func (Functions *functions) compileStmt(chunk []ast.Stmt) {
	for _, stmt := range chunk {
		switch st := stmt.(type) {
		case *ast.AssignStmt:
			Functions.compileAssignStmt(st)
		case *ast.LocalAssignStmt:
			Functions.compileLocalAssignStmt(st)
		case *ast.FuncCallStmt:
			Functions.compileFuncCallStmt(st)
		case *ast.DoBlockStmt:
			Functions.compileDoBlockStmt(st)
		case *ast.WhileStmt:
			Functions.compileWhileStmt(st)
		case *ast.RepeatStmt:
			Functions.compileRepeatStmt(st)
		case *ast.FuncDefStmt:
			Functions.compileFuncDefStmt(st)
		case *ast.ReturnStmt:
			Functions.compileReturnStmt(st)
		case *ast.IfStmt:
			Functions.compileIfStmt(st)
		case *ast.BreakStmt:
			Functions.compileBreakStmt()
		case *ast.NumberForStmt:
			Functions.compileNumberForStmt(st)
		case *ast.GenericForStmt:
			Functions.compileGenericForStmt(st)
		}
	}
}

// Traverse mhm
func (Functions *functions) Traverse(ast []ast.Stmt) {
	Functions.compileStmt(ast)
}