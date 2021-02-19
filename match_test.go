package beautifier

import (
	"strings"
	"testing"

	//"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

const target = `
local function Wrap(Chunk, _IdentExpr_, Env)
	local Instr = Chunk[1];
	local Proto = Chunk[2];
	local Params = Chunk[_NumberExpr_];

	return function(...)

	end
end
`

const pattern = `
local function Wrap(Chunk, _IdentExpr_, Env)
	local Instr = Chunk[1];
	local Proto = Chunk[2];
	local Params = Chunk[_NumberExpr_];

	return function(...)

	end
end
`

func TestMatch(t *testing.T) {
	target, err := parse.Parse(strings.NewReader(target), "")
	if err != nil {
		t.Fatal(err)
	}

	pattern, err := parse.Parse(strings.NewReader(pattern), "")
	if err != nil {
		t.Fatal(err)
	}

	success, exprs := Match(target, pattern)
	t.Log(exprs)
	if !success {
		t.Error(success)
	}
}
