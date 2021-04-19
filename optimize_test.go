package beautifier

import (
	"os"
	"strings"
	"testing"
	_ "embed"

	"github.com/notnoobmaster/beautifier/parse"
)

func TestOptimize(t *testing.T) {
	chunk, err := parse.Parse(strings.NewReader(test), "")
	if err != nil {
		t.Fatal(err)
	}
	
	Optimize(chunk)
	
	file, err := os.Create("tests/optimized.lua")

	file.WriteString(Beautify(chunk))
}