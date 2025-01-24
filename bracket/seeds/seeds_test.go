package seeds

import (
	"reflect"
	"testing"
)

type participants struct {
	name string
}

func genParticipants(size int) []interface{} {
	ps := []interface{}{}
	for i := 0; i < size; i++ {
		ps = append(ps, i+1)
	}
	return ps
}

func TestRandomSeed(t *testing.T) {
	participants := genParticipants(8)

	prev := make([]interface{}, len(participants))
	copy(prev, participants)

	seeded, err := NewSeeds(participants, RANDOM, 8)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(prev, seeded) {
		t.Fatal("should be randomized")
	}
}

func TestBAWSeed(t *testing.T) {
	type testCase struct {
		participants []interface{}
		expected     []int
	}

	tests := []testCase{
		{participants: genParticipants(8), expected: []int{1, 8, 3, 6, 5, 4, 7, 2}},
		{participants: genParticipants(16), expected: []int{1, 16, 3, 14, 5, 12, 7, 10, 9, 8, 11, 6, 13, 4, 15, 2}},
		{participants: genParticipants(12), expected: []int{1, -1, 3, -1, 5, 12, 7, 10, 9, 8, 11, 6, -1, 4, -1, 2}},
		{participants: genParticipants(32), expected: []int{1, 32, 3, 30, 5, 28, 7, 26, 9, 24, 11, 22, 13, 20, 15, 18, 17, 16, 19, 14, 21, 12, 23, 10, 25, 8, 27, 6, 29, 4, 31, 2}},
		{participants: genParticipants(7), expected: []int{1, -1, 3, 6, 5, 4, 7, 2}},
	}

	for _, tc := range tests {
		prev := make([]interface{}, len(tc.participants))
		copy(prev, tc.participants)
		seeded, err := NewSeeds(tc.participants, BEST_AGAINST_WORST, len(tc.expected))
		if err != nil {
			t.Fatal(err)
		}

		if reflect.DeepEqual(prev, seeded) {
			t.Fatalf("array should be %v but got %v", tc.expected, prev)
		}

		seededInts := make([]int, len(seeded))
		for i, v := range seeded {
			if v == nil {
				seededInts[i] = -1
				continue
			}
			seededInts[i] = v.(int)
		}

		if !reflect.DeepEqual(tc.expected, seededInts) {
			t.Fatalf("array should be %v but got %v", tc.expected, seededInts)
		}
	}
}
