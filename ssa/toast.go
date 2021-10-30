package ssa

import (
	"fmt"

	"github.com/notnoobmaster/luautil/ast"
)


func expr(v Value) ast.Expr {
	switch v := v.(type) {
	case Const: // This is fucking ugly
		switch c := v.Value.(type) {
		case nil:
			return &ast.NilExpr{}
		case bool:
			if c {
				return &ast.TrueExpr{}
			} else {
				return &ast.FalseExpr{}
			}
		case float64:
			return &ast.NumberExpr{Value: c}
		case string:
			return &ast.StringExpr{Value: c}
		default:
			panic("unimplemented constant type")
		}
	case *Local:
		return &ast.IdentExpr{Value: v.Comment}
	case VarArg:
		return &ast.Comma3Expr{}
	case AttrGet:
		return &ast.AttrGetExpr{
			Object: expr(v.Object),
			Key:    expr(v.Key),
		}
	case Table:
		panic("implement")
	case Arithmetic:
		return &ast.ArithmeticOpExpr{
			Operator: v.Op,

			Lhs: expr(v.Lhs),
			Rhs: expr(v.Rhs),
		}
	case Concat:
		return &ast.StringConcatOpExpr{
			Lhs: expr(v.Lhs),
			Rhs: expr(v.Rhs),
		}
	case Relation:
		return &ast.RelationalOpExpr{
			Operator: v.Op,
			
			Lhs: expr(v.Lhs),
			Rhs: expr(v.Rhs),
		}
	case Logic:
		return &ast.LogicalOpExpr{
			Operator: v.Op,

			Lhs: expr(v.Lhs),
			Rhs: expr(v.Rhs),
		}
	case Unary:
		return &ast.UnaryOpExpr{
			Operator: v.Op,
			Expr:    expr(v.Value),
		}
	default:
		panic("unimplemented"+ fmt.Sprint(v))
	}
}

func (b *BasicBlock) ToAst() (chunk ast.Chunk) {
	for _, inst := range b.Instrs {
		switch i := inst.(type) {	
		case *Assign:	
			if l, ok := i.Lhs.(*Local); ok && !l.declared {
				chunk = append(chunk, &ast.LocalAssignStmt{
					Names: []string{l.Comment},
					Exprs: []ast.Expr{expr(i.Rhs)},
				})
				l.declared = true
			} else {
				chunk = append(chunk, &ast.AssignStmt{
					Lhs: []ast.Expr{expr(i.Lhs)},
					Rhs: []ast.Expr{expr(i.Rhs)},
				})
			}
		default:
			panic("reached")
		}
	}

	return
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	// prob should work downwards from root and not loop
	for _, b := range f.Blocks {
		chunk = append(chunk, b.ToAst()...)
		if len(b.Succs) == 2 {
			// might be if statement
		}
	}
	return 
}