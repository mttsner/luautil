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
	case Call:
		return &ast.FuncCallExpr{
			Func:     expr(v.Func),
			Receiver: expr(v.Recv),
			Method:   v.Method,
			Args:     exprs(v.Args),
		}
	case *Global:
		return &ast.IdentExpr{Value: v.Comment}
	case nil:
		return nil
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
					Func:     expr(i.Func),
					Receiver: expr(i.Recv),
					Method:   i.Method,
					Args:     exprs(i.Args),
				},
			})
		case *Jump:
			if c.fn.breakBlock != nil && i.Target.Index == c.fn.breakBlock.Index { // Break
				chunk = append(chunk, &ast.BreakStmt{})
			}

		case *If, *GenericFor, *NumberFor:
			panic("shouldn't reach controlflow related instructions")
		}
	}
	return
}

func (c *converter) block(b *BasicBlock, ignoreRepeat bool) ast.Chunk {
	switch {
	case b.isGenericForLoop():
		loop := b
		body := b.Succs[0]
		done := b.Succs[1]

		c.fn.breakBlock = done
		c.fn.continueBlock = loop

		instr := loop.Instrs[0].(*GenericFor)
		names := make([]string, len(instr.Locals))
		for i, l := range instr.Locals {
			names[i] = l.String()
		}
		return ast.Chunk{&ast.GenericForStmt{
			Names: names,
			Exprs: exprs(instr.Values),
			Chunk: c.chunk(frame{
				start: body.Index,
				end:   done.Index,
			}),
		}}
	case b.isNumberForLoop():
		loop := b
		body := b.Succs[0]
		done := b.Succs[1]

		c.fn.breakBlock = done
		c.fn.continueBlock = loop

		instr := loop.Instrs[0].(*NumberFor)
		return ast.Chunk{&ast.NumberForStmt{
			Name:  instr.Local.String(),
			Init:  expr(instr.Init),
			Limit: expr(instr.Limit),
			Step:  expr(instr.Step),
			Chunk: c.chunk(frame{
				start: body.Index,
				end:   done.Index,
			}),
		}}
	case b.isWhileLoop(c.domFrontier):
		loop := b // target of 'continue'
		body := b.Succs[0]
		done := b.Succs[1] // target of 'break'

		c.fn.breakBlock = done
		c.fn.continueBlock = loop

		instr := loop.Instrs[0].(*If)
		return ast.Chunk{&ast.WhileStmt{
			Condition: expr(instr.Cond),
			Chunk: c.chunk(frame{
				start: body.Index,
				end:   done.Index,
			}),
		}}
	case b.isIfElse(c.domFrontier):
		lastI := len(b.Instrs) - 1
		instr := b.Instrs[lastI].(*If)
		stmts := c.stmts(b.Instrs[:lastI])
		stmt := &ast.IfStmt{
			Condition: expr(instr.Cond),
			Then: c.chunk(frame{
				start: b.Index,
				end:   b.Succs[1].Index,
			}),
			Else: c.chunk(frame{
				start: b.Succs[1].Index,
				end:   c.domFrontier[b.Succs[0].Index][0].Index,
			}),
		}
		return append(stmts, stmt)
	case !ignoreRepeat && b.isRepeat():
		loop := b.Preds[1]    // target of 'continue'
		done := loop.Succs[1] // target of 'break'

		c.fn.breakBlock = done
		c.fn.continueBlock = loop

		instr := b.Preds[1].Instrs[0].(*If)
		stmts := c.block(b, true)
		stmt := &ast.RepeatStmt{
			Condition: expr(instr.Cond),
			Chunk: append(stmts, c.chunk(frame{
				start: b.Index,
				end:   done.Index,
			})...),
		}
		c.skipBlock() // skip if stmt
		return ast.Chunk{stmt}
	case b.isIf(c.domFrontier):
		lastI := len(b.Instrs) - 1
		instr := b.Instrs[lastI].(*If)
		stmts := c.stmts(b.Instrs[:lastI])
		stmt := &ast.IfStmt{
			Condition: expr(instr.Cond),
			Then: c.chunk(frame{
				start: b.Index,
				end:   b.Succs[1].Index,
			}),
		}
		return append(stmts, stmt)
	default:
		return c.stmts(b.Instrs)
	}
}

type converter struct {
	domFrontier  DomFrontier
	fn           *Function
	idx          int
}

type frame struct {
	start, end int
}

func (f frame) size() int {
	return f.end - f.start
}

func (c *converter) nextBlock(f frame) *BasicBlock {
	if f.start <= c.idx && c.idx < f.end {
		c.idx++
		return c.fn.Blocks[c.idx-1]
	}
	return nil
}

func (c *converter) skipBlock() {
	c.idx++
}

func (c *converter) chunk(f frame) ast.Chunk {
	chunk := make(ast.Chunk, 0, f.size()*10)
	for b := c.nextBlock(f); b != nil; b = c.nextBlock(f) {
		chunk = append(chunk, c.block(b, false)...)
	}
	return chunk
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	buildDomTree(f)
	BuildDomFrontier(f)
	c := converter{
		domFrontier: f.DomFrontier,
		fn:          f,
	}
	return c.chunk(frame{
		start: c.idx,
		end:   len(f.Blocks),
	})
}
