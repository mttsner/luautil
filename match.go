package beautifier

import (
	"github.com/yuin/gopher-lua/ast" 

)

type state struct {
	Pattern []ast.Stmt
	Exprs []ast.Expr
	Exit bool
}

func (s *state) quickTraverseExpr(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.AttrGetExpr:
		s.quickTraverseExpr(ex.Object)
		s.quickTraverseExpr(ex.Key)
	case *ast.ArithmeticOpExpr:
		s.quickTraverseExpr(ex.Lhs)
		s.quickTraverseExpr(ex.Rhs)
	case *ast.StringConcatOpExpr:
		s.quickTraverseExpr(ex.Lhs)
		s.quickTraverseExpr(ex.Rhs)
	case *ast.RelationalOpExpr:
		s.quickTraverseExpr(ex.Lhs)
		s.quickTraverseExpr(ex.Rhs)
	case *ast.LogicalOpExpr:
		s.quickTraverseExpr(ex.Lhs)
		s.quickTraverseExpr(ex.Rhs)
	case *ast.UnaryMinusOpExpr:
		s.quickTraverseExpr(ex.Expr)
	case *ast.UnaryNotOpExpr:
		s.quickTraverseExpr(ex.Expr)
	case *ast.UnaryLenOpExpr:
		s.quickTraverseExpr(ex.Expr)
	case *ast.FunctionExpr:
		if s.match(ex.Stmts) {
			s.Exit = true // KINDA FUCKING HACKY
		}
	case *ast.TableExpr:
		for _, field := range ex.Fields {
			if field.Key != nil {
				s.quickTraverseExpr(field.Key)
			}
			s.quickTraverseExpr(field.Value)
		}
	case *ast.FuncCallExpr:
		if ex.Func != nil {
			s.quickTraverseExpr(ex.Func)
		} else {
			s.quickTraverseExpr(ex.Receiver)
		}
		s.quickTraverseExprs(ex.Args)
	}
}

