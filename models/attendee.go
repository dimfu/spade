package models

import "database/sql"

type Attendee struct {
	Id           int
	TournamentID string
	PlayerID     string
	CurrentSeat  sql.NullInt64
}

type AttendeeModel struct {
	db *sql.DB
}
