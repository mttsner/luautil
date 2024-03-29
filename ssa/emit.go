package ssa

func (f *Function) emitIf(cond Value, tblock, fblock *BasicBlock) {
	b := f.currentBlock
	b.emit(&If{Cond: cond})
	addEdge(b, tblock)
	addEdge(b, fblock)
	f.currentBlock = nil
}

func (f *Function) emitGenericFor(locals, values []Value, body, done *BasicBlock) {
	b := f.currentBlock
	b.emit(&GenericFor{
		Locals: locals,
		Values: values,
	})
	addEdge(b, body)
	addEdge(b, done)
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
	addEdge(b, body)
	addEdge(b, done)
	f.currentBlock = nil
}

func (f *Function) emitCompoundAssign(op string, lhs Value, rhs Value) {
	f.emit(&CompoundAssign{
		Op:  op,
		Lhs: lhs,
		Rhs: rhs,
	})
}

func (f *Function) EmitAssign(lhs Value, rhs Value) {
	f.emit(&Assign{
		Lhs: lhs,
		Rhs: rhs,
	})
}

func (f *Function) emitReturn(cond Value, body *BasicBlock, done *BasicBlock) {
	f.emit(&Return{})
}

func (f *Function) emitLocalAssign(name string, value Value) {
	local := f.addLocal(name)
	f.EmitAssign(local, value)
}

func (f *Function) emitJump(target *BasicBlock) {
	b := f.currentBlock
	b.emit(new(Jump))
	addEdge(b, target)
	f.currentBlock = nil
}
