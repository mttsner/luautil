package ssa

// This file implements the Function and BasicBlock types.

import (
	"fmt"
	"strings"

	"github.com/notnoobmaster/luautil/ast"
)

// addEdge adds a control-flow graph edge from from to to.
func addEdge(from, to *BasicBlock) {
	from.Succs = append(from.Succs, to)
	to.Preds = append(to.Preds, from)
}

// Parent returns the function that contains block b.
func (b *BasicBlock) Parent() *Function { return b.parent }

// emit appends an instruction to the current basic block.
// If the instruction defines a Value, it is returned.
//
func (b *BasicBlock) emit(i Instruction) Value {
	i.SetBlock(b)
	b.Instrs = append(b.Instrs, i)
	v, _ := i.(Value)
	return v
}

// predIndex returns the i such that b.Preds[i] == c or panics if
// there is none.
func (b *BasicBlock) predIndex(c *BasicBlock) int {
	for i, pred := range b.Preds {
		if pred == c {
			return i
		}
	}
	panic(fmt.Sprintf("no edge %s -> %s", c, b))
}

// hasPhi returns true if b.Instrs contains φ-nodes.
func (b *BasicBlock) hasPhi() bool {
	_, ok := b.Instrs[0].(*Phi)
	return ok
}

// phis returns the prefix of b.Instrs containing all the block's φ-nodes.
func (b *BasicBlock) phis() []Instruction {
	for i, instr := range b.Instrs {
		if _, ok := instr.(*Phi); !ok {
			return b.Instrs[:i]
		}
	}
	return nil // unreachable in well-formed blocks
}

// replacePred replaces all occurrences of p in b's predecessor list with q.
// Ordinarily there should be at most one.
//
func (b *BasicBlock) replacePred(p, q *BasicBlock) {
	for i, pred := range b.Preds {
		if pred == p {
			b.Preds[i] = q
		}
	}
}

// replaceSucc replaces all occurrences of p in b's successor list with q.
// Ordinarily there should be at most one.
//
func (b *BasicBlock) replaceSucc(p, q *BasicBlock) {
	for i, succ := range b.Succs {
		if succ == p {
			b.Succs[i] = q
		}
	}
}

// removePred removes all occurrences of p in b's
// predecessor list and φ-nodes.
// Ordinarily there should be at most one.
//
func (b *BasicBlock) removePred(p *BasicBlock) {
	phis := b.phis()

	// We must preserve edge order for φ-nodes.
	j := 0
	for i, pred := range b.Preds {
		if pred != p {
			b.Preds[j] = b.Preds[i]
			// Strike out φ-edge too.
			for _, instr := range phis {
				phi := instr.(*Phi)
				phi.Edges[j] = phi.Edges[i]
			}
			j++
		}
	}
	// Nil out b.Preds[j:] and φ-edges[j:] to aid GC.
	for i := j; i < len(b.Preds); i++ {
		b.Preds[i] = nil
		for _, instr := range phis {
			instr.(*Phi).Edges[i] = nil
		}
	}
	b.Preds = b.Preds[:j]
	for _, instr := range phis {
		phi := instr.(*Phi)
		phi.Edges = phi.Edges[:j]
	}
}

// Destinations associated with a labelled block.
// We populate these as labels are encountered in forward gotos or
// labelled statements.
//
type lblock struct {
	_goto     *BasicBlock
	_break    *BasicBlock
	_continue *BasicBlock
}

// addParam adds a (non-escaping) parameter to f.Params of the
// specified name, type and source position.
//
func (f *Function) addParam(name string) {
	f.Params = append(f.Params, f.addLocal(name))
}

func (f *Function) addLocal(name string) *Local {
	local := &Local{
		Comment: name,
		Value:   Nil{},
		Num:     len(f.Locals),
	}
	f.Locals = append(f.Locals, local)
	f.currentScope.names[name] = local
	return local
}

func (f *Function) addGlobal(name string) *Global {
	global := &Global{
		Comment: name,
		Value:   String{Value: name},
	}
	//f.Globals = append(f.Globals, global)
	return global
}

func (f *Function) addFunction(syntax *ast.FunctionExpr) *Function {
	fn := &Function{
		parent: f,
		syntax: syntax,
		num:    len(f.Functions) + 1,
	}
	f.Functions = append(f.Functions, fn)
	return fn
}

// StartBody initializes the function prior to generating SSA code for its body.
// Precondition: f.Type() already set.
//
func (f *Function) StartBody() {
	f.currentBlock = f.NewBasicBlock("entry")
}

func (f *Function) newScope() *Scope {
	old := f.currentScope
	f.currentScope = &Scope{f, f.currentScope, make(map[string]Variable)}
	return old
}

