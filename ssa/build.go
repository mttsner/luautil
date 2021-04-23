package ssa

import "github.com/notnoobmaster/beautifier/ast"

type builder struct {}

func (b *builder) stmt(fn *Function, _s ast.Stmt) {
	switch s := _s.(type) {
	case *ast.LocalAssignStmt:
		
	}
}

func (b *builder) stmtList(fn *Function, chunk []ast.Stmt) {
	for _, s := range chunk {
		b.stmt(fn, s)
	}
}