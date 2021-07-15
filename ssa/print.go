package ssa

import (
	"fmt"
	"strings"
)


func (s *If) String() string {
	// Be robust against malformed CFG.
	tblock, fblock := -1, -1
	if s.block != nil && len(s.block.Succs) == 2 {
		tblock = s.block.Succs[0].Index
		fblock = s.block.Succs[1].Index
	}
	return fmt.Sprintf("if %s goto %s else %d", s.Cond, tblock, fblock)
}

func (v *Assign) String() string {
	return fmt.Sprintf("%s = %s", v.Lhs, v.Rhs)
}

func (v *CompoundAssign) String() string {
	return fmt.Sprintf("%s %s= %s", v.Lhs, v.Op, v.Rhs)
}

func (v *While) String() string {
	return fmt.Sprintf("while %s do %s end", v.Cond, v.Body)
}

func (v *NumberFor) String() string {
	return fmt.Sprintf("for %s = %s, %s, %s do %s end", v.Local, v.Init, v.Limit, v.Step)
}

func (v *GenericFor) String() string {
	var b strings.Builder
	b.WriteString("for ")
	b.WriteString(fmt.Sprintf("%s", v.Locals[0]))
	for i := 1; i < len(v.Locals); i++ {
		b.WriteString(fmt.Sprintf("%s,", v.Locals[0]))
	}
	b.WriteString(" in ")
	return b.String()
}