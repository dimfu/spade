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
	Has_Started         bool
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

func (tm *TournamentsModel) GetById(id string) (*Tournament, error) {
	t := &Tournament{}
	q := `
		SELECT t.id, t.name, t.tournament_types_id, t.starting_at, t.created_at,
			   tt.id, tt.size, tt.bracket_type, tt.has_third_winner
	 	FROM tournaments t
		JOIN tournament_types tt ON t.tournament_types_id = tt.id
		WHERE t.id = ?`

	err := tm.DB.QueryRow(q, id).Scan(
		&t.ID, &t.Name, &t.Tournament_Types_ID, &t.Starting_At, &t.Created_At,
		&t.TournamentType.ID, &t.TournamentType.Size, &t.TournamentType.Bracket_Type,
		&t.TournamentType.Has_Third_Winner,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(fmt.Sprintf("Could not find any record that have id of %s", id))
	}

	if err != nil {
		return nil, err
	}
	return t, nil
}

func (tm *TournamentsModel) Update(t *Tournament) error {
	// TODO: update starting_at when tournament is releasing
	q := "UPDATE tournaments SET name = ?, has_started = ? WHERE id = ?"
	_, err := tm.DB.Exec(q, &t.Name, &t.Has_Started, &t.ID)
	if err != nil {
		return err
	}
	return nil
}

func (tm *TournamentsModel) Delete(id string) error {
	_, err := tm.GetById(id)

	if err != nil {
		return err
	}

	q := "DELETE FROM tournaments WHERE id = ?"
	_, err = tm.DB.Exec(q, id)
	return err
}
