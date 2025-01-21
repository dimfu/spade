package models

import (
	"database/sql"
)

type Player struct {
	ID        []uint8
	Name      string
	DiscordID string
}

type PlayerModel struct {
	DB *sql.DB
}

func NewPlayerModel(db *sql.DB) *PlayerModel {
	return &PlayerModel{
		DB: db,
	}
}

func (m *PlayerModel) FindById(id string) (*Player, error) {
	p := &Player{}
	q := `SELECT id, name, discord_id FROM players WHERE id = ?`

	err := m.DB.QueryRow(q, id).Scan(&p.ID, &p.Name, &p.DiscordID)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (m *PlayerModel) FindByDiscordId(discordId string) (*Player, error) {
	p := &Player{}
	q := `SELECT id, name, discord_id FROM players WHERE discord_id = ?`

	err := m.DB.QueryRow(q, discordId).Scan(&p.ID, &p.Name, &p.DiscordID)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (m *PlayerModel) Insert(tx *sql.Tx, p *Player) error {
	q := `INSERT INTO players (id, name, discord_id) VALUES (?, ?, ?)`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(p.ID, p.Name, p.DiscordID)
	if err != nil {
		return err
	}

	return nil
}
