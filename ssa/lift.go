package ssa

import "math/big"

// DomFrontier maps each block to the set of blocks in its dominance
// frontier.  The outer slice is conceptually a map keyed by
// Block.Index.  The inner slice is conceptually a set, possibly
// containing duplicates.
//
// domFrontier's methods mutate the slice's elements but not its
// length, so their receivers needn't be pointers.
type DomFrontier [][]*BasicBlock

// newPhiMap records for each basic block, the set of newPhis that
// must be prepended to the block.
type newPhiMap map[*BasicBlock][]newPhi

// newPhi is a pair of a newly introduced φ-node and the local it replaces.
type newPhi struct {
	phi   *Phi
	local *Local
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

// lift generates pruned SSA form
// Preconditions:
// - fn has no dead blocks (blockopt has run).
// - Def/use info (Operands and Referrers) is up-to-date.
// - The dominator tree is up-to-date.
// - The dominancce frontier is up-to-date,
func lift(fn *Function) {
	// Creation of φ-nodes.
	newPhis := make(newPhiMap)
	for _, l := range fn.Locals {
		placePhis(fn.DomFrontier, l, newPhis)
	}
	// Renaming.
	entry := fn.Blocks[0]
	locals := make([]*Local, 0, len(fn.Locals))
	renaming := make([]Value, len(fn.Locals))
	fn.Locals = rename(entry, renaming, locals, newPhis)
}

// placePhis uses the (Cytron et al) SSA construction algorithm to place phi nodes
func placePhis(df DomFrontier, local *Local, newPhis newPhiMap) {
	// Compute defblocks, the set of blocks containing a
	// refrence to the local
	var defblocks blockSet
	for _, instr := range *local.Referrers() {
		if instr, ok := instr.(*Assign); ok {
			defblocks.add(instr.Block())
		}
	}
	// The block the local was defined in itself counts 
	// as a (zero) reference of the local.
	defblocks.add(local.DefBlock())

	fn := local.Parent()

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
				phi := &Phi{
					Edges:   make([]Value, len(v.Preds)),
					Comment: local.Comment,
				}

				phi.block = v
				newPhis[v] = append(newPhis[v], newPhi{phi, local})

				if work.add(v) {
					W.add(v)
				}
			}
		}
	}
}

func renamed(renaming []Value, v Value) Value {
	switch v := v.(type) {
	case *Local:
		return renaming[v.Index]
	case *Global, Nil, True, False, Number, String, VarArg:
		return v
	case Table:
		for _, field := range v.Fields {
			field.Key = renamed(renaming, field.Key)
			field.Value = renamed(renaming, field.Value)
		}
		return v
	case AttrGet:
		v.Object = renamed(renaming, v.Object)
		v.Key = renamed(renaming, v.Key)
		return v
	case Unary:
		v.Value = renamed(renaming, v.Value)
		return v
	case Arithmetic: // All of these could definitely be turned into one
		v.Lhs = renamed(renaming, v.Lhs)
		v.Rhs = renamed(renaming, v.Rhs)
		return v
	case Concat:
		v.Lhs = renamed(renaming, v.Lhs)
		v.Rhs = renamed(renaming, v.Rhs)
		return v
	case Relation:
		v.Lhs = renamed(renaming, v.Lhs)
		v.Rhs = renamed(renaming, v.Rhs)
		return v
	case Logic:
		v.Lhs = renamed(renaming, v.Lhs)
		v.Rhs = renamed(renaming, v.Rhs)
		return v
	default:
		panic("unimplemented")
	}
}

func replaced(renaming []Value, new []*Local, v Value) (Value, []*Local) {
	if l, ok := v.(*Local); ok {
		newLocal := &Local{
			Comment: l.Comment,
			Index:   len(new),
		}
		renaming[l.Index] = newLocal
		return newLocal, append(new, newLocal)
	}
	return v, new
}

// rename implements the (Cytron et al) SSA renaming algorithm.
func rename(b *BasicBlock, renaming []Value, newLocals []*Local, newPhis newPhiMap) []*Local {
	// Each φ-node becomes the new name for its associated local.
	for _, np := range newPhis[b] {
		phi := np.phi
		local := np.local
		renaming[local.Index] = phi
	}
	// For each instruction in the basicblock, rename and replace all usages of locals.
	for _, instr := range b.Instrs {
		switch instr := instr.(type) {
		case *Assign:
			for i, v := range instr.Rhs {
				instr.Rhs[i] = renamed(renaming, v)
			}
			for i, v := range instr.Lhs {
				instr.Lhs[i], newLocals = replaced(renaming, newLocals, v)
			}
		case *CompoundAssign:
			panic("todo")
		case *NumberFor:
			instr.Init = renamed(renaming, instr.Init)
			instr.Step = renamed(renaming, instr.Step)
			instr.Limit = renamed(renaming, instr.Limit)
			instr.Local, newLocals = replaced(renaming, newLocals, instr.Local)
		case *GenericFor:
			for i, v := range instr.Values {
				instr.Values[i] = renamed(renaming, v)
			}
			for i, v := range instr.Locals {
				instr.Locals[i], newLocals = replaced(renaming, newLocals, v)
			}
		case *If:
			instr.Cond = renamed(renaming, instr.Cond)
		case *Call:
			if instr.Func != nil {
				instr.Func = renamed(renaming, instr.Func)
			} else {
				instr.Recv = renamed(renaming, instr.Recv)
			}
			for i,v := range instr.Args {
				instr.Args[i] = renamed(renaming, v)
			}
		case *Return:
			for i, v := range instr.Values {
				instr.Values[i] = renamed(renaming, v)
			}
		default:
			panic("unimplemented")
		}
	}
	// For each φ-node in a CFG successor, rename the edge.
	for _, v := range b.Succs {
		phis := newPhis[v]
		if len(phis) == 0 {
			continue
		}
		i := v.predIndex(b)
		for _, np := range phis {
			phi := np.phi
			alloc := np.local
			newval := renamed(renaming, alloc)
			phi.Edges[i] = newval
		}
	}
	// Continue depth-first recursion over domtree, pushing a
	// fresh copy of the renaming map for each subtree.
	for i, v := range b.dom.children {
		r := renaming
		if i < len(b.dom.children)-1 {
			// On all but the final iteration, we must make
			// a copy to avoid destructive update.
			r = make([]Value, len(renaming))
			copy(r, renaming)
		}
		newLocals = rename(v, r, newLocals, newPhis)
	}
	return newLocals
}
