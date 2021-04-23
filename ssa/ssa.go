package ssa

type Value interface {
	Name() string
	String() string
	Parent() *Function
	Referrers() *[]Instruction
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

	Synthetic string        // provenance of synthetic function; "" for true source functions
	syntax    ast.Node      // *ast.Func{Decl,Lit}; replaced with simple ast.Node after build, unless debug mode
	parent    *Function     // enclosing function if anon; nil if global
	Params    []*Parameter  // function parameters; for methods, includes receiver
	Locals    []*Alloc      // local variables of this function
	Blocks    []*BasicBlock // basic blocks of the function; nil => external

	AnonFuncs []*Function   // anonymous functions directly beneath this one
	referrers []Instruction // referring instructions (iff Parent() != nil)

	// The following fields are set transiently during building,
	// then cleared.
	currentBlock *BasicBlock             // where to emit code
	objects      map[types.Object]Value  // addresses of local variables
	namedResults []*Alloc                // tuple of named results
	targets      *targets                // linked stack of branch targets
	lblocks      map[*ast.Object]*lblock // labelled blocks
}

type BasicBlock struct {
	Index        int            // index of this block within Parent().Blocks
	Comment      string         // optional label; no semantic significance
	parent       *Function      // parent function
	Instrs       []Instruction  // instructions in order
	Preds, Succs []*BasicBlock  // predecessors and successors
	succs2       [2]*BasicBlock // initial space for Succs
	dom          domInfo        // dominator tree info
	gaps         int            // number of nil Instrs (transient)
	rundefers    int            // number of rundefers (transient)
}

type Parameter struct {
	name      string
	parent    *Function
	referrers []Instruction
}

type Phi struct {
	register
	Comment string  // a hint as to its purpose
	Edges   []Value // Edges[i] is value for Block().Preds[i]
}

type Call struct {
	register
	Call CallCommon
}

type BinOp struct {
	register
	// One of:
	// ADD SUB MUL QUO REM          + - * / %
	// AND OR XOR SHL SHR AND_NOT   & | ^ << >> &^
	// EQL NEQ LSS LEQ GTR GEQ      == != < <= < >=
	Op   token.Token
	X, Y Value
}

type UnOp struct {
	register
	Op      token.Token // One of: NOT SUB ARROW MUL XOR ! - <- * ^
	X       Value
	CommaOk bool
}

type If struct {
	anInstruction
	Cond Value
}

type Return struct {
	anInstruction
	Results []Value
}

type register struct {
	anInstruction
	num       int        // "name" of virtual register, e.g. "t0".  Not guaranteed unique.
	referrers []Instruction
}

type anInstruction struct {
	block *BasicBlock // the basic block of this instruction
}

