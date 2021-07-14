package luautil

import (
	"github.com/notnoobmaster/luautil/ast"
)

type Handlers struct {
	// Expressions
	IdentExpr          func(*ast.IdentExpr) ast.Expr
	UnaryOpExpr        func(*ast.UnaryOpExpr) ast.Expr
	FuncCallExpr       func(*ast.FuncCallExpr) ast.Expr
	AttrGetExpr        func(*ast.AttrGetExpr) ast.Expr
	StringConcatOpExpr func(*ast.StringConcatOpExpr) ast.Expr
	RelationalOpExpr   func(*ast.RelationalOpExpr) *ast.Expr
	ArithmeticOpExpr   func(*ast.ArithmeticOpExpr) ast.Expr
	LogicalOpExpr      func(*ast.LogicalOpExpr) ast.Expr
	TableExpr          func(*ast.TableExpr) ast.Expr
	FunctionExpr       func(*ast.FunctionExpr) ast.Expr
	// Statements
	WhileStmt          func(*ast.WhileStmt)
	RepeatStmt         func(*ast.RepeatStmt)
	DoBlockStmt        func(*ast.DoBlockStmt)
	LocalAssignStmt    func(*ast.LocalAssignStmt)
	FuncDefStmt        func(*ast.FuncDefStmt)
	AssignStmt         func(*ast.AssignStmt)
	ReturnStmt         func(*ast.ReturnStmt)
	IfStmt             func(*ast.IfStmt)
	NumberForStmt      func(*ast.NumberForStmt)
	GenericForStmt     func(*ast.GenericForStmt)
	CompoundAssignStmt func(*ast.CompoundAssignStmt)
	LabelStmt          func(*ast.LabelStmt)
	GotoStmt           func(*ast.GotoStmt)
	BreakStmt          func()
	ContinueStmt       func()
}

