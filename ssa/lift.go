package ssa

import "math/big"

// domFrontier maps each block to the set of blocks in its dominance
// frontier.  The outer slice is conceptually a map keyed by
// Block.Index.  The inner slice is conceptually a set, possibly
// containing duplicates.
//
// domFrontier's methods mutate the slice's elements but not its
// length, so their receivers needn't be pointers.
//

type DomFrontier [][]*BasicBlock

func (df DomFrontier) add(u, v *BasicBlock) {
	p := &df[u.Index]
	*p = append(*p, v)
}

func (df DomFrontier) build(u *BasicBlock) {
	for _, child := range u.dom.children {
		df.build(child)
	}
	for _, vb := range u.Succs {
		if v := vb.dom; v.idom != u {
			df.add(u, vb)
		}
	}
	for _, vb := range u.UnSuccs {
		if v := vb.dom; v.idom != u {
			df.add(u, vb)
		}
	}
	for _, w := range u.dom.children {
		for _, vb := range df[w.Index] {
			if v := vb.dom; v.idom != u {
				df.add(u, vb)
			}
		}
	}
}

func BuildDomFrontier(fn *Function) {
	fn.DomFrontier = make(DomFrontier, len(fn.Blocks))
	fn.DomFrontier.build(fn.Blocks[0])
}

func removeInstr(refs []Instruction, instr Instruction) []Instruction {
	i := 0
	for _, ref := range refs {
		if ref == instr {
			continue
		}
		refs[i] = ref
		i++
	}
	for j := i; j != len(refs); j++ {
		refs[j] = nil // aid GC
	}
	return refs[:i]
}

// lift replaces local and new Allocs accessed only with
// load/store by SSA registers, inserting φ-nodes where necessary.
// The result is a program in classical pruned SSA form.
//
// Preconditions:
// - fn has no dead blocks (blockopt has run).
// - Def/use info (Operands and Referrers) is up-to-date.
// - The dominator tree is up-to-date.
func lift(fn *Function) {
	BuildDomFrontier(fn)
	df := fn.DomFrontier
	newPhis := make(newPhiMap)

	// Determine which allocs we can lift and number them densely.
	// The renaming phase uses this numbering for compact maps.
	numAllocs := 0
	for _, b := range fn.Blocks {
		b.gaps = 0
		for _, instr := range b.Instrs {
			switch instr := instr.(type) {
			case *Define:
				liftAlloc(df, instr, newPhis)
				instr.index = numAllocs
				numAllocs++
			}
		}
	}

	// renaming maps an alloc (keyed by index) to its replacement
	// value.  Initially the renaming contains nil, signifying the
	// zero constant of the appropriate type; we construct the
	// Const lazily at most once on each path through the domtree.
	// TODO(adonovan): opt: cache per-function not per subtree.
	renaming := make([]Value, numAllocs)

	// Renaming.
	rename(fn.Blocks[0], renaming, newPhis)

	// Eliminate dead φ-nodes.
	removeDeadPhis(fn.Blocks, newPhis)

	// Prepend remaining live φ-nodes to each block.
	for _, b := range fn.Blocks {
		nps := newPhis[b]
		j := len(nps)

		if j+b.gaps == 0 {
			continue // fast path: no new phis or gaps
		}

		// Compact nps + non-nil Instrs into a new slice.
		// TODO(adonovan): opt: compact in situ (rightwards)
		// if Instrs has sufficient space or slack.
		dst := make([]Instruction, len(b.Instrs)+j-b.gaps)
		for i, np := range nps {
			dst[i] = np.phi
		}
		for _, instr := range b.Instrs {
			if instr == nil {
				continue
			}
			dst[j] = instr
			j++
		}
		b.Instrs = dst
	}

	// Remove any fn.Locals that were lifted.
	j := 0
	for _, l := range fn.Locals {
		if l.Num < 0 {
			fn.Locals[j] = l
			j++
		}
	}
	// Nil out fn.Locals[j:] to aid GC.
	for i := j; i < len(fn.Locals); i++ {
		fn.Locals[i] = nil
	}
	fn.Locals = fn.Locals[:j]
}

