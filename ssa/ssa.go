package ssa

type Value interface {

}

type Instruction interface {
	String() string
	Parent() *Function
	Block() *BasicBlock
	setBlock(*BasicBlock)
	Operands(rands []*Value) []*Value
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
	name      string

	parent    *Function     // enclosing function if anon; nil if global
	Params    []*Parameter  // function parameters; for methods, includes receiver
	Locals    []*Local      // local variables of this function
	Blocks    []*BasicBlock // basic blocks of the function; nil => external
	referrers []Instruction // referring instructions (iff Parent() != nil)
	continueBlock *BasicBlock
	breakBlock *BasicBlock

	// The following fields are set transiently during building,
	// then cleared.
	currentBlock *BasicBlock             // where to emit code
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

// A Const represents the value of a constant expression.
//
// The underlying type of a constant may be any boolean, numeric, or
// string type.  In addition, a Const may represent the nil value of
// any reference type---interface, map, channel, pointer, slice, or
// function---but not "untyped nil".
//
// All source-level constant expressions are represented by a Const
// of the same type and value.
//
// Value holds the value of the constant, independent of its Type(),
// using go/constant representation, or nil for a typed nil value.
//
// Pos() returns token.NoPos.
//
// Example printed form:
// 	42:int
//	"hello":untyped string
//	3+4i:MyComplex
//
type Const struct {
	Value constant.Value
}

type Phi struct {
	register
	Comment string  // a hint as to its purpose
	Edges   []Value // Edges[i] is value for Block().Preds[i]
}

type Local struct {
	Comment string
	Value Value
}

type Assign struct {
	Lhs VariableInterface
	Rhs Value
}

type CompoundAssign struct {
	Op string
	Lhs VariableInterface
	Rhs Value
}

// The BinOp instruction yields the result of binary operation X Op Y.
//
// Pos() returns the ast.BinaryExpr.OpPos, if explicit in the source.
//
// Example printed form:
// 	t1 = t0 + 1:int
//
type BinOp struct {
	// One of:
	// ADD SUB MUL QUO REM          + - * / %
	// AND OR XOR SHL SHR AND_NOT   & | ^ << >> &^
	// EQL NEQ LSS LEQ GTR GEQ      == != < <= < >=
	Op   token.Token
	X, Y Value
}

// The UnOp instruction yields the result of Op X.
// ARROW is channel receive.
// MUL is pointer indirection (load).
// XOR is bitwise complement.
// SUB is negation.
// NOT is logical negation.
//
// If CommaOk and Op=ARROW, the result is a 2-tuple of the value above
// and a boolean indicating the success of the receive.  The
// components of the tuple are accessed using Extract.
//
// Pos() returns the ast.UnaryExpr.OpPos, if explicit in the source.
// For receive operations (ARROW) implicit in ranging over a channel,
// Pos() returns the ast.RangeStmt.For.
// For implicit memory loads (STAR), Pos() returns the position of the
// most closely associated source-level construct; the details are not
// specified.
//
// Example printed form:
// 	t0 = *x
// 	t2 = <-t1,ok
//
type UnOp struct {
	Op      token.Token // One of: NOT SUB ARROW MUL XOR ! - <- * ^
	X       Value
}



type Return struct {
	//TODO
}

type While struct {
	Condtion Value
	Body *BasicBlock
	Done *BasicBlock
}

type NumberFor struct {
	Local Local

	Init Value
	Limit Value
	Step Value

	Body *BasicBlock
	Done *BasicBlock
}

type GenericFor struct {
	Locals []Local
	Values []Value

	Body *BasicBlock
	Done *BasicBlock
}

type If struct {
	Cond Value
	True *BasicBlock
	False *BasicBlock
}

type Call struct {
	Args []Value
	Func Value
	Method string
	Recv Value
}

type anInstruction struct {
	block *BasicBlock // the basic block of this instruction
}



// Expressions

type Nil struct {}
type False struct {}
type True struct {}
type VarArg struct {}

type Number struct {
	Value float64
}

type String struct {
	Value string
}

type AttrGet struct {
	Object Value
	Key Value
}

type Table struct {}

type Arithmetic struct {
	Op string
	Lhs Value
	Rhs Value
}

type Unary struct {
	Op string
	Value Value
}

type Concat struct {
	Lhs Value
	Rhs Value
}

type Relation struct {
	Op string
	Lhs Value
	Rhs Value
}

type Logic struct {
	Op string
	Lhs Value
	Rhs Value
}



func (s *Call) Value() *Call  { return s }

func (v *Function) Parent() *Function    { return v.parent }
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