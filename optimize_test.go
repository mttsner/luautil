package beautifier

import (
	"os"
	"testing"

	"github.com/yuin/gopher-lua/parse"
)

func TestOptimize(t *testing.T) {
	file, err := os.Open("obfuscated.lua")
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := parse.Parse(file, "")
	if err != nil {
		t.Fatal(err)
	}
	
	Optimize(chunk)
	
	newfile, err := os.Create("beautified.lua")

	newfile.WriteString(Beautify(&chunk))
}