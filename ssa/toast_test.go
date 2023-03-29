package ssa

import (
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
	
	repeat
		t0 = 1
	until true end
	` 

	fn := build(input, t)
	t.Error(fn.String())

	fn.Chunk()
}