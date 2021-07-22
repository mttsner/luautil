package optimize 

import (
	"github.com/notnoobmaster/luautil/ssa"
)

type emulator struct {}

func toBool(v ssa.Value) bool {
	switch v := v.(type) {
	case ssa.Const:
		switch v := v.Value.(type) {
		case bool:
			return v
		case string:
			return true
		case float64:
			return true
		case nil:
			return false
		default:
			panic("unsupported constant type")
		}
	}
	return false
}

func (e *emulator) instruction(inst ssa.Instruction) {
	switch i := inst.(type) {
	case *ssa.If:
		if _, ok := i.Cond.(ssa.Unknown); ok {
			break
		}
		if Bool(i.Cond) {
			e.block(i.Block().Succs[0])
		} else {
			e.block(i.Block().Succs[1])
		}
	}
}

func (e *emulator) blocks(bl []*ssa.BasicBlock) {
	for _, b := range bl {
		e.block(b)
	}
}

func (e *emulator) block(b *ssa.BasicBlock) {
	for _, inst := range b.Instrs {
		e.instruction(inst)
	}
}

func Optimize(f *ssa.Function) {
	e := &emulator{}
	e.blocks(f.Blocks)
}