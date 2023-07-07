package ssa

func (f *Function) emitIf(cond Value, tblock, fblock *BasicBlock) {
	b := f.currentBlock
	b.emit(&If{Cond: cond})
	AddEdge(b, tblock)
	AddEdge(b, fblock)
	f.currentBlock = nil
}

func (f *Function) emitGenericFor(locals, values []Value, body, done *BasicBlock) {
	b := f.currentBlock
	b.emit(&GenericFor{
		Locals: locals,
		Values: values,
	})
	AddEdge(b, body)
	AddEdge(b, done)
	f.currentBlock = nil
}

func (f *Function) emitNumberFor(local, init, limit, step Value, body, done *BasicBlock) {
	b := f.currentBlock
	b.emit(&NumberFor{
		Local: local,
		Init:  init,
		Limit: limit,
		Step:  step,
	})
	AddEdge(b, body)
	AddEdge(b, done)
	f.currentBlock = nil
}

func (f *Function) emitCompoundAssign(op string, lhs []Value, rhs []Value) {
	f.Emit(&CompoundAssign{
		Op:  op,
		Lhs: lhs,
		Rhs: rhs,
	})
}

func (f *Function) EmitAssign(locals []*Local) {
	f.Emit(&Define{
		Locals: locals,
	})
}

func (f *Function) emitReturn(cond Value, body *BasicBlock, done *BasicBlock) {
	f.Emit(&Return{})
}

func (f *Function) emitLocalAssign(names []string, values []Value) {
	assign := &Define{
		Locals: make([]*Local, len(names)),
	}

	for i, name := range names {
		assign.Locals[i] = f.addLocal(name)
		assign.Locals[i].Value = values[i]
	}
	f.Emit(assign)
}

func (f *Function) emitJump(target *BasicBlock) {
	b := f.currentBlock
	// from to
	b.emit(new(Jump))
	AddEdge(b, target)
	f.currentBlock = nil
}
