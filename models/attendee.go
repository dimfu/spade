package models

import (
	"database/sql"
	"log"
)

type Attendee struct {
	Id           int
	TournamentID string
	PlayerID     string
	CurrentSeat  sql.NullInt64
	Player       Player
	Tournament   Tournament
}

type AttendeeWithResult struct {
	Attendee
	Result int
}

type AttendeeModel struct {
	DB *sql.DB
}

func NewAttendeeModel(db *sql.DB) *AttendeeModel {
	return &AttendeeModel{
		DB: db,
	}
}

func (m *AttendeeModel) FindById(tournamentId, playerId string) (*Attendee, error) {
	a := &Attendee{}
	q := `SELECT id, tournament_id, player_id FROM attendees WHERE tournament_id = ? AND player_id = ?`
	err := m.DB.QueryRow(q, tournamentId, playerId).Scan(&a.Id, &a.TournamentID, &a.PlayerID)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (m *AttendeeModel) UpdateSeat(id, seat int) error {
	q := `UPDATE attendees SET current_seat = ? WHERE id = ?`
	_, err := m.DB.Exec(q, seat, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *AttendeeModel) ResetSeatPos(tournamentId string) error {
	q := `UPDATE attendees SET current_seat = NULL WHERE tournament_id = ?`
	result, err := m.DB.Exec(q, tournamentId)
	if err != nil {
		return err
	}
	log.Println(result.RowsAffected())
	return nil
}

func (m *AttendeeModel) List(tournamentId string, seeded bool) ([]Attendee, error) {
	attendees := []Attendee{}
	q := `SELECT a.id, a.tournament_id, a.player_id, a.current_seat, p.id, p.name, p.discord_id
		  FROM attendees a JOIN players p ON a.player_id = p.id
		  WHERE a.tournament_id = ? `

	if seeded {
		q += `AND a.current_seat IS NOT NULL`
	}

	rows, err := m.DB.Query(q, tournamentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		a := &Attendee{}
		err := rows.Scan(
			&a.Id, &a.TournamentID, &a.PlayerID, &a.CurrentSeat, &a.Player.ID,
			&a.Player.Name, &a.Player.DiscordID,
		)
		if err != nil {
			log.Fatal(err)
		}
		attendees = append(attendees, *a)
	}

	return attendees, nil
}
