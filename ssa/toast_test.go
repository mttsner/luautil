package ssa

import (
	"strings"
	"testing"
)

func TestToAst(t *testing.T) {
	const input = `
	local t0 = 0

	for i,v in pairs(t0) do
		t0 = 1
		if t0 then
			break
		end
	end

	t0 = 2
	` 

	fn := build(input, t)
	b := &strings.Builder{}
	deleteUnreachableBlocks(fn)
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())
	ast := fn.Chunk()
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())
	t.Error("\n"+ast.String())
}