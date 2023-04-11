package ssa

import (
	"fmt"

	"github.com/notnoobmaster/luautil/ast"
)

type builder struct{}

func (b *builder) expr(fn *Function, expr ast.Expr) Value {
	switch ex := expr.(type) {
	case *ast.NilExpr:
		return Nil{}
	case *ast.FalseExpr:
		return False{}
	case *ast.TrueExpr:
		return True{}
	case *ast.NumberExpr:
		return Number{ex.Value}
	case *ast.StringExpr:
		return String{ex.Value}
	case *ast.IdentExpr:
		return fn.lookup(ex.Value)
	case *ast.Comma3Expr:
		return VarArg{}
	case *ast.AttrGetExpr:
		return AttrGet{
			Object: b.expr(fn, ex.Object),
			Key:    b.expr(fn, ex.Key),
		}
	case *ast.TableExpr:
		tbl := Table{}
		for _, fi := range ex.Fields {
			field := &Field{
				Value: b.expr(fn, fi.Value),
			}
			if fi.Key != nil {
				field.Key = b.expr(fn, fi.Key)
			}
			tbl.Fields = append(tbl.Fields, field)
		}
		return tbl
	case *ast.ArithmeticOpExpr:
		return Arithmetic{
			Op: ex.Operator,

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
		return b.funcCallExpr(fn, ex)
	case *ast.FunctionExpr:
		return b.functionExpr(fn, ex)
	default:
		panic(fmt.Sprintf("unexpected expr type: %T", ex))
	}
}

func (b *builder) funcCallExpr(fn *Function, ex *ast.FuncCallExpr) Call {
	call := Call{
		Args: make([]Value, len(ex.Args)),
	}

	for i, arg := range ex.Args {
		call.Args[i] = b.expr(fn, arg)
	}

	if ex.Func != nil { // hoge.func()
		call.Func = b.expr(fn, ex.Func)
	} else { // hoge:func()
		call.Recv = b.expr(fn, ex.Receiver)
		call.Method = ex.Method
	}
	return call
}

func (b *builder) functionExpr(fn *Function, ex *ast.FunctionExpr) Value {
	f := fn.addFunction(ex)
	b.buildFunction(f)
	return f
}

// buildFunction builds SSA code for the body of function fn.  Idempotent.
func (b *builder) buildFunction(fn *Function) {
	fn.StartBody()
	f := fn.syntax
	for _, name := range f.ParList.Names {
		fn.addParam(name)
	}
	fn.VarArg = f.ParList.HasVargs
	b.chunk(fn, f.Chunk)
	fn.finishBody()
}

func Build(chunk ast.Chunk) *Function {
	var b builder
	fn := &Function{
		syntax: &ast.FunctionExpr{
			Chunk:   chunk,
			ParList: &ast.ParList{HasVargs: true},
		},
		Name: "main",
	}
	b.buildFunction(fn)
	return fn
}

// repeat stmtemits to fn code for the repeat statement s
func (b *builder) repeatStmt(fn *Function, s *ast.RepeatStmt) {
	loop := fn.CreateBasicBlock("repeat.loop") // target of 'continue'
	body := fn.NewBasicBlock("repeat.body")
	done := fn.CreateBasicBlock("repeat.done") // target of 'break'

	breakBlock := fn.breakBlock
	continueBlock := fn.continueBlock

	fn.breakBlock = done
	fn.continueBlock = loop

	addEdge(fn.currentBlock, body)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	addEdge(fn.currentBlock, loop)

	fn.currentBlock = loop
	fn.emitIf(b.expr(fn, s.Condition), body, done)
	fn.currentBlock = done

	fn.breakBlock = breakBlock
	fn.continueBlock = continueBlock

	fn.AddBasicBlock(loop)
	fn.AddBasicBlock(done)
}

func (b *builder) whileStmt(fn *Function, s *ast.WhileStmt) {
	loop := fn.NewBasicBlock("while.loop") // target of 'continue'
	body := fn.NewBasicBlock("while.body")
	done := fn.CreateBasicBlock("while.done") // target of 'break'

	breakBlock := fn.breakBlock
	continueBlock := fn.continueBlock

	fn.breakBlock = done
	fn.continueBlock = loop

	addEdge(fn.currentBlock, loop)
	fn.currentBlock = loop
	fn.emitIf(b.expr(fn, s.Condition), body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	fn.emitJump(loop)

	fn.AddBasicBlock(done)

	fn.currentBlock = done
	fn.breakBlock = breakBlock
	fn.continueBlock = continueBlock
}

func (b *builder) numberForStmt(fn *Function, s *ast.NumberForStmt) {
	loop := fn.NewBasicBlock("for.loop") // target of 'continue'
	body := fn.NewBasicBlock("for.body")
	done := fn.CreateBasicBlock("for.done") // target of 'break'

	local := fn.addLocal(s.Name)

	limit := b.expr(fn, s.Limit)
	init := b.expr(fn, s.Init)
	step := b.expr(fn, s.Step)

	breakBlock := fn.breakBlock
	continueBlock := fn.continueBlock

	fn.breakBlock = done
	fn.continueBlock = loop

	addEdge(fn.currentBlock, loop)
	fn.currentBlock = loop
	fn.emitNumberFor(local, init, limit, step, body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	addEdge(fn.currentBlock, loop)

	fn.AddBasicBlock(done)

	fn.currentBlock = done
	fn.breakBlock = breakBlock
	fn.continueBlock = continueBlock
}

func (b *builder) genericForStmt(fn *Function, s *ast.GenericForStmt) {
	loop := fn.NewBasicBlock("for.loop") // target of 'continue'
	body := fn.NewBasicBlock("for.body")
	done := fn.CreateBasicBlock("for.done") // target of 'break'

	locals := make([]Value, len(s.Names))
	values := make([]Value, len(s.Exprs))

	for i, name := range s.Names {
		locals[i] = fn.addLocal(name)
	}

	for i, expr := range s.Exprs {
		values[i] = b.expr(fn, expr)
	}

	breakBlock := fn.breakBlock
	continueBlock := fn.continueBlock

	fn.breakBlock = done
	fn.continueBlock = loop

	addEdge(fn.currentBlock, loop)
	fn.currentBlock = loop
	fn.emitGenericFor(locals, values, body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	addEdge(fn.currentBlock, loop)

	fn.AddBasicBlock(done)

	fn.currentBlock = done
	fn.breakBlock = breakBlock
	fn.continueBlock = continueBlock
}

// chunk emits to fn code for all statements in list.
func (b *builder) chunk(fn *Function, list ast.Chunk) {
	old := fn.newScope()
	for _, s := range list {
		b.stmt(fn, s)
	}
	fn.currentScope = old
}

// stmt lowers statement s to SSA form, emitting code to fn.
func (b *builder) stmt(fn *Function, st ast.Stmt) {
	switch s := st.(type) {
	case *ast.AssignStmt:
		l, r := len(s.Lhs), len(s.Rhs)
		lhs, rhs := make([]Value, l), []Value{}

		for i, l := range s.Lhs {
			lhs[i] = b.expr(fn, l)
			if i >= r {
				rhs = append(rhs, Nil{})
			} else {
				rhs = append(rhs, b.expr(fn, s.Rhs[i]))
			}
		}

		fn.EmitMultiAssign(lhs, rhs)
	case *ast.CompoundAssignStmt:
		l, r := len(s.Lhs), len(s.Rhs)
		lhs, rhs := make([]Value, l), []Value{}

		for i, l := range s.Lhs {
			lhs[i] = b.expr(fn, l)
			if i >= r {
				rhs = append(rhs, Nil{})
			} else {
				rhs = append(rhs, b.expr(fn, s.Rhs[i]))
			}
		}

		fn.emitCompoundAssign(s.Operator, lhs, rhs)
	case *ast.LocalAssignStmt:
		n, e := len(s.Names), len(s.Exprs)
		// Locals defined in the beginning of a function without a value are ignored and set to nil.
		if len(fn.Blocks) == 1 && len(fn.Blocks[0].Instrs) == 0 && e == 0 {
			for _, name := range s.Names {
				fn.addLocal(name)
			}
			break
		}

		values := []Value{}

		for i := 0; i < e; i++ {
			if i >= n {
				values = append(values, Nil{})
			} else {
				values = append(values, b.expr(fn, s.Exprs[i]))
			}
		}
		fn.emitLocalAssign(s.Names, values)
	case *ast.FuncCallStmt:
		call := b.funcCallExpr(fn, s.Expr.(*ast.FuncCallExpr))
		fn.Emit(&call)
	case *ast.DoBlockStmt:
		b.chunk(fn, s.Chunk)
	case *ast.WhileStmt:
		b.whileStmt(fn, s)
	case *ast.RepeatStmt:
		b.repeatStmt(fn, s)
	case *ast.LocalFunctionStmt:
		f := fn.addFunction(s.Func)
		f.Name = s.Name
		fn.emitLocalAssign([]string{s.Name}, []Value{f})
		b.buildFunction(f)
	case *ast.FunctionStmt:
		var lhs Value
		f := fn.addFunction(s.Func)
		if s.Name.Func != nil {
			lhs = b.expr(fn, s.Name.Func)
			switch e := s.Name.Func.(type) {
			case *ast.IdentExpr: // function func()
				f.Name = e.Value
			case *ast.AttrGetExpr: // function hoge.func()
				f.Name = e.Key.(*ast.StringExpr).Value
			}
		} else { // function hoge:func(). We need to prepend self to args and convert the recv and method fields to recv.method .
			lhs = AttrGet{b.expr(fn, s.Name.Receiver), String{s.Name.Method}}
			f.Name = s.Name.Method
			f.addParam("self")
		}
		b.buildFunction(f)
		fn.EmitAssign(lhs, f)
	case *ast.ReturnStmt:
		values := make([]Value, len(s.Exprs))
		for i, expr := range s.Exprs {
			values[i] = b.expr(fn, expr)
		}

		fn.Emit(&Return{
			Values: values,
		})
	case *ast.IfStmt:
		base := fn.currentBlock
		fn.Emit(&If{Cond: b.expr(fn, s.Condition)})

		then := fn.NewBasicBlock("if.then")
		fn.currentBlock = then
		b.chunk(fn, s.Then)
		addEdge(base, then)
		then = fn.currentBlock

		if s.Else != nil {
			els := fn.NewBasicBlock("if.else")
			fn.currentBlock = els
			b.chunk(fn, s.Else)
			addEdge(base, els)
			els = fn.currentBlock

			done := fn.NewBasicBlock("if.done")
			fn.currentBlock = then
			fn.emitJump(done)

			addEdge(then, done)
			addEdge(els, done)
			fn.currentBlock = done
		} else {
			done := fn.NewBasicBlock("if.done")
			addEdge(then, done)
			addEdge(base, done)
			fn.currentBlock = done
		}
	case *ast.BreakStmt:
		block := fn.NewBasicBlock("unreachable")
		AddUnreachableEdge(fn.currentBlock, block)
		fn.emitJump(fn.breakBlock)
		fn.currentBlock = block

	case *ast.ContinueStmt:
		fn.emitJump(fn.continueBlock)
	case *ast.NumberForStmt:
		b.numberForStmt(fn, s)
	case *ast.GenericForStmt:
		b.genericForStmt(fn, s)
	default:
		panic(fmt.Sprintf("unexpected statement kind: %T", s))
	}
}