// buildReferrers populates the def/use information in all non-nil
// Value.Referrers slice.
// Precondition: all such slices are initially empty.
/*
func buildReferrers(f *Function) {
	var rands []*Value
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			rands = instr.Operands(rands[:0]) // recycle storage
			for _, rand := range rands {
				if r := *rand; r != nil {
					if ref := r.Referrers(); ref != nil {
						*ref = append(*ref, instr)
					}
				}
			}
		}
	}
}
*/
// finishBody() finalizes the function after SSA code generation of its body.
func (f *Function) finishBody() {
	f.currentBlock = nil

	//buildReferrers(f)

	//lift(f)
}

// removeNilBlocks eliminates nils from f.Blocks and updates each
// BasicBlock.Index.  Use this after any pass that may delete blocks.
//
func (f *Function) removeNilBlocks() {
	j := 0
	for _, b := range f.Blocks {
		if b != nil {
			b.Index = j
			f.Blocks[j] = b
			j++
		}
	}
	// Nil out f.Blocks[j:] to aid GC.
	for i := j; i < len(f.Blocks); i++ {
		f.Blocks[i] = nil
	}
	f.Blocks = f.Blocks[:j]
}

func (s *Scope) lookup(name string) Value {
	if v, ok := s.names[name]; ok {
		return v
	}
	if s.parent == nil {
		return s.function.addGlobal(name)
	}
	return s.parent.lookup(name)
}

func (f *Function) lookup(name string) Value {
	return f.currentScope.lookup(name)
}

// Emit emits the specified instruction to function f.
func (f *Function) Emit(instr Instruction) Value {
	return f.currentBlock.emit(instr)
}

// NewBasicBlock adds to f a new basic block and returns it.  It does
// not automatically become the current block for subsequent calls to emit.
// comment is an optional string for more readable debugging output.
//
func (f *Function) NewBasicBlock(comment string) *BasicBlock {
	b := &BasicBlock{
		Index:   len(f.Blocks),
		Comment: comment,
		parent:  f,
	}
	b.Succs = b.succs2[:0]
	f.Blocks = append(f.Blocks, b)
	return b
}

func (f *Function) Syntax() *ast.FunctionExpr { return f.syntax }

func WriteFunction(b *strings.Builder, f *Function) {
	for _, fn := range f.Functions {
		WriteFunction(b, fn)
	}

	const punchcard = 80

	b.WriteString("\nfunction ")
	b.WriteString(f.Name)
	b.WriteString("(")
	for i, arg := range f.Params {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.String())
	}
	if f.VarArg {
		if len(f.Params) > 0 {
			b.WriteString(", ")
		}
		b.WriteString("...")
	}
	bmsg := fmt.Sprintf("locals:%d upvalues:%d", len(f.Locals), len(f.UpValues))
	fmt.Fprintf(b, ")%*s%s\n", punchcard-1-len(bmsg)-len(b.String()), "", bmsg)

	for _, block := range f.Blocks {
		if block == nil {
			// Corrupt CFG.
			b.WriteString(".nil:\n")
			continue
		}

		n, _ := fmt.Fprintf(b, "%d:", block.Index)
		bmsg := fmt.Sprintf("%s P:%d S:%d", block.Comment, len(block.Preds), len(block.Succs))
		fmt.Fprintf(b, "%*s%s\n", punchcard-1-n-len(bmsg), "", bmsg)

		if false { // CFG debugging
			fmt.Fprintf(b, "\t# CFG: %s --> %s --> %s\n", block.Preds, block, block.Succs)
		}

		for _, instr := range block.Instrs {
			b.WriteString("\t")
			if instr == nil {
				b.WriteString("<deleted>\n")
				continue
			}
			b.WriteString(instr.String())
			b.WriteString("\n")
		}
	}
	fmt.Fprintf(b, "end\n")
}

func WriteCfgDot(b *strings.Builder, f *Function) {
	//fmt.Fprintln(buf, "//", f)
	fmt.Fprintln(b, "digraph cfg {")
	for _, block := range f.Blocks {
		fmt.Fprintf(b, "\tn%d [label=\"", block.Index)
		for _, instr := range block.Instrs {
			if instr == nil {
				b.WriteString("<deleted>\n")
				continue
			}
			b.WriteString(instr.String())
			b.WriteString("\\n")
		}
		b.WriteString("\",shape=\"rectangle\"];\n")
		// CFG edges.
		for _, pred := range block.Preds {
			fmt.Fprintf(b, "\tn%d -> n%d [style=\"solid\",weight=100];\n", pred.Index, block.Index)
		}
	}
	fmt.Fprintln(b, "}")
}