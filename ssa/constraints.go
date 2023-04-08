package ssa

// isIf checks if the basic block has 2 successors && the block where the dominance of our block
// ends is equal to the false path block then our block must be the entry to an if statement
func (b *BasicBlock) isIf(dom DomFrontier) bool {
	// TODO: check if last element is a if instruction
	return len(b.Succs) == 2
}

// isIfElse checks if the basic block has 2 successors &&
// the dominance over both paths ends at the same node
func (b *BasicBlock) isIfElse(dom DomFrontier) bool {
	if len(b.Succs) != 2 {
		return false
	}
	tFront := dom[b.Succs[0].Index]
	fFront := dom[b.Succs[1].Index]
	return len(tFront) == len(fFront) && tFront[0].Index == fFront[0].Index
}

func (b *BasicBlock) isRepeat() bool {
	return len(b.Preds) == 2 && b.Dominates(b.Preds[1])
}

func (b *BasicBlock) isWhileLoop() bool {
	return len(b.Instrs) == 1 &&
		len(b.Succs) == 2 &&
		((len(b.Preds) > 1 && b.Dominates(b.Preds[1])) ||
			len(b.Succs[1].Preds) >= 2)
}

func (b *BasicBlock) isGoto() bool {
	lastI := len(b.Instrs) - 1
	_, ok := b.Instrs[lastI].(*Jump)
	return ok && len(b.Succs) == 1
}
