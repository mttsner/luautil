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
			Expr:    expr(v.Value),
		}
	default:
		panic("unimplemented"+ fmt.Sprint(v))
	}
}

func (b *BasicBlock) ToAst(dom domFrontier) (chunk ast.Chunk) {
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
		case *If:
			tFront := dom[b.Succs[0].Index]
			fFront := dom[b.Succs[1].Index]
			if len(tFront) == len(fFront) &&
				tFront[0].Index == fFront[0].Index {
				panic("if then end")
			} else {
				panic("if then else end")
			}
		default:
			panic("reached")
		}
	}

	return
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	root := f.Blocks[0]
	return root.ToAst()
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