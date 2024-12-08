package templates

var Top16 Matches = Matches{
	{Seats: [2]int{1, 3}, WinnerTo: 2},
	{Seats: [2]int{5, 7}, WinnerTo: 6},
	{Seats: [2]int{9, 11}, WinnerTo: 10},
	{Seats: [2]int{13, 15}, WinnerTo: 14},
	{Seats: [2]int{17, 19}, WinnerTo: 18},
	{Seats: [2]int{21, 23}, WinnerTo: 22},
	{Seats: [2]int{25, 27}, WinnerTo: 26},
	{Seats: [2]int{29, 31}, WinnerTo: 30},

	// Quarterfinals
	{Seats: [2]int{2, 6}, WinnerTo: 4},
	{Seats: [2]int{10, 14}, WinnerTo: 12},
	{Seats: [2]int{18, 22}, WinnerTo: 20},
	{Seats: [2]int{26, 30}, WinnerTo: 28},

	// Semifinals
	{Seats: [2]int{4, 12}, WinnerTo: 8},
	{Seats: [2]int{20, 28}, WinnerTo: 24},

	// Finals
	{Seats: [2]int{8, 24}, WinnerTo: 16},
}
