package luautil

import (
	_ "embed"
	"os"
	"strings"
	"testing"
)

//go:embed tests/test.lua
var test string

func TestBeautify(t *testing.T) {
	f, err := os.Create("test.lua")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	str, err := Beautify(strings.NewReader(test))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(str)
}
