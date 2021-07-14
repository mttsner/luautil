package ssa

import "github.com/notnoobmaster/luautil/ast"

type Value interface {
}

type Instruction interface {
	String() string
	Parent() *Function
	Block() *BasicBlock
	setBlock(*BasicBlock)
}

type Node interface {
	// Common methods:
	String() string
	Parent() *Function

	// Partial methods:
	Operands(rands []*Value) []*Value // nil for non-Instructions
	Referrers() *[]Instruction        // nil for non-Values
}

type Function struct {
	name string

	syntax        *ast.FunctionExpr
	parent        *Function          // enclosing function if anon; nil if global
	Params        []*Parameter       // function parameters; for methods, includes receiver
	Locals        []*Local           // local variables of this function
	Blocks        []*BasicBlock      // basic blocks of the function; nil => external
	Globals       map[string]*Global // global variables of this function
	Names         map[string]*Local
	referrers     []Instruction // referring instructions (iff Parent() != nil)
	continueBlock *BasicBlock
	breakBlock    *BasicBlock

	// The following fields are set transiently during building,
	// then cleared.

	currentBlock *BasicBlock // where to emit code
}

type BasicBlock struct {
	Index        int            // index of this block within Parent().Blocks
	Comment      string         // optional label; no semantic significance
	parent       *Function      // parent function
	Instrs       []Instruction  // instructions in order
	Preds, Succs []*BasicBlock  // predecessors and successors
	succs2       [2]*BasicBlock // initial space for Succs
	gaps         int            // number of nil Instrs (transient)
	rundefers    int            // number of rundefers (transient)
}

type Parameter struct {
	name      string
	parent    *Function
	referrers []Instruction
}

type anInstruction struct {
	block *BasicBlock // the basic block of this instruction
}

type Phi struct {
	Comment string  // a hint as to its purpose
	Edges   []Value // Edges[i] is value for Block().Preds[i]
}

type Local struct {
	Comment string
	Value   Value
}

type Global struct {
	Comment string
	Value   string
}

type Assign struct {
	anInstruction
	Lhs Value
	Rhs Value
}

type CompoundAssign struct {
	anInstruction
	Op  string
	Lhs Value
	Rhs Value
}

type Return struct {
	//TODO
}

type While struct {
	anInstruction
	Cond Value
	Body *BasicBlock
	Done *BasicBlock
}

type NumberFor struct {
	anInstruction
	Local Local

	Init  Value
	Limit Value
	Step  Value

	Body *BasicBlock
	Done *BasicBlock
}

type GenericFor struct {
	anInstruction
	Locals []Local
	Values []Value

	Body *BasicBlock
	Done *BasicBlock
}

type If struct {
	anInstruction
	Cond  Value
	True  *BasicBlock
	False *BasicBlock
}

type Call struct {
	anInstruction
	Args   []Value
	Func   Value
	Method string
	Recv   Value
}

// Expressions

type Nil struct{}
type False struct{}
type True struct{}
type VarArg struct{}

type Number struct {
	Value float64
}

type String struct {
	Value string
}

type Table struct {
	array     []Value
	hash      map[Value]Value
	metaTable *Table
}

type AttrGet struct {
	Object Value
	Key    Value
}

type Arithmetic struct {
	Op  string
	Lhs Value
	Rhs Value
}

type Unary struct {
	Op    string
	Value Value
}

type Concat struct {
	Lhs Value
	Rhs Value
}

type Relation struct {
	Op  string
	Lhs Value
	Rhs Value
}

type Logic struct {
	Op  string
	Lhs Value
	Rhs Value
}

func (s *Call) Value() *Call { return s }

func (v *Function) Parent() *Function { return v.parent }
func (v *Function) Referrers() *[]Instruction {
	if v.parent != nil {
		return &v.referrers
	}
	return nil
}

func (v *Parameter) Name() string              { return v.name }
func (v *Parameter) Referrers() *[]Instruction { return &v.referrers }
func (v *Parameter) Parent() *Function         { return v.parent }

func (v *anInstruction) Parent() *Function          { return v.block.parent }
func (v *anInstruction) Block() *BasicBlock         { return v.block }
func (v *anInstruction) setBlock(block *BasicBlock) { v.block = block }
func (v *anInstruction) Referrers() *[]Instruction  { return nil }

func (v *BinOp) Operands(rands []*Value) []*Value {
	return append(rands, &v.X, &v.Y)
}

func (s *Call) Operands(rands []*Value) []*Value {
	return s.Call.Operands(rands)
}

func (v *Phi) Operands(rands []*Value) []*Value {
	for i := range v.Edges {
		rands = append(rands, &v.Edges[i])
	}
	return rands
}

func (s *Return) Operands(rands []*Value) []*Value {
	for i := range s.Results {
		rands = append(rands, &s.Results[i])
	}
	return rands
}

func (v *UnOp) Operands(rands []*Value) []*Value {
	return append(rands, &v.X)
}

// Non-Instruction Values:
func (v *Const) Operands(rands []*Value) []*Value     { return rands }
func (v *Function) Operands(rands []*Value) []*Value  { return rands }
func (v *Parameter) Operands(rands []*Value) []*Value { return rands }
