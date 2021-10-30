package ast

import (
	"fmt"
)

type Position struct {
	Source string
	Line   int
	Column int
}

type Token struct {
	Type int
	Name string
	Str  string
	Num  float64
	Pos  Position
}

func (t *Token) String() string {
	return fmt.Sprintf("<type:%v, str:%v>", t.Name, t.Str)
}
