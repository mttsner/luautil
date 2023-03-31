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

func (fn *Function) stmts(b *BasicBlock) (chunk ast.Chunk) {
	for _, instr := range b.Instrs {
		switch i := instr.(type) {
		case *Assign:
			if l, ok := i.Lhs[0].(*Local); ok && !l.declared {
				chunk = append(chunk, &ast.LocalAssignStmt{
					Names: []string{l.Comment},
					Exprs: []ast.Expr{expr(i.Rhs[0])},
				})
				l.declared = true
			} else {
				chunk = append(chunk, &ast.AssignStmt{
					Lhs: []ast.Expr{expr(i.Lhs[0])},
					Rhs: []ast.Expr{expr(i.Rhs[0])},
				})
			}
		case *If:
		case *Jump:

		}
	}
}

func isWhileLoop(b *BasicBlock) bool {
	return len(b.Preds) > 1 && b.Dominates(b.Preds[1])
}

func (fn *Function) toWhileLoop(b *BasicBlock) ast.Chunk {
	switch last := b.Instrs[len(b.Instrs)-1].(type) {
	case *If:
		fn.breakBlock = b.Succs[1]
		return append(fn.stmts(b), &ast.WhileStmt{
			Condition: expr(last.Cond),
			Chunk: fn.block(b.Succs[0]),
		})
	case *Jump:
		panic("possible while true do loop")
	default:
		panic("can't decompile possible while loop")
	}
}

func isIfElse(tFront, fFront []*BasicBlock) bool {
	return len(tFront) == len(fFront) && tFront[0].Index == fFront[0].Index
}

func isRepeat(b *BasicBlock) bool {
	return len(b.Preds) == 2 && b.Dominates(b.Preds[1]) && len(b.Succs) == 1 
}

func isIf(tFront []*BasicBlock, b *BasicBlock) bool {
	return tFront[0] == b.Succs[1] 
}

func (fn *Function) block(b *BasicBlock) (chunk ast.Chunk) {
	switch {
	case isRepeat(b):
		panic("repeat until")
	case len(b.Succs) != 2:
		panic("possible jump")
	case isIf(b):
		panic("if")
	case isIfElse(tFront, fFront):
		panic("if-else")
	case isWhileLoop(b):
		panic("while loop")
	default:
		panic("can't solve cf")
	}
}

func (f *Function) Chunk() {
	buildDomTree(f)
	BuildDomFrontier(f)
	c := ast.Chunk{}
	for _, b := range f.Blocks {
		c = append(c, f.block(b)...)
	}
}