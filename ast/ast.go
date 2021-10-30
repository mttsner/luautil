package ast

type PositionHolder interface {
	Line() int
	SetLine(int)
	LastLine() int
	SetLastLine(int)
}

type Node struct {
	line     int
	lastline int
}

func (n *Node) Line() int {
	return n.line
}

func (n *Node) SetLine(line int) {
	n.line = line
}

func (n *Node) LastLine() int {
	return n.lastline
}

func (n *Node) SetLastLine(line int) {
	n.lastline = line
}