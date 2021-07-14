package luautil 

import (
	"testing"
	"strings"

	"github.com/notnoobmaster/luautil/parse"
)

var optimizeTest = `
	local _ = 1+2
	local _ = 1-2
	local _ = 1*2
	local _ = 1/2
	local _ = 1%2
	local _ = 1^2
	local _ = "a".."z"
	local _ = true and false
	local _ = true or false
`

var optimizeTarget = `local _ = 3;
local _ = -1;
local _ = 2;
local _ = 0.5;
local _ = 1;
local _ = 1;
local _ = "az";
local _ = false;
local _ = true;
`



func TestOptimize(t *testing.T) {
	chunk, err := parse.Parse(strings.NewReader(optimizeTest), "")
	if err != nil {
		t.Fatal(err)
	}

	Optimize(chunk)

	if chunk.String() != optimizeTarget {
		t.Error("Test failed")
	}
}