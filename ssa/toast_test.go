package ssa

import (
	"testing"
)

func TestToAst(t *testing.T) {
	const input = `
	local a = 1
	` 

	fn := build(input, t)
	t.Error(fn.String())

	chunk := fn.Chunk()
	t.Error("\n" + chunk.String())
}