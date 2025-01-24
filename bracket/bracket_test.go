package bracket

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/dimfu/spade/bracket/seeds"
	"github.com/dimfu/spade/bracket/templates"
)

type payload struct {
	name string
}

func TestGenerate(t *testing.T) {
	type testCase struct {
		size, winnerRoot int
		expected         func(bt *BracketTree, t *testing.T)
	}

	validateSeats := func(players, expectedSeats, got int) {
		if got != expectedSeats {
			t.Fatalf("expected to have %d from %d players but got %d", expectedSeats, players, got)
		}
	}

	validateWinnerNode := func(node *BracketTree, expected int) {
		if node.Root.Position != expected {
			t.Fatalf("expected winner root to be seat pos %d but got %d", expected, node.Root.Position)
		}
	}

	tests := []testCase{
		{
			size: 2,
			expected: func(bt *BracketTree, t *testing.T) {
				validateSeats(2, 3, bt.Size())
				validateWinnerNode(bt, 2)
			},
		},
		{
			size: 4,
			expected: func(bt *BracketTree, t *testing.T) {
				validateSeats(4, 7, bt.Size())
				validateWinnerNode(bt, 4)
			},
		},
		{
			size: 8,
			expected: func(bt *BracketTree, t *testing.T) {
				validateSeats(8, 15, bt.Size())
				validateWinnerNode(bt, 8)
			},
		},
		{
			size: 16,
			expected: func(bt *BracketTree, t *testing.T) {
				validateSeats(16, 31, bt.Size())
				validateWinnerNode(bt, 16)
			},
		},
		{
			size: 32,
			expected: func(bt *BracketTree, t *testing.T) {
				validateSeats(32, 63, bt.Size())
				validateWinnerNode(bt, 32)
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Generate tournament from %d players", tc.size), func(t *testing.T) {
			bt, err := GenerateFromTemplate(tc.size)
			if err != nil {
				t.Fatal(err.Error())
			}
			// bt.Print()
			tc.expected(bt, t)
		})
	}
}

func TestSeedingPosition(t *testing.T) {
	type testCase struct {
		size     int
		expected func(bt *BracketTree, t *testing.T)
	}

	validateCorrectSeat := func(expected, got []int) {
		if !reflect.DeepEqual(expected, got) {
			t.Fatalf("seat position is not correct as expected, got %v expected %v", got, expected)
		}
	}

	tests := []testCase{
		{
			size: templates.TOP_2,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: templates.TOP_4,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3, 5, 7}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: templates.TOP_8,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3, 5, 7, 9, 11, 13, 15}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: templates.TOP_16,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: templates.TOP_32,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{
					1, 3, 5, 7, 9, 11, 13,
					15, 17, 19, 21, 23, 25,
					27, 29, 31, 33, 35, 37,
					39, 41, 43, 45, 47, 49,
					51, 53, 55, 57, 59, 61, 63,
				}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Generate tournament from %d players", tc.size), func(t *testing.T) {
			bt, err := GenerateFromTemplate(tc.size)
			if err != nil {
				t.Fatal(err.Error())
			}
			tc.expected(bt, t)
		})
	}
}
func TestInsertDuplicatePos(t *testing.T) {
	type seed struct {
		payload payload
		seat    int
	}

	type testCase struct {
		nodePos        int
		expectedPlayer string
	}

	seeds := []seed{
		{payload: payload{name: "Player 1"}, seat: 6},
		{payload: payload{name: "Player 2"}, seat: 6},
		{payload: payload{name: "Player 3"}, seat: 7},
		{payload: payload{name: "Player 4"}, seat: 5},
	}

	tests := []testCase{
		{nodePos: 11, expectedPlayer: "Player 4"},
		{nodePos: 13, expectedPlayer: "Player 2"},
		{nodePos: 15, expectedPlayer: "Player 3"},
	}

	bt, err := GenerateFromTemplate(templates.TOP_8)
	if err != nil {
		t.Fatal(err.Error())
	}

	// should not seed beyond the set positions bound
	_, err = bt.Seed(999, payload{name: "Player Unknown"})
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}

	res := make(map[int]*Node)
	for _, seed := range seeds {
		node, err := bt.Seed(seed.seat, seed.payload)
		if err != nil {
			t.Fatal(err)
		}
		res[node.Position] = node
	}

	for _, tc := range tests {
		node, exists := res[tc.nodePos]
		if !exists {
			t.Fatalf("node at position %d does not exist", tc.nodePos)
		}

		payload, ok := node.Payload.(payload)
		if !ok {
			t.Fatalf("expected payload to be of type payload at position %d", tc.nodePos)
		}

		if payload.name != tc.expectedPlayer {
			t.Fatalf("expected player %q at position %d, but got %q", tc.expectedPlayer, tc.nodePos, payload.name)
		}
	}
}

func TestSimulateWinner(t *testing.T) {
	type payload struct {
		name string
	}
	type seed struct {
		payload payload
		pos     int
	}

	bt, err := GenerateFromTemplate(templates.TOP_8)
	if err != nil {
		t.Fatal(err.Error())
	}

	s := make([]interface{}, templates.TOP_8)
	for i := 0; i < templates.TOP_8; i++ {
		s[i] = payload{name: fmt.Sprintf("Player %d", i+1)}
	}

	newSeeds, err := seeds.NewSeeds(s, seeds.BEST_AGAINST_WORST, templates.TOP_8)
	if err != nil {
		t.Fatal(err)
	}

	for i, seed := range newSeeds {
		if _, err := bt.Seed(i, seed); err != nil {
			log.Fatal(err)
		}
	}

	node, err := bt.MatchWinner(3)
	if err != nil {
		t.Fatal(err)
	}
	node, err = bt.MatchWinner(node.Position)
	if err != nil {
		t.Fatal(err)
	}
	node, err = bt.MatchWinner(node.Position)
	if err != nil {
		t.Fatal(err)
	}

	winner, err := bt.Winner()
	if err != nil {
		t.Fatal(err)
	}

	if winner.Payload.(payload).name != node.Payload.(payload).name {
		t.Fatalf("the bracket winner should be %s", node.Payload.(payload).name)
	}
}
