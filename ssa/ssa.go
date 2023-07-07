package ssa

import (
	"github.com/notnoobmaster/luautil/ast"
)

type Value interface {
	String() string
	Referrers() *[]Instruction
}

type Instruction interface {
	String() string
	Parent() *Function
	Block() *BasicBlock
	SetBlock(*BasicBlock)
	Operands(rands []*Value) []*Value
	//Equal(Instruction) bool
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
	Name      string
	Params    []*Local // function parameters; for methods, includes receiver
	Locals    []*Local
	UpValues  []*Local
	Functions []*Function   // nested functions defined inside this one
	Blocks    []*BasicBlock // basic blocks of the function; nil => external
	VarArg    bool

	DomFrontier

	syntax        *ast.FunctionExpr
	parent        *Function     // enclosing function if anon; nil if global
	referrers     []Instruction // referring instructions (iff Parent() != nil)
	continueBlock *BasicBlock
	breakBlock    *BasicBlock
	num           int

	// The following fields are set transiently during building,
	// then cleared.
	currentScope *Scope      // lexical scope of this function
	currentBlock *BasicBlock // where to emit code
}

type Scope struct {
	function *Function
	parent   *Scope
	names    map[string]*Local
}

type BasicBlock struct {
	Index            int            // index of this block within Parent().Blocks
	Comment          string         // optional label; no semantic significance
	parent           *Function      // parent function
	Instrs           []Instruction  // instructions in order
	Preds, Succs     []*BasicBlock  // predecessors and successors
	succs2           [2]*BasicBlock // initial space for Succs
	gaps             int            // number of nil Instrs (transient)
	dom              domInfo        // dominator tree info
	reachable        bool
	UnPreds, UnSuccs []*BasicBlock // unreachable predecessors and successors
}

type anInstruction struct {
	block *BasicBlock // the basic block of this instruction
}

type Jump struct {
	anInstruction
}

type Phi struct {
	anInstruction
	Comment string  // a hint as to its purpose
	Edges   []Value // Edges[i] is value for Block().Preds[i]
}

type Local struct {
	Comment string
	Value   Value
	Index   int
	Name string // Used as the variable name in AST

	declared  bool
	upvalue   bool
	def *BasicBlock
	parent *Function
	referrers []Instruction
}

type Global struct {
	Comment   string
	Value     Value
	referrers []Instruction
}

type Assign struct {
	anInstruction
	Lhs []Value
	Rhs []Value
}

type CompoundAssign struct {
	anInstruction
	Op  string
	Lhs []Value
	Rhs []Value
}

type Return struct {
	anInstruction
	Values []Value
}

type NumberFor struct {
	anInstruction
	Local Value
	Init  Value
	Limit Value
	Step  Value
}

type GenericFor struct {
	anInstruction
	Locals []Value
	Values []Value
}

type If struct {
	anInstruction
	Cond Value
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

type True struct{}

type False struct{}

type Number struct {
	Value float64
}

type String struct {
	Value string
}

type VarArg struct{}

type Field struct {
	Key   Value
	Value Value
}

type Table struct {
	Fields []*Field
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
	anInstruction
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

func (v *anInstruction) Parent() *Function          { return v.block.parent }
func (v *anInstruction) Block() *BasicBlock         { return v.block }
func (v *anInstruction) SetBlock(block *BasicBlock) { v.block = block }
func (v *anInstruction) Referrers() *[]Instruction  { return nil }

// Non-Instruction Values:
// func (v *Const) Operands(rands []*Value) []*Value    { return rands }
func (v *Function) Operands(rands []*Value) []*Value { return rands }

func (v *Local) MarkAsUpvalue() { v.upvalue = true }
// DefBlock returns the basicblock the local was defined in
func (v *Local) DefBlock() *BasicBlock { return v.def}
func (v *Local) Parent() *Function { return v.parent }

func (v *Jump) Operands(rands []*Value) []*Value { return rands }
func (v *Phi) Operands(rands []*Value) []*Value {
	for i := range v.Edges {
		rands = append(rands, &v.Edges[i])
	}
	return rands
}

func (v Nil) Referrers() *[]Instruction        { return nil }
func (v True) Referrers() *[]Instruction       { return nil }
func (v False) Referrers() *[]Instruction      { return nil }
func (v Number) Referrers() *[]Instruction     { return nil }
func (v String) Referrers() *[]Instruction     { return nil }
func (v VarArg) Referrers() *[]Instruction     { return nil }
func (v Field) Referrers() *[]Instruction      { return nil }
func (v Table) Referrers() *[]Instruction      { return nil }
func (v AttrGet) Referrers() *[]Instruction    { return nil }
func (v Arithmetic) Referrers() *[]Instruction { return nil }
func (v Unary) Referrers() *[]Instruction      { return nil }
func (v Concat) Referrers() *[]Instruction     { return nil }
func (v Relation) Referrers() *[]Instruction   { return nil }
func (v Logic) Referrers() *[]Instruction      { return nil }

func (v *Local) Referrers() *[]Instruction        { return &v.referrers }
func (v *Local) Operands(rands []*Value) []*Value { return rands }

func (v *Global) Referrers() *[]Instruction        { return &v.referrers }
func (v *Global) Operands(rands []*Value) []*Value { return rands }

func (v *CompoundAssign) Operands(rands []*Value) []*Value { return rands }
func (v *Return) Operands(rands []*Value) []*Value         { return rands }
func (v *NumberFor) Operands(rands []*Value) []*Value      { return rands }
func (v *GenericFor) Operands(rands []*Value) []*Value     { return rands }
func (v *If) Operands(rands []*Value) []*Value             { return append(rands, &v.Cond) }

// func (v *Assign) Referrers() *[]Instruction { return &v.Local.referrers }
func (v *Assign) Operands(rands []*Value) []*Value {
	for _, l := range v.Lhs{
		rands = append(rands, &l)
	}
	return rands
}

func (v Call) Referrers() *[]Instruction { return nil }
func (v *Call) Operands(rands []*Value) []*Value {
	for _, val := range v.Args {
		rands = append(rands, &val)
	}
	return rands
}
