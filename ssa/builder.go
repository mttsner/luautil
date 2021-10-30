package ssa

import (
	"fmt"

	"github.com/notnoobmaster/luautil/ast"
)

type builder struct {
	version int
}

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
	loop := fn.NewBasicBlock("repeat.loop") // target of 'continue'
	body := fn.NewBasicBlock("repeat.body")
	done := fn.NewBasicBlock("repeat.done") // target of 'break'

	fn.emitJump(body)
	fn.emitReturn(b.expr(fn, s.Condition), body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	fn.emitJump(loop)
	fn.currentBlock = done
}

func (b *builder) whileStmt(fn *Function, s *ast.WhileStmt) {
	loop := fn.NewBasicBlock("while.loop") // target of 'continue'
	body := fn.NewBasicBlock("while.body")
	done := fn.NewBasicBlock("while.done") // target of 'break'

	fn.emitJump(loop)
	fn.currentBlock = loop
	fn.emitIf(b.expr(fn, s.Condition), body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	fn.emitJump(loop)
	fn.currentBlock = done
}

func (b *builder) numberForStmt(fn *Function, s *ast.NumberForStmt) {
	loop := fn.NewBasicBlock("for.loop") // target of 'continue'
	body := fn.NewBasicBlock("for.body")
	done := fn.NewBasicBlock("for.done") // target of 'break'

	local := fn.addLocal(s.Name)

	limit := b.expr(fn, s.Limit)
	init := b.expr(fn, s.Init)
	step := b.expr(fn, s.Step)

	fn.emitJump(loop)
	fn.emitNumberFor(local, init, limit, step, body, done)

	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	fn.emitJump(loop)
	fn.currentBlock = done
}

func (b *builder) genericForStmt(fn *Function, s *ast.GenericForStmt) {
	loop := fn.NewBasicBlock("for.loop") // target of 'continue'
	body := fn.NewBasicBlock("for.body")
	done := fn.NewBasicBlock("for.done") // target of 'break'

	locals := make([]Value, len(s.Names))
	values := make([]Value, len(s.Exprs))

	for i, name := range s.Names {
		locals[i] = fn.lookup(name)
	}

	for i, expr := range s.Exprs {
		values[i] = b.expr(fn, expr)
	}

	fn.emitJump(loop)
	fn.emitGenericFor(locals, values, body, done)
	fn.currentBlock = body
	b.chunk(fn, s.Chunk)
	fn.emitJump(loop)
	fn.currentBlock = done
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
		if len(s.Lhs) <= len(s.Rhs) { // a, b = 1, 2 or a, b = 1, 2, 3
			for i, ex := range s.Lhs {
				fn.EmitAssign(b.expr(fn, ex), b.expr(fn, s.Rhs[i]))
			}
		} else { // a, b = 1
			i, l, r := 0, len(s.Lhs), len(s.Rhs)
			for ; i < l; i++ {
				fn.EmitAssign(b.expr(fn, s.Lhs[i]), b.expr(fn, s.Rhs[i]))
			}
			for ; i < r; i++ {
				fn.EmitAssign(b.expr(fn, s.Lhs[i]), b.expr(fn, &ast.NilExpr{}))
			}
		}
	case *ast.CompoundAssignStmt:
		if len(s.Lhs) <= len(s.Rhs) { // a, b = 1, 2 or a, b = 1, 2, 3
			for i, ex := range s.Lhs {
				fn.emitCompoundAssign(s.Operator, b.expr(fn, ex), b.expr(fn, s.Rhs[i]))
			}
		} else { // a, b = 1
			i, l, r := 0, len(s.Lhs), len(s.Rhs)
			for ; i < l; i++ {
				fn.emitCompoundAssign(s.Operator, b.expr(fn, s.Lhs[i]), b.expr(fn, s.Rhs[i]))
			}
			for ; i < r; i++ {
				fn.emitCompoundAssign(s.Operator, b.expr(fn, s.Lhs[i]), b.expr(fn, &ast.NilExpr{}))
			}
		}
	case *ast.LocalAssignStmt:
		switch {
		case len(s.Names) <= len(s.Exprs): // local a, b = 1, 2
			for i, name := range s.Names {
				fn.emitLocalAssign(name, b.expr(fn, s.Exprs[i]))
			}
		case len(s.Exprs) == 0: // local a, b
			for _, name := range s.Names {
				fn.emitLocalAssign(name, b.expr(fn, &ast.NilExpr{}))
			}
		default: // local a, b = 1
			i, e, n := 0, len(s.Exprs), len(s.Names)
			for ; i < e; i++ {
				fn.emitLocalAssign(s.Names[i], b.expr(fn, s.Exprs[i]))
			}
			for ; i < n; i++ {
				fn.emitLocalAssign(s.Names[i], b.expr(fn, &ast.NilExpr{}))
			}
		}
	case *ast.FuncCallStmt:
		call := b.funcCallExpr(fn, s.Expr.(*ast.FuncCallExpr))
		fn.emit(&call)
	case *ast.DoBlockStmt:
		b.chunk(fn, s.Chunk)
	case *ast.WhileStmt:
		b.whileStmt(fn, s)
	case *ast.RepeatStmt:
		b.repeatStmt(fn, s)
	case *ast.LocalFunctionStmt:
		f := fn.addFunction(s.Func)
		f.Name = s.Name
		fn.emitLocalAssign(s.Name, f)
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
			lhs = AttrGet{b.expr(fn, s.Name.Receiver), Const{s.Name.Method}}
			f.Name = s.Name.Method
			f.addParam("self")
		}
		b.buildFunction(f)
		fn.EmitAssign(lhs, f)
	case *ast.ReturnStmt:
		// some some trickery to convert the exprs into useable values
		// something like
		// exprs(s.Exprs)
		//fn.emit(&Return{Results: s.Exprs, pos: s.Return})
		fn.currentBlock = fn.NewBasicBlock("unreachable")
	case *ast.IfStmt:
		then := fn.NewBasicBlock("if.then")
		done := fn.NewBasicBlock("if.done")
		els := done
		if s.Else != nil {
			els = fn.NewBasicBlock("if.else")
		}

		fn.emitIf(b.expr(fn, s.Condition), then, els) //TODO: convert to jmp stuff like while is
		fn.currentBlock = then
		b.chunk(fn, s.Then)
		fn.emitJump(done)

		if s.Else != nil {
			fn.currentBlock = els
			b.chunk(fn, s.Else)
			fn.emitJump(done)
		}
		fn.currentBlock = done
	case *ast.BreakStmt:
		fn.emitJump(fn.breakBlock)
		fn.currentBlock = fn.NewBasicBlock("unreachable")
	case *ast.ContinueStmt:
		fn.emitJump(fn.continueBlock)
		fn.currentBlock = fn.NewBasicBlock("unreachable")
	case *ast.NumberForStmt:
		b.numberForStmt(fn, s)
	case *ast.GenericForStmt:
		b.genericForStmt(fn, s)
	default:
		panic(fmt.Sprintf("unexpected statement kind: %T", s))
	}
}
