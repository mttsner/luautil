package ssa

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