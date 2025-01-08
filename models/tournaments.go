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
	Starting_At         sql.NullString
	Created_At          string
}

type TournamentsModel struct {
	DB *sql.DB
}

func NewTournamentsModel(db *sql.DB) *TournamentsModel {
	return &TournamentsModel{
		DB: db,
	}
}

func (tm *TournamentsModel) GetById(id string) (*Tournament, error) {
	t := &Tournament{}
	q := "SELECT id, name, tournament_types_id, starting_at, created_at FROM tournaments WHERE id = ?"
	err := tm.DB.QueryRow(q, id).Scan(
		&t.ID, &t.Name, &t.Tournament_Types_ID, &t.Starting_At, &t.Created_At,
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
	q := "UPDATE tournaments SET name = ? WHERE id = ?"
	_, err := tm.DB.Exec(q, &t.Name, &t.ID)
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
