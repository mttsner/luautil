package ast

type Stmt interface {
	PositionHolder
	stmtMarker()
	String() string
}

type StmtBase struct {
	Node
}

func (stmt *StmtBase) stmtMarker() {}

type AssignStmt struct {
	StmtBase

	Lhs []Expr
	Rhs []Expr
}

type CompoundAssignStmt struct {
	StmtBase

	Operator string
	Lhs []Expr
	Rhs []Expr
}

type LocalAssignStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
}

type FuncCallStmt struct {
	StmtBase

	Expr Expr
}

type DoBlockStmt struct {
	StmtBase

	Chunk Chunk
}

type WhileStmt struct {
	StmtBase

	Condition Expr
	Chunk     Chunk
}

type RepeatStmt struct {
	StmtBase

	Condition Expr
	Chunk     Chunk
}

type IfStmt struct {
	StmtBase

	Condition Expr
	Then      Chunk
	Else      Chunk
}

type NumberForStmt struct {
	StmtBase

	Name  string
	Init  Expr
	Limit Expr
	Step  Expr
	Chunk Chunk
}

type GenericForStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
	Chunk Chunk
}

type LocalFunctionStmt struct {
	StmtBase

	Name string
	Func *FunctionExpr
}

type FunctionStmt struct {
	StmtBase

	Name *FuncName
	Func *FunctionExpr
}

type ReturnStmt struct {
	StmtBase

	Exprs []Expr
}

type BreakStmt struct {
	StmtBase
}

type ContinueStmt struct {
	StmtBase
}

type LabelStmt struct {
	StmtBase

	Name string
}

type GotoStmt struct {
	StmtBase

	Label string
}