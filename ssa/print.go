package ssa

import (
	"fmt"
	"strings"
)


func (v *If) String() string {
	return fmt.Sprintf("if %s then %s else %s", v.Cond, v.True, v.False)
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
	return fmt.Sprintf("for %s = %s, %s, %s do %s end", v.Local, v.Init, v.Limit, v.Step, v.Body)
}

func (v *GenericFor) String() string {
	var b strings.Buffer
	b.WriteString("for ")
	b.WriteString(fmt.Sprintf("%s", v.Locals[0]))
	for i := 1; i < len(v.Locals); i++ {
		b.WriteString(fmt.Sprintf("%s,", v.Locals[0]))
	}
	b.WriteString(" in ")
	return b.String()
}