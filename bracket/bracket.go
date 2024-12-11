package bracket

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/dimfu/spade/bracket/templates"
)

type BracketTree struct {
	InsertionOrder []int
	Root           *Node
	Matches        []templates.Match
	StartingSeats  []int
}

func NewBracketTree(node *Node, matches []templates.Match) *BracketTree {
	return &BracketTree{
		InsertionOrder: []int{},
		Root:           node,
		Matches:        matches,
		StartingSeats:  []int{},
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

func GenerateFromTemplate(size int) (*BracketTree, error) {
	if size <= 0 {
		return nil, errors.New("size cannot be <= 0")
	}

	matches, err := templates.WithTemplate(size)
	if err != nil {
		return nil, err
	}

	// generate how many rounds/depths it takes from N players
	rounds := calculateRounds(size)
	// and we find out how many nodes needed to have tree with N rounds
	tNodes := int(math.Pow(2, float64(rounds+1))) - 1

	seats := make([]int, tNodes)
	for i := 0; i < tNodes; i++ {
		seats[i] = i + 1
	}

	var depth float64
	var visited int
	seedingPos := make(map[int][]int)

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
		visited++
		curr := queue[0]
		queue = queue[1:]

		depth = math.Ceil(math.Log(float64(visited+1)) / math.Log(2))
		seedingPos[int(depth)] = append(seedingPos[int(depth)], curr.Node.Position)

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

	bt := NewBracketTree(root, matches)
	bt.InsertionOrder = insertionOrder
	bt.StartingSeats = seedingPos[rounds+1] // rounds + 1 because we dont count the root
	return bt, nil
}

func (bt *BracketTree) Search(pos int) (*Node, error) {
	if bt.Root == nil {
		return nil, errors.New("root is empty")
	}
	node := bt.Root.search(pos)
	if node == nil {
		return nil, errors.New("node not found")
	}

	return node, nil
}

func (bt BracketTree) findSeedPos(pos int) (int, error) {
	if pos < 1 || pos > len(bt.StartingSeats) {
		return -1, fmt.Errorf("position %d is out of bounds", pos)
	}
	return bt.StartingSeats[pos-1], nil
}

func (bt *BracketTree) Seed(pos int, payload interface{}) (*Node, error) {
	idx, err := bt.findSeedPos(pos)
	if err != nil {
		return nil, err
	}
	node, err := bt.Search(idx)
	if err != nil {
		return nil, err
	}

	if node.Payload == nil {
		node.Payload = payload
		return node, nil
	}

	for curPos := len(bt.StartingSeats); curPos >= pos; curPos-- {
		curIdx, err := bt.findSeedPos(curPos)
		if err != nil {
			return nil, err
		}
		curNode, err := bt.Search(curIdx)
		if err != nil {
			return nil, err
		}

		if curPos == pos {
			curNode.Payload = payload
			return curNode, nil
		}

		prevNode, err := bt.Search(curIdx - 2)
		if err != nil {
			return nil, err
		}

		curNode.Payload = prevNode.Payload
	}

	return node, nil
}

func (bt *BracketTree) Winner() (*Node, error) {
	if bt.Root.Payload != nil {
		return bt.Root, nil
	} else {
		return nil, errors.New("bracket winner has not yet determined")
	}
}

func (bt *BracketTree) MatchWinner(seat int) (*Node, error) {
	node, err := bt.Search(seat)
	if err != nil {
		return nil, err
	}

	for _, match := range bt.Matches {
		for _, s := range match.Seats {
			if s == seat {
				toNode, err := bt.Search(match.WinnerTo)
				if err != nil {
					return nil, err
				}
				toNode.Payload = node.Payload
				return toNode, nil
			}
		}
	}

	return nil, errors.New("seat not found")
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
