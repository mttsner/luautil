package ssa

import (
	"testing"
)

// if ifelse repeat while

func TestToAst(t *testing.T) {
	const input = `
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
	` 

	fn := build(input, t)
	t.Error(fn.String())

	chunk := fn.Chunk()
	t.Error("\n" + chunk.String())
}