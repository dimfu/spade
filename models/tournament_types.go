package models

import (
	"database/sql"
)

type TournamentType struct {
	ID               int
	Size             string
	Bracket_Type     string
	Has_Third_Winner bool
}
type TournamentTypesModel struct {
	DB *sql.DB
}

func NewTournamentTypesModel(db *sql.DB) *TournamentTypesModel {
	return &TournamentTypesModel{
		DB: db,
	}
}

func (ttm *TournamentTypesModel) List() ([]TournamentType, error) {
	types := []TournamentType{}
	q := "SELECT id, size, bracket_type, has_third_winner FROM tournament_types"
	rows, err := ttm.DB.Query(q)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tt TournamentType
		err := rows.Scan(&tt.ID, &tt.Size, &tt.Bracket_Type, &tt.Has_Third_Winner)
		if err != nil {
			return nil, err
		}
		types = append(types, tt)
	}

	return types, nil
}
