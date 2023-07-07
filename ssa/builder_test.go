package ssa

import (
	"strings"
	"testing"

	"github.com/notnoobmaster/luautil/parse"
)

// Helper function
func build(str string, t *testing.T) *Function {
	chunk, err := parse.Parse(strings.NewReader(str), "")
	if err != nil {
		t.Fatal(err)
	}
	return Build(chunk)
}

func TestClosure(t *testing.T) {
	const input = `
	local t0 = 0
	local t1 = 1
	local t2 = 2
	for t3, t4 in t0, t0, t0 do
		t0 = t3
		break
	end
	t0 = t2
	`
	fn := build(input, t)
	b := &strings.Builder{}
	WriteCfgDot(b, fn)
	WriteFunction(b, fn)
	t.Error("\n" + b.String())
}

func TestExpr(t *testing.T) {
	const input = `
	-- Expressions
	local t1, t2, t3, t3, t5 = 1, nil, false, true, "str"	-- Const
	local t6  = ... 										-- VarArg
	local t7  = _G.print 									-- AttrGet
	local t8  = {1,2,3,"str",[5]=nil,hello=""}				-- Table
	local t9  = 1+1-1*1/1^1									-- Arithmetic 
	local t10 = ""..""										-- String concat
	local t11 = 1 < 2										-- Relational
	local t12 = true and false or true 						-- Logical
	local t13 = not true									-- Unary
	local t13 = print() 									-- Call
	local _   = function() end								-- Function
	`
	fn := build(input, t)
	t.Error(fn.String())
}
func TestBuild(t *testing.T) {
	const input = `
	-- Statements
	local a = 1 -- Local assign
	b = a -- Assign
	b += 1 -- Compound assign
	do local b = 3 end -- Do block
	print(b) -- Func call 
	while false do end -- while loop
	repeat until false end -- repeat loop
	function c() -- Function definition
		return 1 -- return statement
	end
	if true then -- If statement
		local b = 1 
	else 
		local b = 2 
	end

	for i,v in next, {}, nil do -- Generic for loop
		break -- Break statement
	end

	for i=1,2,3 do -- Numerical for loop
		continue -- Continue statement
	end

	
	`
	fn := build(input, t)
	t.Error(fn.String())
}

func TestStuff(t *testing.T) {
	const input = `
	local t0 = 0

	if global then
		t0 = 1
	end

	print(t0+1)
	`
	fn := build(input, t)
	buildReferrers(fn)
	buildDomTree(fn)
	BuildDomFrontier(fn)

	lift(fn) 
	b := &strings.Builder{}
	WriteFunction(b, fn)
	WriteCfgDot(b, fn)
	t.Error("\n" + b.String())
}