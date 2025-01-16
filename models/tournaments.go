package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Tournament struct {
	ID                  []uint8
	Name                string
	Tournament_Types_ID int
	Thread_ID           sql.NullString
	Published           bool
	Starting_At         sql.NullString
	Created_At          string
	TournamentType      TournamentType
}

type TournamentsModel struct {
	DB *sql.DB
}

func NewTournamentsModel(db *sql.DB) *TournamentsModel {
	return &TournamentsModel{
		DB: db,
	}
}

func (tm *TournamentsModel) Length() (int, error) {
	var count int
	q := "SELECT COUNT(*) FROM tournaments"
	err := tm.DB.QueryRow(q).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (tm *TournamentsModel) GetTournamentIDInThread(threadID string) ([]uint8, error) {
	var tournamentID []uint8
	q := `SELECT id FROM tournaments WHERE thread_id = ?`
	err := tm.DB.QueryRow(q, threadID).Scan(&tournamentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(fmt.Sprintf("Could not find any record that have thread id of %s", threadID))
		}
		return nil, fmt.Errorf("Error querying tournament: %v", err.Error())
	}
	return tournamentID, nil
}

func (tm *TournamentsModel) GetById(id string) (*Tournament, error) {
	t := &Tournament{}
	var published int
	q := `
		SELECT t.id, t.name, t.tournament_types_id, t.starting_at, t.created_at, t.published, t.thread_id,
			   tt.id, tt.size, tt.bracket_type, tt.has_third_winner
	 	FROM tournaments t
		JOIN tournament_types tt ON t.tournament_types_id = tt.id
		WHERE t.id = ?`

	err := tm.DB.QueryRow(q, id).Scan(
		&t.ID, &t.Name, &t.Tournament_Types_ID, &t.Starting_At, &t.Created_At, &published, &t.Thread_ID,
		&t.TournamentType.ID, &t.TournamentType.Size, &t.TournamentType.Bracket_Type,
		&t.TournamentType.Has_Third_Winner,
	)

	t.Published = published == 1

	if err == sql.ErrNoRows {
		return nil, errors.New(fmt.Sprintf("Could not find any record that have id of %s", id))
	}

	return t, nil
}

func (tm *TournamentsModel) Update(t *Tournament) error {
	q := "UPDATE tournaments SET name = ?, published = ?, thread_id = ? WHERE id = ?"
	_, err := tm.DB.Exec(q, t.Name, t.Published, t.Thread_ID, t.ID)
	if err != nil {
		return err
	}
	return nil
}

func (tm *TournamentsModel) Delete(id string) (*Tournament, error) {
	t, err := tm.GetById(id)

	if err != nil {
		return nil, err
	}

	q := "DELETE FROM tournaments WHERE id = ?"
	_, err = tm.DB.Exec(q, id)
	return t, err
}
