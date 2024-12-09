package bracket

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/dimfu/spade/bracket/templates"
)

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

func TestInsertSeed(t *testing.T) {
	type payload struct {
		name string
	}

	type seed struct {
		payload payload
		pos     int
	}

	type testCase struct {
		nodePos        int
		expectedPlayer string
	}

	seeds := []seed{
		{payload: payload{name: "Player 1"}, pos: 7},
		{payload: payload{name: "Player 2"}, pos: 7},
		{payload: payload{name: "Player 3"}, pos: 8},
		{payload: payload{name: "Player 4"}, pos: 6},
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
		node, err := bt.Seed(seed.pos, seed.payload)
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
