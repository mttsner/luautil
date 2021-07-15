package ssa

import (
	"testing"
	"strings"

	"github.com/notnoobmaster/luautil/ast"
	"github.com/notnoobmaster/luautil/parse"
)

func TestBuilder(t *testing.T) {
	const input = `
	print("Hello World!")
	`
	
	chunk, err := parse.Parse(strings.NewReader(input), "")
	if err != nil {
		t.Fatal(err)
	}

	var b builder

	fn := &Function{
		name: "main",
		syntax: &ast.FunctionExpr{Chunk: chunk},
	}
	b.buildFunction(fn)

}