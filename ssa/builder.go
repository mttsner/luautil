package ssa

import (
	"fmt"

	"github.com/notnoobmaster/luautil/ast"
)

type builder struct{}


func (b *builder) expr(fn *Function, expr ast.Expr) Value {
	switch ex := expr.(type) {
	case *ast.NumberExpr:
		return Const{Value: ex.Value}
	case *ast.NilExpr:
		return Const{Value: nil}
	case *ast.FalseExpr:
		return Const{Value: false}
	case *ast.TrueExpr:
		return Const{Value: true}
	case *ast.IdentExpr:
		return fn.lookup(ex.Value)
	case *ast.Comma3Expr:
		return VarArg{}
	case *ast.StringExpr:
		return Const{Value: ex.Value}
	case *ast.AttrGetExpr:
		return AttrGet{
			Object: b.expr(fn, ex.Object),
			Key:    b.expr(fn, ex.Key),
		}
	case *ast.TableExpr:
		panic("Serialize table")
		return Table{}
	case *ast.ArithmeticOpExpr:
		return Arithmetic{
			Op:  ex.Operator,
			Lhs: b.expr(fn, ex.Lhs),
			Rhs: b.expr(fn, ex.Rhs),
		}
	case *ast.StringConcatOpExpr:
		return Concat{
			Lhs: b.expr(fn, ex.Lhs),
			Rhs: b.expr(fn, ex.Rhs),
		}
	case *ast.RelationalOpExpr:
		return Relation{
			Op:  ex.Operator,
			Lhs: b.expr(fn, ex.Lhs),
			Rhs: b.expr(fn, ex.Rhs),
		}
	case *ast.LogicalOpExpr:
		return Logic{
			Op:  ex.Operator,
			Lhs: b.expr(fn, ex.Lhs),
			Rhs: b.expr(fn, ex.Rhs),
		}
	case *ast.UnaryOpExpr:
		return Unary{
			Op:    ex.Operator,
			Value: b.expr(fn, ex.Expr),
		}
	case *ast.FuncCallExpr:
		call := Call{
			Args: make([]Value, len(ex.Args)),
		}

		if ex.Func != nil {
			call.Func = b.expr(fn, ex.Func)
		} else {
			receiver := b.expr(fn, ex.Receiver)
			// Prepend self
			call.Args = append([]Value{receiver}, call.Args...)
			call.Recv = receiver
			call.Method = ex.Method
		}
		return call
	case *ast.FunctionExpr:
		f := &Function{syntax: ex}
		b.buildFunction(f)
		//return fn.emit(fn)
	}
	panic("unimplemented expression")
}

// buildFunction builds SSA code for the body of function fn.  Idempotent.
func (b *builder) buildFunction(fn *Function) {
	
}

// repeat stmtemits to fn code for the repeat statement s
func (b *builder) repeatStmt(fn *Function, s *ast.RepeatStmt) {
	body := fn.newBasicBlock("repeat.body")
	done := fn.newBasicBlock("repeat.done") // target of 'break'
	loop := fn.newBasicBlock("repeat.loop") // target of 'continue'

	emitJump(fn, body)

	fn.addReturn(b.expr(fn, s.Condition), body, done)

	fn.currentBlock = body
	b.stmtList(fn, s.Chunk)
	emitJump(fn, loop)
	fn.currentBlock = done
}

func (b *builder) whileStmt(fn *Function, s *ast.WhileStmt) {
	body := fn.newBasicBlock("while.body")
	done := fn.newBasicBlock("while.done") // target of 'break'
	loop := fn.newBasicBlock("while.loop") // target of 'continue'

	emitJump(fn, loop)

	fn.addWhile(b.expr(fn, s.Condition), body, done)

	fn.currentBlock = body
	b.stmtList(fn, s.Chunk)
	emitJump(fn, loop)
	fn.currentBlock = done
}

func (b *builder) numberForStmt(fn *Function, s *ast.NumberForStmt) {
	body := fn.newBasicBlock("for.body")
	done := fn.newBasicBlock("for.done") // target of 'break'
	loop := fn.newBasicBlock("for.loop") // target of 'continue'

	emitJump(fn, loop)

	fn.addNumberFor(b, s, body, done)

	fn.currentBlock = body
	b.stmtList(fn, s.Chunk)
	emitJump(fn, loop)
	fn.currentBlock = done
}

