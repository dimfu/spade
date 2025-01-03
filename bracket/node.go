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

func (n *Node) search(pos int) *Node {
	curr := n
	for curr != nil {
		if curr.Position == pos {
			return curr
		} else if curr.Position > pos {
			curr = curr.Left
		} else {
			curr = curr.Right
		}
	}
	return curr
}

func (n *Node) Print(w io.Writer, ns int, ch rune) {
	if n == nil {
		return
	}

	for i := 0; i < ns; i++ {
		fmt.Fprint(w, " ")
	}
	fmt.Fprintf(w, "%c:%v %v\n", ch, n.Position, n.Payload)
	n.Left.Print(w, ns+2, 'L')
	n.Right.Print(w, ns+2, 'R')
}
