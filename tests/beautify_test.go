package tests

import (
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/notnoobmaster/luautil/parse"
)

var assignment = []string{
	"_ = _;\n",
	"_, _ = _, _;\n",
}

var control = []string{
	"while _ do\nend;\n",
	"repeat\nuntil _;\n",
	"for _ = _, _ do\nend;\n",
	"for _ = _, _, _ do\nend;\n",
	"for _ in _ do\nend;\n",
	"for _, _ in _, _ do\nend;\n",
	"if _ then\nend;\n",
	"if _ then\nelse\n\t_ = _;\nend;\n",
	"if _ then\nelseif _ then\n\t_ = _;\nelse\n\t_ = _;\nend;\n",
}

var terminators = []string{
	"while _ do\n\tbreak;\nend;\n",
	"repeat\n\tbreak;\nuntil _;\n",
	"for _ = _, _ do\n\tbreak;\nend;\n",
	"for _ = _, _, _ do\n\tbreak;\nend;\n",
	"for _ in _ do\n\tbreak;\nend;\n",
	"for _, _ in _, _ do\n\tbreak;\nend;\n",
	"return;\n",
	"return _;\n",
	"return _, _;\n",
}

var declarations = []string{
	"local _;\n",
	"local _ = _;\n",
	"local _, _ = _, _;\n",
}

var expressions = []string{
	"_ = \"\";\n",
	"_ = 0;\n",
	"_ = true;\n",
	"_ = false;\n",
	"_ = nil;\n",
}

var arithmetic = []string{
	"_ = _ + _;\n",
	"_ = _ - _;\n",
	"_ = _ * _;\n",
	"_ = _ / _;\n",
	"_ = _ % _;\n",
	"_ = _ ^ _;\n",
}

var relational = []string{
	"_ = _ == _;\n",
	"_ = _ ~= _;\n",
	"_ = _ < _;\n",
	"_ = _ <= _;\n",
	"_ = _ > _;\n",
	"_ = _ >= _;\n",
}

var logical = []string{
	"_ = _ and _;\n",
	"_ = _ or _;\n",
	"_ = not _;\n",
}

var concatenation = []string{
	"_ = _ .. _;\n",
}

var length = []string{
	"_ = #_;\n",
}

var table = []string{
	"_ = {};\n",
	"_ = {\n\t_\n};\n",
	"_ = {\n\t_,\n\t_\n};\n",
	"_ = {\n\t_ = {}\n};\n",
	"_ = {\n\t[_] = {}\n};\n",
}

var calls = []string{
	"_();\n",
	"_(_);\n",
	"_(_, _);\n",
	"_:_();\n",

	"_ = _();\n",
	"_ = _(_);\n",
	"_ = _(_, _);\n",
	"_ = _:_();\n",
}

var definitions = []string{
	"function _()\nend;\n",
	"function _(_)\nend;\n",
	"function _(...)\nend;\n",
	"function _(_, _)\nend;\n",
	"function _(_, ...)\nend;\n",

	"function _._()\nend;\n",
	"function _._(_)\nend;\n",
	"function _._(...)\nend;\n",
	"function _._(_, _)\nend;\n",
	"function _._(_, ...)\nend;\n",

	"function _:_()\nend;\n",
	"function _:_(_)\nend;\n",
	"function _:_(...)\nend;\n",
	"function _:_(_, _)\nend;\n",
	"function _:_(_, ...)\nend;\n",

	"function _._:_()\nend;\n",
	"function _._:_(_)\nend;\n",
	"function _._:_(...)\nend;\n",
	"function _._:_(_, _)\nend;\n",
	"function _._:_(_, ...)\nend;\n",

	"local function _()\nend;\n",
	"local function _(_)\nend;\n",
	"local function _(...)\nend;\n",
	"local function _(_, _)\nend;\n",
	"local function _(_, ...)\nend;\n",
}

var bitwise = []string{
	"_ = _ & _;\n",
	"_ = _ | _;\n",
	"_ = _ ~ _;\n",
	"_ = _ << _;\n",
	"_ = _ >> _;\n",
	"_ = ~_;\n",
}

var tests = map[string][]string{
	"Assignment": assignment,
	"Structures":    control,
	"Terminators": terminators,
	"Declarations": declarations,
	"Expressions": expressions,
	"Arithmetic": arithmetic,
	"Relational": relational,
	"Logical":    logical,
	"Concatenation": concatenation,
	"Length":      length,
	"Table":       table,
	"Calls":       calls,
	"Definitions": definitions,
	"Bitwise": bitwise,
}

func TestFormat(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T){
			for _, s := range test {
				chunk, err := parse.Parse(strings.NewReader(s), "")
				if err != nil {
					t.Fatal(err)
				}
				if chunk.String() != s {
					t.Fatalf("\nGot:\n%sExpected:\n%s", chunk, s)
				}
			}
		})
	}
}

func TestNumbers(t *testing.T) {
	numbers := []string{
		"_ = 0;\n",
		"_ = 0.0;\n",

		"_ = 0.0e0;\n",
		"_ = 0.0e+0;\n",
		"_ = 0.0e-0;\n",

		"_ = 0x0;\n",
		"_ = 0X0;\n",
		"_ = 0x0_0__0;\n",
		
		"_ = 0b0;\n",
		"_ = 0B0;\n",
		"_ = 0b0_0__0;\n",

		"_ = 0o0;\n",
		"_ = 0O0;\n",
		"_ = 0o0_0__0;\n",
	}

	for _, s := range numbers {
		chunk, err := parse.Parse(strings.NewReader(s), "")
		if err != nil {
			t.Fatal(err)
		}
		if chunk.String() != "_ = 0;\n" {
			t.Fatalf("\nGot:\n%sExpected:\n_ = 0;\n", chunk)
		}
	}
}

/*
_ = _ or _ and _

*/

//go:embed test.lua
var test string

func TestBeautify(t *testing.T) {
	f, err := os.Create("test.lua")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	chunk, err := parse.Parse(strings.NewReader(test), "")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(chunk.String())
}