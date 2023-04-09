package ssa

import (
	"strings"
	"testing"
)

func TestToAst(t *testing.T) {
	const input = `
	local t0 = 0

	for i,v in pairs(t0) do
		break
	end
	` 

	fn := build(input, t)
	b := &strings.Builder{}
	ast := fn.Chunk()
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())
	t.Error("\n"+ast.String())
}