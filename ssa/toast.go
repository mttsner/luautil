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

func (c *converter) stmts(instrs []Instruction) (chunk ast.Chunk) {
	for _, instr := range instrs {
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
			panic("should never reach this")
		case *Jump:

		}
	}
	return
}

/*
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
*/
func (c *converter) block(b *BasicBlock) {
	switch {
	case b.isIf(c.domFrontier):
		inst := b.Instrs[len(b.Instrs)-1].(*If)
		chunk := c.stmts(b.Instrs[:len(b.Instrs)-1])
		stmt := &ast.IfStmt{
			Condition: expr(inst.Cond),
			Then:      ast.Chunk{},
		}
		c.addChunk(chunk)
		c.addStmt(stmt)
		c.newScope(b.Succs[1].Index, &stmt.Then)
	case b.isIfElse(c.domFrontier):
		inst := b.Instrs[len(b.Instrs)-1].(*If)
		chunk := c.stmts(b.Instrs[:len(b.Instrs)-1])
		stmt := &ast.IfStmt{
			Condition: expr(inst.Cond),
			Then:      ast.Chunk{},
			Else:      ast.Chunk{},
		}
		c.addChunk(chunk)
		c.addStmt(stmt)
		c.newScope(c.domFrontier[b.Succs[0].Index][0].Index, &stmt.Else)
		c.newScope(b.Succs[1].Index, &stmt.Then)
	/*
		case b.isWhileLoop():
			stmt := b.Instrs[len(b.Instrs)-1].(*If)
			chunk = append(fn.stmts(b.Instrs[:len(b.Instrs)-1]), &ast.WhileStmt{
				Condition: expr(stmt.Cond),
				Chunk:      fn.block(b.Succs[0]),
			})
			return append(chunk, fn.block(b.Succs[1])...)
		case b.isRepeat():
			panic("repeat until")*/
	default:
		c.addChunk(c.stmts(b.Instrs))
	}
}

type converter struct {
	domFrontier  DomFrontier
	currentScope *ast.Chunk
	scopes       map[int]*ast.Chunk
}

func (c *converter) addChunk(chunk ast.Chunk) {
	(*c.currentScope) = append((*c.currentScope), chunk...)
}

func (c *converter) addStmt(stmt ast.Stmt) {
	(*c.currentScope) = append((*c.currentScope), stmt)
}

func (c *converter) newScope(idx int, newScope *ast.Chunk) {
	c.scopes[idx] = c.currentScope
	c.currentScope = newScope
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	buildDomTree(f)
	BuildDomFrontier(f)
	c := converter{
		currentScope: &chunk,
		scopes:       map[int]*ast.Chunk{},
		domFrontier:  f.DomFrontier,
	}

	for _, b := range f.Blocks {
		if scope, ok := c.scopes[b.Index]; ok {
			c.currentScope = scope
		}
		c.block(b)
	}
	return
}