// removeDeadPhis removes φ-nodes not transitively needed by a
// non-Phi, non-DebugRef instruction.
func removeDeadPhis(blocks []*BasicBlock, newPhis newPhiMap) {
	// First pass: find the set of "live" φ-nodes: those reachable
	// from some non-Phi instruction.
	//
	// We compute reachability in reverse, starting from each φ,
	// rather than forwards, starting from each live non-Phi
	// instruction, because this way visits much less of the
	// Value graph.
	livePhis := make(map[*Phi]bool)
	for _, npList := range newPhis {
		for _, np := range npList {
			phi := np.phi
			if !livePhis[phi] && phiHasDirectReferrer(phi) {
				markLivePhi(livePhis, phi)
			}
		}
	}

	// Existing φ-nodes due to && and || operators
	// are all considered live (see Go issue 19622).
	for _, b := range blocks {
		for _, phi := range b.phis() {
			markLivePhi(livePhis, phi.(*Phi))
		}
	}

	// Second pass: eliminate unused phis from newPhis.
	for block, npList := range newPhis {
		j := 0
		for _, np := range npList {
			if livePhis[np.phi] {
				npList[j] = np
				j++
			} else {
				// discard it, first removing it from referrers
				for _, val := range np.phi.Edges {
					if refs := val.Referrers(); refs != nil {
						*refs = removeInstr(*refs, np.phi)
					}
				}
				np.phi.block = nil
			}
		}
		newPhis[block] = npList[:j]
	}
}

// markLivePhi marks phi, and all φ-nodes transitively reachable via
// its Operands, live.
func markLivePhi(livePhis map[*Phi]bool, phi *Phi) {
	livePhis[phi] = true
	for _, rand := range phi.Operands(nil) {
		if q, ok := (*rand).(*Phi); ok {
			if !livePhis[q] {
				markLivePhi(livePhis, q)
			}
		}
	}
}

// phiHasDirectReferrer reports whether phi is directly referred to by
// a non-Phi instruction.  Such instructions are the
// roots of the liveness traversal.
func phiHasDirectReferrer(phi *Phi) bool {
	for _, instr := range *phi.Referrers() {
		if _, ok := instr.(*Phi); !ok {
			return true
		}
	}
	return false
}

type blockSet struct{ big.Int } // (inherit methods from Int)

// add adds b to the set and returns true if the set changed.
func (s *blockSet) add(b *BasicBlock) bool {
	i := b.Index
	if s.Bit(i) != 0 {
		return false
	}
	s.SetBit(&s.Int, i, 1)
	return true
}

// take removes an arbitrary element from a set s and
// returns its index, or returns -1 if empty.
func (s *blockSet) take() int {
	l := s.BitLen()
	for i := 0; i < l; i++ {
		if s.Bit(i) == 1 {
			s.SetBit(&s.Int, i, 0)
			return i
		}
	}
	return -1
}

// newPhi is a pair of a newly introduced φ-node and the lifted Alloc
// it replaces.
type newPhi struct {
	phi   *Phi
	alloc *Define
}

// newPhiMap records for each basic block, the set of newPhis that
// must be prepended to the block.
type newPhiMap map[*BasicBlock][]newPhi

// liftAlloc determines whether alloc can be lifted into registers,
// and if so, it populates newPhis with all the φ-nodes it may require
// and returns true.
//
// fresh is a source of fresh ids for phi nodes.
func liftAlloc(df DomFrontier, define *Define, newPhis newPhiMap) {
	// Compute defblocks, the set of blocks containing a
	// definition of the alloc cell.
	var defblocks blockSet
	for _, instr := range *define.Referrers() {
		if instr, ok := instr.(*Assign); ok {
			defblocks.add(instr.Block())
		}
	}
	// The define itself counts as a (zero) definition of the cell.
	defblocks.add(define.Block())

	fn := define.Parent()

	var hasAlready blockSet

	// Initialize W and work to defblocks.
	var work blockSet = defblocks // blocks seen
	var W blockSet                // blocks to do
	W.Set(&defblocks.Int)

	// Traverse iterated dominance frontier, inserting φ-nodes.
	for i := W.take(); i != -1; i = W.take() {
		u := fn.Blocks[i]
		for _, v := range df[u.Index] {
			if hasAlready.add(v) {
				// Create φ-node.
				// It will be prepended to v.Instrs later, if needed.
				phi := &Phi{
					Edges:   make([]Value, len(v.Preds)),
					Comment: define.Comment,
				}

				phi.block = v
				newPhis[v] = append(newPhis[v], newPhi{phi, define})

				if work.add(v) {
					W.add(v)
				}
			}
		}
	}
}

