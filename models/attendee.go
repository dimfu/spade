package models

import (
	"database/sql"
)

type Attendee struct {
	Id           int
	TournamentID string
	PlayerID     string
	CurrentSeat  sql.NullInt64
	Player       *Player
	Tournament   *Tournament
}

type AttendeeModel struct {
	DB *sql.DB
}

func NewAttendeeModel(db *sql.DB) *AttendeeModel {
	return &AttendeeModel{
		DB: db,
	}
}

func (m *AttendeeModel) GetAttendeeById(tournamentId, playerId string) (*Attendee, error) {
	a := &Attendee{}
	q := `SELECT id, tournament_id, player_id FROM attendees WHERE tournament_id = ? AND player_id = ?`
	err := m.DB.QueryRow(q, tournamentId, playerId).Scan(&a.Id, &a.TournamentID, &a.PlayerID)
	if err != nil {
		return nil, err
	}
	return a, nil
}
