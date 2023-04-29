// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// Simple block optimizations to simplify the control flow graph.

func markReachable(b *BasicBlock) {
	b.reachable = true
	for _, succ := range b.Succs {
		if !succ.reachable {
			markReachable(succ)
		}
	}
}

/*
	if len(b.Succs) > 0 {
		for _, succ := range b.UnSuccs {
			newUnPreds := make([]*BasicBlock, 0, len(b.UnPreds))
			for _, pred := range succ.UnPreds {
				if pred != b {
					newUnPreds = append(newUnPreds, pred)
				}
			}
			succ.UnPreds = newUnPreds
		}
	}
*/

func MarkUnreachableBlocks(f *Function) {
	markReachable(f.Blocks[0])
	for i, b := range f.Blocks {
		if !b.reachable {
			for _, c := range b.Succs {
				c.UnPreds = append(c.UnPreds, b)
				if c.reachable {
					c.removePred(b) // delete reachable->unreachable edge
				}
			}
			// fallthrough edge to unreachable block
			b.UnPreds = append(b.UnPreds, f.Blocks[i-1])
			//b.UnPreds = append(b.UnPreds, f.Blocks[i-1])
			b.UnSuccs = append(b.UnSuccs, b.Succs...)
			b.succs2 = [2]*BasicBlock{}
			b.Succs = b.succs2[:0]
		}
	}
}

// jumpThreading attempts to apply simple jump-threading to block b,
// in which a->b->c become a->c if b is just a Jump.
// The result is true if the optimization was applied.
func jumpThreading(f *Function, b *BasicBlock) bool {
	if b.Index == 0 {
		return false // don't apply to entry block
	}
	if b.Instrs == nil {
		return false
	}
	if _, ok := b.Instrs[0].(*Jump); !ok {
		return false // not just a jump
	}
	c := b.Succs[0]
	if c == b {
		return false // don't apply to degenerate jump-to-self.
	}
	if c.hasPhi() {
		return false // not sound without more effort
	}
	for j, a := range b.Preds {
		a.replaceSucc(b, c)

		// If a now has two edges to c, replace its degenerate If by Jump.
		if len(a.Succs) == 2 && a.Succs[0] == c && a.Succs[1] == c {
			a.Succs = a.Succs[:1]
			c.removePred(b)
		} else {
			if j == 0 {
				c.replacePred(b, a)
			} else {
				c.Preds = append(c.Preds, a)
			}
		}
	}
	f.Blocks[b.Index] = nil // delete b
	return true
}

// fuseBlocks attempts to apply the block fusion optimization to block
// a, in which a->b becomes ab if len(a.Succs)==len(b.Preds)==1.
// The result is true if the optimization was applied.
func fuseBlocks(f *Function, a *BasicBlock) bool {
	if len(a.Succs) != 1 {
		return false
	}
	b := a.Succs[0]
	if len(b.Preds) != 1 {
		return false
	}

	// Degenerate &&/|| ops may result in a straight-line CFG
	// containing φ-nodes. (Ideally we'd replace such them with
	// their sole operand but that requires Referrers, built later.)
	if b.hasPhi() {
		return false // not sound without further effort
	}

	// Eliminate jump at end of A, then copy all of B across.
	a.Instrs = append(a.Instrs[:len(a.Instrs)-1], b.Instrs...)
	for _, instr := range b.Instrs {
		instr.SetBlock(a)
	}

	// A inherits B's successors
	a.Succs = append(a.succs2[:0], b.Succs...)

	// Fix up Preds links of all successors of B.
	for _, c := range b.Succs {
		c.replacePred(b, a)
	}

	f.Blocks[b.Index] = nil // delete b
	return true
}

// optimizeBlocks() performs some simple block optimizations on a
// completed function: dead block elimination, block fusion, jump
// threading.
func optimizeBlocks(f *Function) {

	// Loop until no further progress.
	changed := true
	for changed {
		changed = false

		for _, b := range f.Blocks {
			// f.Blocks will temporarily contain nils to indicate
			// deleted blocks; we remove them at the end.
			if b == nil {
				continue
			}

			// Fuse blocks.  b->c becomes bc.
			if fuseBlocks(f, b) {
				changed = true
			}

			// a->b->c becomes a->c if b contains only a Jump.
			if jumpThreading(f, b) {
				changed = true
				continue // (b was disconnected)
			}
		}
	}
	f.removeNilBlocks()
}
