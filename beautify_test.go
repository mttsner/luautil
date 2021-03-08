package beautifier

import (
	"os"
	"strings"
	"testing"
	_ "embed"

	"github.com/yuin/gopher-lua/parse"
)

func TestBeautifiy(t *testing.T) {
	chunk, err := parse.Parse(strings.NewReader(test), "")
	if err != nil {
		t.Fatal(err)
	}
	
	file, err := os.Create("beautified.lua")

	file.WriteString(Beautify(chunk))
}