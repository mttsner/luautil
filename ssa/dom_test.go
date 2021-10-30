package ssa

import (
	"fmt"
	"bytes"
	"testing"
)

func TestBuildDomTree(t *testing.T) {
	const input = `
	while true do
		if x < 10 then
			print("hi")
		else
			print("bye")
		end
	end

	if x < 10 then
		print("hi")
	end
	`
	fn := build(input, t)
	t.Error(fn.String())
	
	str := &bytes.Buffer{}
	str.WriteRune('\n')
	buildDomTree(fn)

	for b, d := range buildDomFrontier(fn) {
		str.WriteString(fmt.Sprintf("%d: ", b))
		for _, b2 := range d {
			str.WriteString(b2.String() + ", ")
		}
		str.WriteRune('\n')
	}

	printDomTreeDot(str, fn)

	t.Error(str.String())
}