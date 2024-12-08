package templates

var Top8 Matches = Matches{
	// Quarters
	{Seats: [2]int{1, 3}, WinnerTo: 2},
	{Seats: [2]int{5, 7}, WinnerTo: 6},
	{Seats: [2]int{9, 11}, WinnerTo: 10},
	{Seats: [2]int{13, 15}, WinnerTo: 14},

	// Semifinal
	{Seats: [2]int{2, 6}, WinnerTo: 4},
	{Seats: [2]int{10, 14}, WinnerTo: 12},

	// Final
	{Seats: [2]int{4, 12}, WinnerTo: 8},
}
