package ssa

import (
	"strings"
	"testing"
)

func TestToAst(t *testing.T) {
	const input = `
	local t0 = 0

	if global then
		t0 = 1
	end

	print(t0+1)
	`

	fn := build(input, t)
	b := &strings.Builder{}

	buildReferrers(fn)
	buildDomTree(fn)

	BuildDomFrontier(fn)
	MarkUnreachableBlocks(fn)

	lift(fn) 

	ast := fn.Chunk()
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())
	t.Error("\n"+ast.String())
}