func (s *state) exprEqual(expr ast.Expr, selector ast.Expr) bool {
	switch ex := expr.(type) {
	case *ast.StringExpr:
		if _, ok := selector.(*ast.StringExpr); ok {
			return true
		}
	case *ast.NumberExpr:
		if _, ok := selector.(*ast.NumberExpr); ok {
			return true
		} else if custom, ok := selector.(*ast.IdentExpr); ok && custom.Value == "_NumberExpr_" {
			s.Exprs = append(s.Exprs, ex)
			return true
		}
	case *ast.NilExpr:
		if _, ok := selector.(*ast.NilExpr); ok {
			return true
		}
	case *ast.FalseExpr:
		if _, ok := selector.(*ast.FalseExpr); ok {
			return true
		}
	case *ast.TrueExpr:
		if _, ok := selector.(*ast.TrueExpr); ok {
			return true
		}
	case *ast.Comma3Expr:
		if _, ok := selector.(*ast.Comma3Expr); ok {
			return true
		}
	case *ast.IdentExpr:
		if ident, ok := selector.(*ast.IdentExpr); ok {
			if ident.Value == "_IdentExpr_" {	
				s.Exprs = append(s.Exprs, ex)
			}
			return true
		}
	case *ast.UnaryMinusOpExpr:
		if unary, ok := selector.(*ast.UnaryMinusOpExpr); ok {
			return s.exprEqual(ex.Expr, unary.Expr)
		} 
	case *ast.UnaryNotOpExpr:
		if not, ok := selector.(*ast.UnaryNotOpExpr); ok {
			return s.exprEqual(ex.Expr, not.Expr)
		} 
	case *ast.UnaryLenOpExpr:
		if unaryLen, ok := selector.(*ast.UnaryLenOpExpr); ok {
			return s.exprEqual(ex.Expr, unaryLen.Expr)
		} 
	case *ast.ArithmeticOpExpr:
		if arith, ok := selector.(*ast.ArithmeticOpExpr); ok && ex.Operator == arith.Operator {
			return s.exprEqual(ex.Lhs, arith.Lhs) && s.exprEqual(ex.Rhs, arith.Rhs)
		}
	case *ast.RelationalOpExpr:
		if rel, ok := selector.(*ast.RelationalOpExpr); ok && ex.Operator == rel.Operator {
			return s.exprEqual(ex.Lhs, rel.Lhs) && s.exprEqual(ex.Rhs, rel.Rhs)
		}
	case *ast.LogicalOpExpr:
		if logic, ok := selector.(*ast.LogicalOpExpr); ok && ex.Operator == logic.Operator {
			return s.exprEqual(ex.Lhs, logic.Lhs) && s.exprEqual(ex.Rhs, logic.Rhs)
		}
	case *ast.AttrGetExpr:
		if attr, ok := selector.(*ast.AttrGetExpr); ok {
			return s.exprEqual(ex.Object, attr.Object) && s.exprEqual(ex.Key, attr.Key)
		}
	case *ast.StringConcatOpExpr:
		if str, ok := selector.(*ast.StringConcatOpExpr); ok {
			return s.exprEqual(ex.Lhs, str.Lhs)  && s.exprEqual(ex.Rhs, str.Rhs)
		}
	case *ast.TableExpr: // TODO Frankly in-depth table comparison is useless
		if _, ok := selector.(*ast.TableExpr); ok {
			return true
		}
	case *ast.FuncCallExpr:
		if f, ok := selector.(*ast.FuncCallExpr); ok {
			if ex.Func != nil && f.Func != nil {
				return ex.AdjustRet == f.AdjustRet && s.exprEqual(ex.Func, f.Func) && s.exprsEqual(ex.Args, f.Args)
			}
			if ex.Receiver != nil && f.Receiver != nil {
				return ex.AdjustRet == f.AdjustRet && s.exprEqual(ex.Receiver, f.Receiver) && s.exprsEqual(ex.Args, f.Args)
			}
		}
	case *ast.FunctionExpr:
		if f, ok := selector.(*ast.FunctionExpr); ok {
			if ex.ParList.HasVargs == f.ParList.HasVargs && 
			len(ex.ParList.Names) == len(f.ParList.Names) && 
			s.stmtsEqual(ex.Stmts, f.Stmts) {
				for i, name := range f.ParList.Names {
					if name == "_IdentExpr_" {
						s.Exprs = append(s.Exprs, &ast.IdentExpr{Value: ex.Parlist.Name[i])
					}
				}
				return true
			}
		}
	}
	return false
}

func (s *state) exprsEqual(exprs []ast.Expr, selector []ast.Expr) bool {
	for idx, expr := range exprs {
		if !s.exprEqual(expr, selector[idx]) {
			return false
		}
	}
	return true
}

func (s *state) quickTraverseExprs(exprs []ast.Expr) {
	for _, ex := range exprs {
		s.quickTraverseExpr(ex)
	}
}

// Equality functions

func (s *state) assignEqual(first *ast.AssignStmt, second *ast.AssignStmt) bool {
	return len(first.Lhs) == len(second.Lhs) &&
		len(first.Rhs) == len(second.Rhs) &&
		s.exprsEqual(first.Lhs, second.Lhs) &&
		s.exprsEqual(first.Rhs, second.Rhs)
}

func (s *state) localAssignEqual(first *ast.LocalAssignStmt, second *ast.LocalAssignStmt) bool {
	return len(second.Names) == len(first.Names) && 
		len(second.Exprs) == len(first.Exprs) && 
		s.exprsEqual(first.Exprs, second.Exprs)
}

func (s *state) funcCallEqual(first *ast.FuncCallStmt, second *ast.FuncCallStmt) bool {
	return s.exprEqual(first.Expr, second.Expr)
}

func (s *state) doBlockEqual(first *ast.DoBlockStmt, second *ast.DoBlockStmt) bool {
	return s.stmtsEqual(first.Stmts, second.Stmts)
}

func (s *state) whileEqual(first *ast.WhileStmt, second *ast.WhileStmt) bool {
	return s.exprEqual(first.Condition, second.Condition) &&
		s.stmtsEqual(first.Stmts, second.Stmts)
}

func (s *state) repeatEqual(first *ast.RepeatStmt, second *ast.RepeatStmt) bool {
	return s.exprEqual(first.Condition, second.Condition) && 
		s.stmtsEqual(first.Stmts, second.Stmts)
}

func (s *state) funcDefEqual(first *ast.FuncDefStmt, second *ast.FuncDefStmt) bool {
	// TODO add FuncName equality
	return s.exprEqual(first.Func, second.Func)
}

func (s *state) returnEqual(first *ast.ReturnStmt, second *ast.ReturnStmt) bool {
	return len(first.Exprs) == len(second.Exprs) &&
		s.exprsEqual(first.Exprs, second.Exprs)
}

func (s *state) ifEqual(first *ast.IfStmt, second *ast.IfStmt) bool {
	return s.exprEqual(first.Condition, second.Condition) && 
		s.stmtsEqual(first.Then, second.Then) &&
		s.stmtsEqual(first.Else, second.Else)
}

func (s *state) numberForEqual(first *ast.NumberForStmt, second *ast.NumberForStmt) bool {
	return first.Step == second.Step &&
		len(first.Name) == len(second.Name) &&
		s.stmtsEqual(first.Stmts, second.Stmts)
}

func (s *state) genericForEqual(first *ast.GenericForStmt, second *ast.GenericForStmt) bool {
	return len(first.Names) == len(second.Names) &&
		len(first.Exprs) == len(second.Exprs) &&
		s.exprsEqual(first.Exprs, second.Exprs) &&
		s.stmtsEqual(first.Stmts, second.Stmts)
}

func (s *state) stmtsEqual(chunk []ast.Stmt, pattern []ast.Stmt) bool {
	if len(pattern) == 0 {
		return true
	}

	if len(chunk) != len(pattern) {
		return false
	}

	for pos, st := range chunk {
		cStmt := pattern[pos]

		switch stmt := st.(type) {
		case *ast.AssignStmt:
			if result, ok := cStmt.(*ast.AssignStmt); ok && s.assignEqual(stmt, result) {
				break
			}
			return false
		case *ast.LocalAssignStmt:
			if result, ok := cStmt.(*ast.LocalAssignStmt); ok && s.localAssignEqual(stmt, result) {
				break
			}
			return false
		case *ast.FuncCallStmt:
			if result, ok := cStmt.(*ast.FuncCallStmt); ok && s.funcCallEqual(stmt, result) {
				break
			}
			return false
		case *ast.DoBlockStmt:
			if result, ok := cStmt.(*ast.DoBlockStmt); ok && s.doBlockEqual(stmt, result) {
				break
			}
			return false
		case *ast.WhileStmt:
			if result, ok := cStmt.(*ast.WhileStmt); ok && s.whileEqual(stmt, result) {
				break
			}
			return false
		case *ast.RepeatStmt:
			if result, ok := cStmt.(*ast.RepeatStmt); ok && s.repeatEqual(stmt, result) {
				break
			}
			return false
		case *ast.FuncDefStmt:
			if result, ok := cStmt.(*ast.FuncDefStmt); ok && s.funcDefEqual(stmt, result) {
				break
			}
			return false
		case *ast.ReturnStmt:
			if result, ok := cStmt.(*ast.ReturnStmt); ok && s.returnEqual(stmt, result) {
				break
			}
			return false
		case *ast.IfStmt:
			if result, ok := cStmt.(*ast.IfStmt); ok && s.ifEqual(stmt, result) {
				break
			}
			return false
		case *ast.BreakStmt:
			if _, ok := cStmt.(*ast.BreakStmt); !ok {
				return false
			}
		case *ast.NumberForStmt:
			if result, ok := cStmt.(*ast.NumberForStmt); ok && s.numberForEqual(stmt, result) {
				break
			}
			return false
		case *ast.GenericForStmt:
			if result, ok := cStmt.(*ast.GenericForStmt); ok && s.genericForEqual(stmt, result) {
				break
			}
			return false
		}
	}
	return true
}

func (s *state) match(chunk []ast.Stmt) (success bool) {
	pos, pLen := 0, len(s.Pattern)
	fStmt := s.Pattern[pos]

	for _, st := range chunk {
		cStmt := s.Pattern[pos]
		success = false

		switch stmt := st.(type) {
		case *ast.AssignStmt:
			if result, ok := cStmt.(*ast.AssignStmt); ok && s.assignEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.AssignStmt); ok && s.assignEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExprs(stmt.Lhs)
				s.quickTraverseExprs(stmt.Rhs)
			}
		case *ast.LocalAssignStmt:
			if result, ok := cStmt.(*ast.LocalAssignStmt); ok && s.localAssignEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.LocalAssignStmt); ok && s.localAssignEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExprs(stmt.Exprs)
			}
		case *ast.FuncCallStmt:
			if result, ok := cStmt.(*ast.FuncCallStmt); ok && s.funcCallEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.FuncCallStmt); ok && s.funcCallEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExpr(stmt.Expr)
			}
		case *ast.DoBlockStmt:
			if result, ok := cStmt.(*ast.DoBlockStmt); ok && s.doBlockEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.DoBlockStmt); ok && s.doBlockEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				if s.match(stmt.Stmts){
					return true
				}
			}
		case *ast.WhileStmt:
			if result, ok := cStmt.(*ast.WhileStmt); ok && s.whileEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.WhileStmt); ok && s.whileEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExpr(stmt.Condition)
				if s.match(stmt.Stmts){
					return true
				}
			}
		case *ast.RepeatStmt:
			if result, ok := cStmt.(*ast.RepeatStmt); ok && s.repeatEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.RepeatStmt); ok && s.repeatEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExpr(stmt.Condition)
				if s.match(stmt.Stmts){
					return true
				}	
			}
		case *ast.FuncDefStmt:
			if result, ok := cStmt.(*ast.FuncDefStmt); ok && s.funcDefEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.FuncDefStmt); ok && s.funcDefEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExpr(stmt.Func)
			}
		case *ast.ReturnStmt:
			if result, ok := cStmt.(*ast.ReturnStmt); ok && s.returnEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.ReturnStmt); ok && s.returnEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExprs(stmt.Exprs)
			}
		case *ast.IfStmt:
			if result, ok := cStmt.(*ast.IfStmt); ok && s.ifEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.IfStmt); ok && s.ifEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExpr(stmt.Condition)
				s.match(stmt.Then)
				s.match(stmt.Else)
			}
		case *ast.BreakStmt:
			if _, ok := cStmt.(*ast.BreakStmt); ok {
				success = true
			} else if _, ok := fStmt.(*ast.BreakStmt); ok {
				pos = 0
				success = true
			}
		case *ast.NumberForStmt:
			if result, ok := cStmt.(*ast.NumberForStmt); ok && s.numberForEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.NumberForStmt); ok && s.numberForEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.match(stmt.Stmts)
			}
		case *ast.GenericForStmt:
			if result, ok := cStmt.(*ast.GenericForStmt); ok && s.genericForEqual(stmt, result) {
				success = true
			} else if result, ok := fStmt.(*ast.GenericForStmt); ok && s.genericForEqual(stmt, result) {
				pos = 0
				success = true
			} else {
				s.quickTraverseExprs(stmt.Exprs)
				s.match(stmt.Stmts)
			}
		}


		if s.Exit {
			return true
		}

		if success {
			pos++
		} else {
			pos = 0
		}

		if pos == pLen {
			return true
		}
	}
	return false
}

// Match pattern in ast.
func Match(chunk []ast.Stmt, pattern []ast.Stmt) (bool, []ast.Expr) {
	st := state{Pattern: pattern}
	return st.match(chunk), st.Exprs
}
