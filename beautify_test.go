package luautil

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed tests/test.lua
var test string

func TestBeautify(t *testing.T) {
	t.Error(Beautify(strings.NewReader(test)))
}
