package bracket

import (
	"fmt"
	"reflect"
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
			size: 8,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3, 5, 7, 9, 11, 13, 15}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: 16,
			expected: func(bt *BracketTree, t *testing.T) {
				expected := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}
				validateCorrectSeat(expected, bt.StartingSeats)
			},
		},
		{
			size: 32,
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
