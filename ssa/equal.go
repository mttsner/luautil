package ssa

func (f *Function) Equal(fn *Function) bool {
	if f.VarArg != fn.VarArg {
		return false
	}
	if len(f.Params) != len(fn.Params) {
		return false
	}
	if len(f.UpValues) != len(fn.UpValues) {
		return false
	}
	//if len(f.Locals) != len(fn.Locals) {
	//	return false
	//}

	if len(f.Blocks) != len(fn.Blocks) {
		return false
	}

	for i, b := range f.Blocks {
		if !b.Equal(fn.Blocks[i]) {
			return false
		}
	}

	return true
}

func (b *BasicBlock) Equal(bb *BasicBlock) bool {
	if len(b.Preds) != len(bb.Preds) {
		return false
	}

	if len(b.Succs) != len(bb.Succs) {
		return false
	}

	if len(b.Instrs) != len(bb.Instrs) {
		return false
	}

	for i, instr := range b.Instrs {
		if instr != b.Instrs[i] {
			return false
		}
		//if !instr.Equal(b.Instrs[i]) {
		//	return false
		//}
	}
	return true
}

func (i Call) Equal(instr Instruction) bool {
	panic("Call: Not Implemented")
}

func (i Return) Equal(instr Instruction) bool {
	_, success := instr.(*Return)
	if !success {
		return false
	}
	return true
}

func (i If) Equal(instr Instruction) bool {
	panic("If: Not Implemented")
}

func (i GenericFor) Equal(instr Instruction) bool {
	panic("GenericFor: Not Implemented")
}

func (i NumberFor) Equal(instr Instruction) bool {
	panic("NumberFor: Not Implemented")
}

func (i CompoundAssign) Equal(instr Instruction) bool {
	panic("CompoundAssign: Not Implemented")
}

func (s Concat) Equal(instr Instruction) bool {
	_, success := instr.(*Assign)
	if !success {
		return false
	}
	return true
}

func (s Assign) Equal(instr Instruction) bool {
	v, success := instr.(*Assign)
	if !success {
		return false
	}

	if len(s.Lhs) != len(v.Lhs) {
		return false
	}

	if len(s.Rhs) != len(v.Rhs) {
		return false
	}

	for i, val := range s.Rhs {
		if val != v.Rhs[i] {
			return false
		}
	}
	return true
}


func (i Jump) Equal(instr Instruction) bool {
	panic("Jump: Not Implemented")
}

func (i Phi) Equal(instr Instruction) bool {
	panic("Phi: Not Implemented")
}
