package ssa

import (
	"strings"
	"testing"
)

// if ifelse repeat while
/*
local a = 0

	if true then
		a = 1
	end

	if false then
		a = 2
	else
		a = 3
	end

	while true do
		a = 4
	end

	repeat
		a = 5
	until false end
*/

func TestToAst(t *testing.T) {
	const input = `
	local t0 = 0
	
	if t0 then
		t0 = 3
		if t0 then
			t0 = 1
		else
			t0 = 4
		end
	end
	t0 = 2
	` 

	fn := build(input, t)
	b := &strings.Builder{}
	WriteCfgDot(b, fn)
	t.Log(b.String())
	t.Log(fn.String())

	ast := fn.Chunk()
	t.Error("\n"+ast.String())
}