package bracket

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/dimfu/spade/bracket/templates"
)

type BracketTree struct {
	InsertionOrder []int
	Root           *Node
	Matches        []templates.Match
	StartingSeats  []int
	SeatRoundPos   map[int][]int
}

func NewBracketTree(node *Node, matches []templates.Match) *BracketTree {
	return &BracketTree{
		InsertionOrder: []int{},
		Root:           node,
		Matches:        matches,
		StartingSeats:  []int{},
		SeatRoundPos:   make(map[int][]int),
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
	maxDepth := int(math.Floor(math.Log2(float64(tNodes)))) + 1

	seats := make([]int, tNodes)
	for i := 0; i < tNodes; i++ {
		seats[i] = i + 1
	}

	var depth float64
	var visited int

	mid := len(seats) / 2

	root := &Node{Position: seats[mid]}
	queue := []struct {
		Node  *Node
		Start int
		End   int
	}{{Node: root, Start: 0, End: len(seats) - 1}}

	bt := NewBracketTree(root, matches)
	bt.InsertionOrder = append(bt.InsertionOrder, seats[mid])

	for len(queue) > 0 {
		visited++
		curr := queue[0]
		queue = queue[1:]

		depth = math.Ceil(math.Log(float64(visited+1)) / math.Log(2))
		relativeDepth := maxDepth - int(depth) + 1
		// Store nodes at relative depth
		bt.SeatRoundPos[relativeDepth] = append(bt.SeatRoundPos[relativeDepth], curr.Node.Position)

		// find the middle of the entire current segment (or parent)
		mid := (curr.Start + curr.End) / 2 // 7

		// construct left subtree
		if curr.Start <= mid-1 {
			// find the middle of left sub section before the parent
			leftMid := (curr.Start + mid - 1) / 2
			leftNode := &Node{Position: seats[leftMid]}
			curr.Node.Left = leftNode
			bt.InsertionOrder = append(bt.InsertionOrder, seats[leftMid])
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
			bt.InsertionOrder = append(bt.InsertionOrder, seats[rightMid])
			queue = append(queue, struct {
				Node  *Node
				Start int
				End   int
			}{Node: rightNode, Start: mid + 1, End: curr.End})
		}
	}

	// get the seeding position seats or first round seats
	bt.StartingSeats = bt.SeatRoundPos[1]
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
	if pos < 0 || pos > len(bt.StartingSeats) {
		return -1, fmt.Errorf("position %d is out of bounds", pos)
	}
	return bt.StartingSeats[pos], nil
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

	for curPos := len(bt.StartingSeats) - 1; curPos >= pos; curPos-- {
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

func (bt *BracketTree) NodesInRound(round int) ([]*Node, error) {
	var nodes []*Node
	if seats, exists := bt.SeatRoundPos[round]; exists {
		for _, seat := range seats {
			node, err := bt.Search(seat)
			if err != nil {
				log.Fatal(err)
			}
			nodes = append(nodes, node)
		}
		return nodes, nil
	}
	return nil, errors.New("can't find any nodes for this round")
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
