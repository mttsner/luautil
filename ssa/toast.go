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

func (f *Function) stmt(i Instruction) ast.Stmt {
	switch i := i.(type) {
	case *If:
		tFront := f.DomFrontier[i.Block().Succs[0].Index]
		fFront := f.DomFrontier[i.Block().Succs[1].Index]
		switch {
		case len(i.Block().Preds) > 1 && i.Block().Dominates(i.Block().Preds[1]):
			return &ast.WhileStmt{
				Condition: expr(i.Cond),
				Chunk:     f.chunk(i.Block().Succs[0]),
			}
		case len(tFront) == len(fFront) && tFront[0].Index == fFront[0].Index:
			return &ast.IfStmt{
				Condition: expr(i.Cond),
				Then:      f.chunk(i.Block().Succs[0]),
				Else:      f.chunk(i.Block().Succs[1]),
			}
		default:
			return &ast.IfStmt{
				Condition: expr(i.Cond),
				Then:      f.chunk(i.Block().Succs[0]),
			}
		}
	default:
		panic("unimplemented")
	}
}

func (f *Function) chunk(b *BasicBlock) (chunk ast.Chunk) {
	for _, inst := range b.Instrs {
		chunk = append(chunk, f.stmt(inst))
	}
	return append(chunk, chunk(next_block))
}

func (c converter) stmt(b *BasicBlock, dom DomFrontier) (chunk ast.Chunk) {
	for _, inst := range b.Instrs {
		switch i := inst.(type) {
		case *Assign:
			/*
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
			*/
		case *If:
			// TODO: Add repeat support
			// I have no idea how how I wrote this code but it works

		case *Jump:
			return
		default:
			panic("reached")
		}
	}
	if len(b.Succs) > 0 {
		return append(chunk, b.Succs[0].stmt(dom)...)
	}
	return
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	buildDomTree(f)
	BuildDomFrontier(f)
	return f.chunk(f.Blocks[0])
}

type converter struct {
	*ast.Chunk
}

/*
if cond then
	...
end

if cond then
	...
else
	...
end

while cond do
	...
end

repeat
	...
until cond end

*/
