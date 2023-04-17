package ssa

import (
	"strings"
	"testing"
)

func TestToAst(t *testing.T) {
	const input = `
	local t0, t1 = 0, 1
	while t0 do t0 = t1 end
	` 

	fn := build(input, t)
	b := &strings.Builder{}
	MarkUnreachableBlocks(fn)
	ast := fn.Chunk()
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())
	t.Error("\n"+ast.String())
}