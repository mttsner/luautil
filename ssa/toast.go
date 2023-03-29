package ssa

import (
	"fmt"

	"github.com/notnoobmaster/luautil/ast"
)

func expr(v Value) ast.Expr {
	switch v := v.(type) {
	case Nil:
		return &ast.NilExpr{}
	case True:
		return &ast.TrueExpr{}
	case False:
		return &ast.FalseExpr{}
	case VarArg:
		return &ast.Comma3Expr{}
	case Number:
		return &ast.NumberExpr{Value: v.Value}
	case String:
		return &ast.StringExpr{Value: v.Value}
	case *Local:
		return &ast.IdentExpr{Value: v.Comment}
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
			Expr:     expr(v.Value),
		}
	default:
		panic("unimplemented" + fmt.Sprint(v))
	}
}

func (f *Function) block(b *BasicBlock, dom DomFrontier) {
	if len(b.Preds) == 2 && b.Dominates(b.Preds[1]) && len(b.Succs) == 1 {
		panic("repeat until")
	}

	if len(b.Succs) != 2 {
		return 
	}
	tFront := dom[b.Succs[0].Index]
	fFront := dom[b.Succs[1].Index]

	if tFront[0] == b.Succs[1] {
		panic("if-then")
	}

	if len(tFront) == len(fFront) && tFront[0].Index == fFront[0].Index {
		panic("if-else")
	}

	if len(b.Preds) > 1 && b.Dominates(b.Preds[1]) {
		panic("while")
	}
}

func (f *Function) Chunk() {
	buildDomTree(f)
	BuildDomFrontier(f)
	for _, b := range f.Blocks {
		f.block(b, f.DomFrontier)
	}
}