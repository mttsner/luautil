package beautifier

import (
	"os"
	"strings"
	"testing"
	_ "embed"

	"github.com/notnoobmaster/beautifier/parse"
)

//go:embed tests/test.lua
var test string

func TestBeautifiy(t *testing.T) {
	chunk, err := parse.Parse(strings.NewReader(test), "")
	if err != nil {
		t.Fatal(err)
	}
	
	file, err := os.Create("tests/beautified.lua")

	file.WriteString(Beautify(chunk))
}