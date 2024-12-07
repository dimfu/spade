package bracket

import (
	"fmt"
	"io"
)

type Node struct {
	Left     *Node
	Right    *Node
	Position int
	Payload  interface{}
}

func NewNode(position int, payload interface{}) *Node {
	return &Node{
		Position: position,
		Payload:  payload,
		Left:     nil,
		Right:    nil,
	}
}

func (n *Node) Print(w io.Writer, ns int, ch rune) {
	if n == nil {
		return
	}

	for i := 0; i < ns; i++ {
		fmt.Fprint(w, " ")
	}
	fmt.Fprintf(w, "%c:%v\n", ch, n.Position)
	n.Left.Print(w, ns+2, 'L')
	n.Right.Print(w, ns+2, 'R')
}
