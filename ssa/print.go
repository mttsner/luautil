package ssa

import (
	"fmt"
	"strconv"
	"strings"
)

func (s Nil) String() string { return "nil" }
func (s True) String() string { return "true" }
func (s False) String() string { return "false" }
func (s VarArg) String() string { return "..." }

func (s Number) String() string {
	return strconv.FormatFloat(s.Value, 'f', -1, 64)
}

func (s String) String() string {
	return strconv.Quote(s.Value)
}

func (s Table) String() string {
	b := &strings.Builder{}
	b.WriteRune('{')
	for i, field := range s.Fields {
		if i != 0 {
			b.WriteString(", ")
		}
		if field.Key != nil {
			fmt.Fprintf(b, "[%s] = ", field.Key.String())
		}
		b.WriteString(field.Value.String())
	}
	b.WriteRune('}')
	return b.String()
}

func (s AttrGet) String() string {
	return fmt.Sprintf("%s[%s]", s.Object, s.Key)
}

func (s Arithmetic) String() string {
	return fmt.Sprintf("%s %s %s", s.Lhs, s.Op, s.Rhs)
}

func (s Unary) String() string {
	return fmt.Sprintf("%s%s", s.Op, s.Value)
}

func (s Concat) String() string {
	return fmt.Sprintf("%s .. %s", s.Lhs, s.Rhs)
}

func (s Relation) String() string {
	return fmt.Sprintf("%s %s %s", s.Lhs, s.Op, s.Rhs)
}

func (s Logic) String() string {
	return fmt.Sprintf("%s %s %s", s.Lhs, s.Op, s.Rhs)
}

func (s *Local) String() string {
	return s.Name()
}

func (v *Global) String() string { 
	return v.Comment
}


func (s Call) String() string {
	b := &strings.Builder{}

	if s.Func != nil { // func()
		b.WriteString(s.Func.String())
	} else { // hoge:method()
		b.WriteString(s.Recv.String())
		b.WriteRune(':')
		b.WriteString(s.Method)
	}

	b.WriteRune('(')
	for i, arg := range s.Args {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.String())
	}
	b.WriteRune(')')

	return b.String()
}

func (s *Return) String() string { return "" }

func (f *Function) String() string {
	b := &strings.Builder{}
	WriteFunction(b, f)
	return b.String()
}

func (s *Jump) String() string {
	// Be robust against malformed CFG.
	block := -1
	if s.block != nil && len(s.block.Succs) == 1 {
		block = s.block.Succs[0].Index
	}
	return fmt.Sprintf("jump %d", block)
}

func (s *If) String() string {
	// Be robust against malformed CFG.
	tblock, fblock := -1, -1
	if s.block != nil && len(s.block.Succs) == 2 {
		tblock = s.block.Succs[0].Index
		fblock = s.block.Succs[1].Index
	}
	return fmt.Sprintf("if %s goto %d else %d", s.Cond, tblock, fblock)
}

func (v *Assign) String() string {
	return fmt.Sprintf("%s = %s", v.Lhs, v.Rhs)
}

func (v *CompoundAssign) String() string {
	return fmt.Sprintf("%s %s= %s", v.Lhs, v.Op, v.Rhs)
}

func (v *NumberFor) String() string {
	return fmt.Sprintf("for %s = %s, %s, %d do", v.Local, v.Init, v.Limit, v.Step)
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

func (v *Phi) String() string {
	var b strings.Builder
	/*b.WriteString("phi [")
	for i, edge := range v.Edges {
		if i > 0 {
			b.WriteString(", ")
		}
		// Be robust against malformed CFG.
		if v.block == nil {
			b.WriteString("??")
			continue
		}
		block := -1
		if i < len(v.block.Preds) {
			block = v.block.Preds[i].Index
		}
		fmt.Fprintf(&b, "%d: ", block)
		edgeVal := "<nil>" // be robust
		//if edge != nil {
			//edgeVal = relName(edge, v)
		//}
		b.WriteString(edgeVal)
	}
	b.WriteString("]")
	if v.Comment != "" {
		b.WriteString(" #")
		b.WriteString(v.Comment)
	}*/
	return b.String()
}

// String returns a human-readable label of this block.
// It is not guaranteed unique within the function.
//
func (b *BasicBlock) String() string {
	return fmt.Sprintf("%d", b.Index)
}
