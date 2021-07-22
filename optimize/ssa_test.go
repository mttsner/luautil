package optimize

import (
	"testing"
	"strings"
	"github.com/notnoobmaster/luautil/ssa"
	"github.com/notnoobmaster/luautil/parse"
)

func TestOptimize(t *testing.T) {
	const input = `
	if true then
		print("Hello World")
	end
	`
	chunk, err := parse.Parse(strings.NewReader(input), "")
	if err != nil {
		t.Fatal(err)
	}
	fn := ssa.Build(chunk)
	Optimize(fn)
	b := &strings.Builder{}
	ssa.WriteFunction(b, fn)
	t.Error(b.String())
}