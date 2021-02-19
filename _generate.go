package beautifier

import (
	"bytes"
	"github.com/yuin/gopher-lua/ast"
)

const (
	// Stmt
	doBlockStmt byte = iota
	repeatStmt
	whileStmt
	assignStmt
	funcDefStmt
	funcCallStmt
	numberForStmt
	genericForStmt
	ifStmt
	elseStmt
	breakStmt
	returnStmt
	localAssignStmt
	// Expr
	stringExpr
	numberExpr
	valueExpr
	nilExpr
	falseExpr
	trueExpr
	identExpr
	comma3Expr
	attrGetExpr
	tableExpr
	arithmeticExpr
	stringConcatExpr
	minusExpr
	notExpr
	lenExpr
	relationalExpr
	logicalExpr
	funcCallExpr
	functionExpr
	// Special stuff
	varArg
)

type state struct {
	Hash *bytes.Buffer
	Variables map[string]byte
}

//GenerateHash generates a tree pattern
func GenerateHash(chunk []ast.Stmt, variables map[string]byte) string {
	s := &state{
		Hash: new(bytes.Buffer),
		Variables: variables,
	}
	s.traverseStmts(chunk)
	return s.Hash.String()
}

func (s *state) traverseExprs(exprs []ast.Expr) {
	for _, ex := range exprs {
		s.traverseExpr(ex)
	}
}

func (s *state) traverseExpr(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.StringExpr:
		s.Hash.WriteByte(stringExpr)
	case *ast.NumberExpr:
		s.Hash.WriteByte(numberExpr)
	case *ast.NilExpr:
		s.Hash.WriteByte(nilExpr)
	case *ast.FalseExpr:
		s.Hash.WriteByte(falseExpr)
	case *ast.TrueExpr:
		s.Hash.WriteByte(trueExpr)
	case *ast.Comma3Expr:
		s.Hash.WriteByte(comma3Expr)
	case *ast.AttrGetExpr:
		s.Hash.WriteByte(attrGetExpr)
		s.traverseExpr(ex.Object)
		s.traverseExpr(ex.Key)
	case *ast.ArithmeticOpExpr:
		s.Hash.WriteByte(arithmeticExpr)
		s.Hash.WriteString(ex.Operator)
		s.traverseExpr(ex.Lhs)
		s.traverseExpr(ex.Rhs)
	case *ast.StringConcatOpExpr:
		s.Hash.WriteByte(stringConcatExpr)
		s.traverseExpr(ex.Lhs)
		s.traverseExpr(ex.Rhs)
	case *ast.UnaryMinusOpExpr:
		s.Hash.WriteByte(minusExpr)
		s.traverseExpr(ex.Expr)
	case *ast.UnaryNotOpExpr:
		s.Hash.WriteByte(notExpr)
		s.traverseExpr(ex.Expr)
	case *ast.UnaryLenOpExpr:
		s.Hash.WriteByte(lenExpr)
		s.traverseExpr(ex.Expr)
	case *ast.RelationalOpExpr:
		s.Hash.WriteByte(relationalExpr)
		s.Hash.WriteString(ex.Operator)
		s.traverseExpr(ex.Lhs)
		s.traverseExpr(ex.Rhs)
	case *ast.LogicalOpExpr:
		s.Hash.WriteByte(logicalExpr)
		s.Hash.WriteString(ex.Operator)
		s.traverseExpr(ex.Lhs)
		s.traverseExpr(ex.Rhs)
	case *ast.IdentExpr:
		if v, ok := s.Variables[ex.Value]; ok {
			s.Hash.WriteByte(v)
		} else {
			s.Hash.WriteByte(identExpr)
		}
	case *ast.TableExpr:
		s.Hash.WriteByte(tableExpr)
		for _, field := range ex.Fields {
			if field.Key != nil {
				s.traverseExpr(field.Key)
			}
			s.traverseExpr(field.Value)
		}
	case *ast.FuncCallExpr:
		s.Hash.WriteByte(funcCallExpr)
		if ex.Func != nil {
			s.traverseExpr(ex.Func)
		} else {
			s.traverseExpr(ex.Receiver)
		}
		s.traverseExprs(ex.Args)
	case *ast.FunctionExpr:
		s.Hash.WriteByte(functionExpr)
		for range ex.ParList.Names {
			s.Hash.WriteByte(identExpr)
		}
		if ex.ParList.HasVargs {
			s.Hash.WriteByte(varArg)
		}
		s.traverseStmts(ex.Stmts)
	}
}

func (s *state) traverseStmts(chunk []ast.Stmt) {
	for _, st := range chunk {
		switch stmt := st.(type) {
		case *ast.AssignStmt:
			s.Hash.WriteByte(assignStmt)
			s.traverseExprs(stmt.Lhs)
			s.traverseExprs(stmt.Rhs)
		case *ast.LocalAssignStmt:
			s.Hash.WriteByte(localAssignStmt)
			s.traverseExprs(stmt.Exprs)
		case *ast.FuncCallStmt:
			s.Hash.WriteByte(funcCallExpr)
			ex := stmt.Expr.(*ast.FuncCallExpr)
			if ex.Func != nil {
				s.traverseExpr(ex.Func)
			} else {
				s.traverseExpr(ex.Receiver)
			}
			s.traverseExprs(ex.Args)
		case *ast.DoBlockStmt:
			s.Hash.WriteByte(doBlockStmt)
			s.traverseStmts(stmt.Stmts)
		case *ast.WhileStmt:
			s.Hash.WriteByte(whileStmt)
			s.traverseExpr(stmt.Condition)
			s.traverseStmts(stmt.Stmts)
		case *ast.RepeatStmt:
			s.Hash.WriteByte(repeatStmt)
			s.traverseExpr(stmt.Condition)
			s.traverseStmts(stmt.Stmts)
		case *ast.FuncDefStmt:
			s.Hash.WriteByte(funcDefStmt)
			for range stmt.Func.ParList.Names {
				s.Hash.WriteByte(identExpr)
			}
			if stmt.Func.ParList.HasVargs {
				s.Hash.WriteByte(varArg)
			}
			s.traverseStmts(stmt.Func.Stmts)
		case *ast.ReturnStmt:
			s.Hash.WriteByte(returnStmt)
			s.traverseExprs(stmt.Exprs)
		case *ast.IfStmt:
			s.Hash.WriteByte(ifStmt)
			s.traverseExpr(stmt.Condition)
			s.traverseStmts(stmt.Then)
			if len(stmt.Else) > 0 {
				s.traverseStmts(stmt.Else)
			}
		case *ast.BreakStmt:
			s.Hash.WriteByte(breakStmt)
		case *ast.NumberForStmt:
			s.Hash.WriteByte(numberForStmt)
			s.traverseExpr(stmt.Init)
			s.traverseExpr(stmt.Limit)
			if stmt.Step != nil {
				s.traverseExpr(stmt.Step)
			}
			s.traverseStmts(stmt.Stmts)
		case *ast.GenericForStmt:
			s.Hash.WriteByte(genericForStmt)
			s.traverseExprs(stmt.Exprs)
			s.traverseStmts(stmt.Stmts)
		}
	}
}
