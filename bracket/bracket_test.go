package bracket

import (
	"fmt"
	"testing"
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
			bt := GenerateFromTemplate(tc.size)
			fmt.Println(bt.InsertionOrder)
			tc.expected(bt, t)
		})
	}
}
