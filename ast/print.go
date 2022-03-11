package ast

import (
	"strconv"
	"strings"

	"github.com/notnoobmaster/luautil"
)

func (c Chunk) String() string { 
	s := &builder{
		Str:    &strings.Builder{},
		Indent: -1, // Accounting for the fact that each chunk call increments Indent by one
	}
	s.chunk(c)
	return s.Str.String()
}

// Expressions

func (v *NilExpr) String() string { return "nil" }
func (v *TrueExpr) String() string { return "true" }
func (v *FalseExpr) String() string { return "false" }
func (v *IdentExpr) String() string { return v.Value }
func (v *Comma3Expr) String() string { return "..."}

func (v *NumberExpr) String() string {
	return strconv.FormatFloat(v.Value, 'f', -1, 64)
}

func (v *StringExpr) String() string {
	return luautil.Quote(v.Value)
}

// We pass the value to b.expr because we need to know the indentation level and carry some state.

func (e *AttrGetExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *TableExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *FuncCallExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *LogicalOpExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *RelationalOpExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *StringConcatOpExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *ArithmeticOpExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *UnaryOpExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

func (e *FunctionExpr) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.expr(e, data{})
	return b.Str.String()
}

// Statements

func (s *AssignStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *CompoundAssignStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *LocalAssignStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *FuncCallStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *DoBlockStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *WhileStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *RepeatStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *IfStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *NumberForStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *GenericForStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *LocalFunctionStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *FunctionStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *ReturnStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *BreakStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *ContinueStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *LabelStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}

func (s *GotoStmt) String() string {
	b := &builder{&strings.Builder{}, 0}
	b.stmt(s)
	return b.Str.String()
}