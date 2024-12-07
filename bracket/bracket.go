package bracket

import (
	"math"
	"os"
)

type BracketTree struct {
	InsertionOrder []int
	Root           *Node
}

func NewBracketTree(node *Node) *BracketTree {
	return &BracketTree{
		InsertionOrder: []int{},
		Root:           node,
	}
}

func calculateRounds(players int) int {
	rounds := 0
	for players > 1 {
		rounds++
		players = (players + 1) / 2
	}
	return rounds
}

func GenerateFromTemplate(size int) *BracketTree {
	if size == 0 {
		return nil
	}

	// generate how many rounds/depths it takes from N players
	rounds := calculateRounds(size)
	tNodes := int(math.Pow(2, float64(rounds+1))) - 1 // and we find out how many nodes needed to have tree with N rounds

	seats := make([]int, tNodes)
	for i := 0; i < tNodes; i++ {
		seats[i] = i + 1
	}

	mid := len(seats) / 2
	insertionOrder := []int{}
	insertionOrder = append(insertionOrder, seats[mid])

	root := &Node{Position: seats[mid]}
	queue := []struct {
		Node  *Node
		Start int
		End   int
	}{{Node: root, Start: 0, End: len(seats) - 1}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// find the middle of the entire current segment (or parent)
		mid := (curr.Start + curr.End) / 2 // 7

		// construct left subtree
		if curr.Start <= mid-1 {
			// find the middle of left sub section before the parent
			leftMid := (curr.Start + mid - 1) / 2
			leftNode := &Node{Position: seats[leftMid]}
			curr.Node.Left = leftNode
			insertionOrder = append(insertionOrder, seats[leftMid])
			queue = append(queue, struct {
				Node  *Node
				Start int
				End   int
			}{Node: leftNode, Start: curr.Start, End: mid - 1})
		}

		// construct right subtree
		if mid+1 <= curr.End {
			// find the middle of right sub section after the parent
			rightMid := (mid + 1 + curr.End) / 2
			rightNode := &Node{Position: seats[rightMid]}
			curr.Node.Right = rightNode
			insertionOrder = append(insertionOrder, seats[rightMid])
			queue = append(queue, struct {
				Node  *Node
				Start int
				End   int
			}{Node: rightNode, Start: mid + 1, End: curr.End})
		}
	}

	bt := NewBracketTree(root)
	bt.InsertionOrder = insertionOrder
	return bt
}

// visualization for debugging purposes
func (bt *BracketTree) Print() {
	if bt.Root == nil {
		return
	} else {
		bt.Root.Print(os.Stdout, 0, 'M')
	}
}

func (bt BracketTree) Size() int {
	return len(bt.InsertionOrder)
}