// replaceAll replaces all intraprocedural uses of x with y,
// updating x.Referrers and y.Referrers.
// Precondition: x.Referrers() != nil, i.e. x must be local to some function.
func replaceAll(x, y Value) {
	var rands []*Value
	pxrefs := x.Referrers()
	pyrefs := y.Referrers()
	for _, instr := range *pxrefs {
		rands = instr.Operands(rands[:0]) // recycle storage
		for _, rand := range rands {
			if *rand != nil {
				if *rand == x {
					*rand = y
				}
			}
		}
		if pyrefs != nil {
			*pyrefs = append(*pyrefs, instr) // dups ok
		}
	}
	*pxrefs = nil // x is now unreferenced
}

// renamed returns the value to which alloc is being renamed,
// constructing it lazily if it's the implicit zero initialization.
func renamed(renaming []Value, alloc *Alloc) Value {
	v := renaming[alloc.index]
	if v == nil {
		v = zeroConst(mustDeref(alloc.Type()))
		renaming[alloc.index] = v
	}
	return v
}

// rename implements the (Cytron et al) SSA renaming algorithm, a
// preorder traversal of the dominator tree replacing all loads of
// Alloc cells with the value stored to that cell by the dominating
// store instruction.  For lifting, we need only consider loads,
// stores and φ-nodes.
//
// renaming is a map from *Alloc (keyed by index number) to its
// dominating stored value; newPhis[x] is the set of new φ-nodes to be
// prepended to block x.
func rename(u *BasicBlock, renaming []Value, newPhis newPhiMap) {
	// Each φ-node becomes the new name for its associated Alloc.
	for _, np := range newPhis[u] {
		phi := np.phi
		alloc := np.alloc
		renaming[alloc.index] = phi
	}

	// Rename loads and stores of allocs.
	for i, instr := range u.Instrs {
		switch instr := instr.(type) {
		case *Define:
			if instr.index >= 0 { // store of zero to Alloc cell
				// Replace dominated loads by the zero value.
				renaming[instr.index] = nil
				// Delete the Alloc.
				u.Instrs[i] = nil
				u.gaps++
			}

		case *Assign:
			if alloc, ok := instr.Addr.(*Define); ok && alloc.index >= 0 { // store to Alloc cell
				// Replace dominated loads by the stored value.
				renaming[alloc.index] = instr.Val
				// Remove the store from the referrer list of the stored value.
				if refs := instr.Val.Referrers(); refs != nil {
					*refs = removeInstr(*refs, instr)
				}
				// Delete the Store.
				u.Instrs[i] = nil
				u.gaps++
			}
		}
	}

	// For each φ-node in a CFG successor, rename the edge.
	for _, v := range u.Succs {
		phis := newPhis[v]
		if len(phis) == 0 {
			continue
		}
		i := v.predIndex(u)
		for _, np := range phis {
			phi := np.phi
			alloc := np.alloc
			newval := renamed(renaming, alloc)
			phi.Edges[i] = newval
			if prefs := newval.Referrers(); prefs != nil {
				*prefs = append(*prefs, phi)
			}
		}
	}

	// Continue depth-first recursion over domtree, pushing a
	// fresh copy of the renaming map for each subtree.
	for i, v := range u.dom.children {
		r := renaming
		if i < len(u.dom.children)-1 {
			// On all but the final iteration, we must make
			// a copy to avoid destructive update.
			r = make([]Value, len(renaming))
			copy(r, renaming)
		}
		rename(v, r, newPhis)
	}

}