func (h *Handlers) expr(ex *ast.Expr) *ast.Expr {
	switch e := (*ex).(type) {
	case *ast.IdentExpr:
		if h.IdentExpr != nil {
			if ex := h.IdentExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.AttrGetExpr:
		e.Key = *h.expr(&e.Key)
		e.Object = *h.expr(&e.Object)
		if h.AttrGetExpr != nil {
			if ex := h.AttrGetExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.TableExpr:
		for _, field := range e.Fields {
			if field.Key != nil {
				field.Key = *h.expr(&field.Key)
			}
			field.Value = *h.expr(&field.Value)
		}
		if h.TableExpr != nil {
			if ex := h.TableExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.ArithmeticOpExpr:
		e.Lhs = *h.expr(&e.Lhs)
		e.Rhs = *h.expr(&e.Rhs)
		if h.ArithmeticOpExpr != nil {
			if ex := h.ArithmeticOpExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.StringConcatOpExpr:
		e.Lhs = *h.expr(&e.Lhs)
		e.Rhs = *h.expr(&e.Rhs)
		if h.StringConcatOpExpr != nil {
			if ex := h.StringConcatOpExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.UnaryOpExpr:
		e.Expr = *h.expr(&e.Expr)
		if h.UnaryOpExpr != nil {
			if ret := h.UnaryOpExpr(e); ret != nil {
				return &ret
			}
		}
	case *ast.RelationalOpExpr:
		e.Lhs = *h.expr(&e.Lhs)
		e.Rhs = *h.expr(&e.Rhs)
		if h.RelationalOpExpr != nil {
			if ex := h.RelationalOpExpr(e); ex != nil {
				return ex
			}
		}
	case *ast.LogicalOpExpr:
		e.Lhs = *h.expr(&e.Lhs)
		e.Rhs = *h.expr(&e.Rhs)
		if h.LogicalOpExpr != nil {
			if ex := h.LogicalOpExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.FuncCallExpr:
		for i, arg := range e.Args {
			e.Args[i] = *h.expr(&arg)
		}
		if h.FuncCallExpr != nil {
			if ex := h.FuncCallExpr(e); ex != nil {
				return &ex
			}
		}
		if e.Func != nil { // hoge.func()
			e.Func = *h.expr(&e.Func)
		} else { // hoge:method()
			e.Receiver = *h.expr(&e.Receiver)
		}

		if h.FuncCallExpr != nil {
			if ex := h.FuncCallExpr(e); ex != nil {
				return &ex
			}
		}
	case *ast.FunctionExpr:
		if h.FunctionExpr != nil {
			h.FunctionExpr(e)
		} else {
			h.Traverse(e.Chunk)
		}
	}
	return ex
}

func (h *Handlers) compileStmt(chunk []ast.Stmt) {
	for _, stmt := range chunk {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			for i := range s.Lhs {
				if ex := h.expr(&s.Lhs[i]); ex != nil {
					s.Lhs[i] = *ex
				}
			}
			for i := range s.Rhs {
				if ex := h.expr(&s.Rhs[i]); ex != nil {
					s.Rhs[i] = *ex
				}
			}
			if h.AssignStmt != nil {
				h.AssignStmt(s)
			}
		case *ast.CompoundAssignStmt:
			for i := range s.Lhs {
				if ex := h.expr(&s.Lhs[i]); ex != nil {
					s.Lhs[i] = *ex
				}
			}
			for i := range s.Rhs {
				if ex := h.expr(&s.Rhs[i]); ex != nil {
					s.Rhs[i] = *ex
				}
			}
			if h.CompoundAssignStmt != nil {
				h.CompoundAssignStmt(s)
			}
		case *ast.LocalAssignStmt:
			if len(s.Exprs) > 0 {
				for i, e := range s.Exprs {
					s.Exprs[i] = *h.expr(&e)
				}
			}
			if h.LocalAssignStmt != nil {
				h.LocalAssignStmt(s)
			}
		case *ast.FuncCallStmt:
			h.expr(&s.Expr)
		case *ast.DoBlockStmt:
			if h.DoBlockStmt != nil {
				h.DoBlockStmt(s)
			} else {
				h.Traverse(s.Chunk)
			}
		case *ast.WhileStmt:
			if ex := h.expr(&s.Condition); ex != nil {
				s.Condition = *ex
			}
			if h.WhileStmt != nil {
				h.WhileStmt(s)
			} else {
				h.Traverse(s.Chunk)
			}
		case *ast.RepeatStmt:
			if ex := h.expr(&s.Condition); ex != nil {
				s.Condition = *ex
			}
			if h.RepeatStmt != nil {
				h.RepeatStmt(s)
			} else {
				h.Traverse(s.Chunk)
			}
		case *ast.FuncDefStmt:
			if s.Name.Func == nil {
				if ex := h.expr(&s.Name.Receiver); ex != nil {
					s.Name.Receiver = *ex
				}
				if ex := h.expr(&[]ast.Expr{s.Func}[0]); ex != nil { // This is scuffed af
					//stmt.Func = *ex
					panic("figure this shit out")
				}
			} else {
				//astmt := &ast.AssignStmt{Lhs: []ast.Expr{s.Name.Func}, Rhs: []ast.Expr{s.Func}}
				//h.compileAssignStmt(astmt)
				panic("figure this shit out")
			}
			if h.FuncDefStmt != nil {
				h.FuncDefStmt(s)
			}
		case *ast.ReturnStmt:
			for i := range s.Exprs {
				if ex := h.expr(&s.Exprs[i]); ex != nil {
					s.Exprs[i] = *ex
				}
			}
			if h.ReturnStmt != nil {
				h.ReturnStmt(s)
			}
		case *ast.IfStmt:
			if ex := h.expr(&s.Condition); ex != nil {
				s.Condition = *ex
			}
			if h.IfStmt != nil {
				if len(s.Else) > 0 {
					h.IfStmt(s)
				} else {
					h.IfStmt(s)
				}
			}
		case *ast.BreakStmt:
			if h.BreakStmt != nil {
				h.BreakStmt()
			}
		case *ast.NumberForStmt:
			if ex := h.expr(&s.Init); ex != nil {
				s.Init = *ex
			}
			if ex := h.expr(&s.Limit); ex != nil {
				s.Limit = *ex
			}
			if s.Step != nil {
				if ex := h.expr(&s.Step); ex != nil {
					s.Step = *ex
				}
			}
			if h.NumberForStmt != nil {
				h.NumberForStmt(s)
			} else {
				h.Traverse(s.Chunk)
			}
		case *ast.GenericForStmt:
			for i := range s.Exprs {
				if ex := h.expr(&s.Exprs[i]); ex != nil {
					s.Exprs[i] = *ex
				}
			}
			if h.GenericForStmt != nil {
				h.GenericForStmt(s)
			} else {
				h.Traverse(s.Chunk)
			}
		case *ast.ContinueStmt:
			if h.ContinueStmt != nil {
				h.ContinueStmt()
			}
		case *ast.LabelStmt:
			if h.LabelStmt != nil {
				h.LabelStmt(s)
			}
		case *ast.GotoStmt:
			if h.GotoStmt != nil {
				h.GotoStmt(s)
			}
		}
	}
}

// Traverse Abstract Syntax Tree
func (h *Handlers) Traverse(ast []ast.Stmt) {
	h.compileStmt(ast)
}