func (b *builder) genericForStmt(fn *Function, s *ast.GenericForStmt) {
	body := fn.newBasicBlock("for.body")
	done := fn.newBasicBlock("for.done") // target of 'break'
	loop := fn.newBasicBlock("for.loop") // target of 'continue'

	emitJump(fn, loop)

	locals := make([]Local, len(s.Names))
	values := make([]Value, len(s.Exprs))

	for i, name:= range s.Names {
		locals[i] = f.lookup(name)
	}

	for i, expr := range s.Exprs {
		values[i] = b.expr(f, expr)
	}



	fn.addGenericFor(b, s, body, done)

	fn.currentBlock = body
	b.stmtList(fn, s.Chunk)
	emitJump(fn, loop)
	fn.currentBlock = done
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt) {
	for _, s := range list {
		b.stmt(fn, s)
	}
}

// stmt lowers statement s to SSA form, emitting code to fn.
func (b *builder) stmt(fn *Function, st ast.Stmt) {

	switch s := st.(type) {
	case *ast.AssignStmt:
		if len(s.Lhs) <= len(s.Rhs) { // a, b = 1, 2 or a, b = 1, 2, 3
			for i, ex := range s.Lhs {
				fn.addAssign(b.expr(fn, ex), b.expr(fn, s.Rhs[i]))
			}
		} else { // a, b = 1
			i, l, r := 0, len(s.Lhs), len(s.Rhs)
			for ; i < l; i++ {
				fn.addAssign(b.expr(fn, s.Lhs[i]), b.expr(fn, s.Rhs[i]))
			}
			for ; i < r; i++ {
				fn.addAssign(b.expr(fn, s.Lhs[i]), b.expr(fn, &ast.NilExpr{}))
			}
		}

	case *ast.CompoundAssignStmt:
		if len(s.Lhs) <= len(s.Rhs) { // a, b = 1, 2 or a, b = 1, 2, 3
			for i, ex := range s.Lhs {
				fn.addCompoundAssign(s.Operator, ex, b.expr(fn, s.Rhs[i]))
			}
		} else { // a, b = 1
			i, l, r := 0, len(s.Lhs), len(s.Rhs)
			for ; i < l; i++ {
				fn.addCompoundAssign(s.Operator, s.Lhs[i], b.expr(fn, s.Rhs[i]))
			}
			for ; i < r; i++ {
				fn.addCompoundAssign(s.Operator, s.Lhs[i], b.expr(fn, &ast.NilExpr{}))
			}
		}
	case *ast.LocalAssignStmt:
		switch {
		case len(s.Names) <= len(s.Exprs): // local a, b = 1, 2
			for i, name := range s.Names {
				fn.addLocalAssign(name, b.expr(fn, s.Exprs[i]))
			}
		case len(s.Exprs) == 0: // local a, b
			for _, name := range s.Names {
				fn.addLocalAssign(name, b.expr(fn, &ast.NilExpr{}))
			}
		default: // local a, b = 1
			i, e, n := 0, len(s.Exprs), len(s.Names)
			for ; i < e; i++ {
				fn.addLocalAssign(s.Names[i], b.expr(fn, s.Exprs[i]))
			}
			for ; i < n; i++ {
				fn.addLocalAssign(s.Names[i], b.expr(fn, &ast.NilExpr{}))
			}
		}

	case *ast.FuncCallStmt:
		b.expr(fn, s.Expr)
	case *ast.DoBlockStmt:
	//create new scope somehow
	case *ast.WhileStmt:
		b.whileStmt(fn, s)
	case *ast.RepeatStmt:
		b.repeatStmt(fn, s)
	case *ast.FuncDefStmt:
		//same shit as the expr stuff

	case *ast.ReturnStmt:
		// some some trickery to convert the exprs into useable values
		// something like
		// exprs(s.Exprs)
		//fn.emit(&Return{Results: s.Exprs, pos: s.Return})
		fn.currentBlock = fn.newBasicBlock("unreachable")
	case *ast.IfStmt:
		then := fn.newBasicBlock("if.then")
		done := fn.newBasicBlock("if.done")
		els := done
		if s.Else != nil {
			els = fn.newBasicBlock("if.else")
		}
		fn.addIfStmt(b, s.Condition, then, els)
		fn.currentBlock = then
		b.stmtList(fn, s.Then)
		emitJump(fn, done)

		if s.Else != nil {
			fn.currentBlock = els
			b.stmtList(fn, s.Else)
			emitJump(fn, done)
		}

		fn.currentBlock = done

	case *ast.BreakStmt:
		emitJump(fn, fn.breakBlock)
		fn.currentBlock = fn.newBasicBlock("unreachable")
	case *ast.ContinueStmt:
		emitJump(fn, fn.continueBlock)
		fn.currentBlock = fn.newBasicBlock("unreachable")
	case *ast.NumberForStmt:
		b.numberForStmt(fn, s)
	case *ast.GenericForStmt:
		b.genericForStmt(fn, s)
	default:
		panic(fmt.Sprintf("unexpected statement kind: %T", s))
	}
}