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

func exprs(vals []Value) (exprs []ast.Expr) {
	for _, v := range vals {
		exprs = append(exprs, expr(v))
	}
	return
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
		case *Return:
			chunk = append(chunk, &ast.ReturnStmt{
				Exprs: exprs(i.Values),
			})
		case *Call:
			chunk = append(chunk, &ast.FuncCallStmt{
				Expr: &ast.FuncCallExpr{
					Func:      expr(i.Func),
					Receiver:  expr(i.Recv),
					Method:    i.Method,
					Args:      exprs(i.Args),
				},
			})
		case *If:
			// do nothing
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
	case b.isWhileLoop():
		inst := b.Instrs[len(b.Instrs)-1].(*If)
		chunk := c.stmts(b.Instrs[:len(b.Instrs)-1])
		stmt := &ast.WhileStmt{
			Condition: expr(inst.Cond),
			Chunk:     ast.Chunk{},
		}
		c.addChunk(chunk)
		c.addStmt(stmt)
		c.newScope(b.Succs[1].Index, &stmt.Chunk)
	case b.isRepeat():
		inst := b.Preds[1].Instrs[0].(*If)
		chunk := c.stmts(b.Instrs)
		stmt := &ast.RepeatStmt{
			Condition: expr(inst.Cond),
			Chunk: ast.Chunk{},
		}
		c.addStmt(stmt)
		c.newScope(b.Preds[1].Index, &stmt.Chunk)
		c.addChunk(chunk)
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
	// need to optimize the ssa for buildDomFrontier to work
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
