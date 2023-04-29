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
		return &ast.IdentExpr{Value: v.Name()}
	case AttrGet:
		return &ast.AttrGetExpr{
			Object: expr(v.Object),
			Key:    expr(v.Key),
		}
	case *Table:
		tbl := &ast.TableExpr{}
		for _, fi := range v.Fields {
			field := &ast.Field{
				Value: expr(fi.Value),
			}
			if fi.Key != nil {
				field.Key = expr(fi.Key)
			}
			tbl.Fields = append(tbl.Fields, field)
		}
		return tbl
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
	case *Call:
		return &ast.FuncCallExpr{
			Func:     expr(v.Func),
			Receiver: expr(v.Recv),
			Method:   v.Method,
			Args:     exprs(v.Args),
		}
	case *Global:
		if v, ok := v.Value.(String); ok {
			return &ast.IdentExpr{Value: v.Value}
		}
		panic("Invalid global")
	case nil:
		return nil
	case *Function:
		expr := &ast.FunctionExpr{
			ParList: &ast.ParList{
				HasVargs: v.VarArg,
				Names:    make([]string, len(v.Params)),
			},
			Chunk: v.Chunk(),
		}
		for i, l := range v.Params {
			expr.ParList.Names[i] = l.String()
		}
		return expr
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

func (c *converter) stmt(instr Instruction) ast.Stmt {
	switch i := instr.(type) {
	case *Assign:
		if len(i.Lhs) == 0 || len(i.Rhs) == 0 {
			panic("invalid assign instruction")
		}

		l, okl := i.Lhs[0].(*Local)
		f, okf := i.Rhs[0].(*Function)
		// Very funky code
		switch {
		case !(okl && !l.declared):
			return &ast.AssignStmt{
				Lhs: exprs(i.Lhs),
				Rhs: exprs(i.Rhs),
			}
		case okf &&
			len(i.Lhs) == 1 && len(i.Rhs) == 1 &&
			len(f.UpValues) > 0:
			for _, up := range f.UpValues {
				if up == l {
					return &ast.LocalFunctionStmt{
						Name: l.Name(),
						Func: expr(f).(*ast.FunctionExpr),
					}
				}
			}
			fallthrough
		default:
			names := make([]string, len(i.Lhs))
			for i, l := range i.Lhs {
				if l, ok := l.(*Local); ok && !l.declared {
					names[i] = l.Name()
					l.declared = true
				}
			}
			l.declared = true

			return &ast.LocalAssignStmt{
				Names: names,
				Exprs: exprs(i.Rhs),
			}
		}
	case *Return:
		return &ast.ReturnStmt{
			Exprs: exprs(i.Values),
		}
	case *Call:
		return &ast.FuncCallStmt{
			Expr: &ast.FuncCallExpr{
				Func:     expr(i.Func),
				Receiver: expr(i.Recv),
				Method:   i.Method,
				Args:     exprs(i.Args),
			},
		}
	case *If, *Jump, *GenericFor, *NumberFor:
		panic("shouldn't reach controlflow related instructions")
	default:
		panic("unhandled ssa instruction")
	}
}

func (c *converter) stmts(instrs []Instruction) (chunk ast.Chunk) {
	for _, instr := range instrs {
		chunk = append(chunk, c.stmt(instr))
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

		// Remove jump back instruction
		lBlock := c.fn.Blocks[done.Index-1]
		lBlock.Instrs = lBlock.Instrs[:len(lBlock.Instrs)-1]

		instr := loop.Instrs[0].(*If)
		return ast.Chunk{&ast.WhileStmt{
			Condition: expr(instr.Cond),
			Chunk: c.chunk(frame{
				start: body.Index,
				end:   done.Index,
			}),
		}}
	case b.isIfElse(c.domFrontier):
		then := b.Succs[0]
		els := b.Succs[1]
		done := c.domFrontier[b.Succs[0].Index][0]


		lastI := len(b.Instrs) - 1
		instr := b.Instrs[lastI].(*If)
		stmts := c.stmts(b.Instrs[:lastI])

		// Remove jump to done instruction
		lThen := c.fn.Blocks[els.Index-1]
		lThen.Instrs = lThen.Instrs[:len(lThen.Instrs)-1]

		stmt := &ast.IfStmt{
			Condition: expr(instr.Cond),
			Then: c.chunk(frame{
				start: then.Index,
				end:   els.Index,
			}),
			Else: c.chunk(frame{
				start: els.Index,
				end:   done.Index,
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
	case b.isGoto():
		lastI := len(b.Instrs) - 1
		stmts := c.stmts(b.Instrs[:lastI])

		if c.fn.breakBlock != nil && b.Succs[0].Index == c.fn.breakBlock.Index { // Break
			return append(stmts, &ast.BreakStmt{})
		}
		return append(stmts, &ast.GotoStmt{})
	default:
		return c.stmts(b.Instrs)
	}
}

type converter struct {
	domFrontier DomFrontier
	fn          *Function
	idx         int
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

func (c *converter) chunk(f frame) (chunk ast.Chunk) {
	for b := c.nextBlock(f); b != nil; b = c.nextBlock(f) {
		chunk = append(chunk, c.block(b, false)...)
	}
	return
}

func (f *Function) Chunk() (chunk ast.Chunk) {
	if len(f.Blocks) == 0 {
		return
	}
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
