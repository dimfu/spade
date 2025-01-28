package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type History struct {
	ID         int
	AttendeeID int
	Result     int
	Seat       sql.NullInt64
	CreatedAt  sql.NullInt64
}

type MatchHistory struct {
	Attendee  Attendee
	Histories []History
}

type MatchHistoryModel struct {
	DB *sql.DB
}

func NewMatchHistoryModel(db *sql.DB) *MatchHistoryModel {
	return &MatchHistoryModel{
		DB: db,
	}
}

func (m *MatchHistoryModel) Insert(tx *sql.Tx, h *History) error {
	now := time.Now().Unix()
	q := `INSERT INTO match_histories (attendee_id, result, seat, created_at) VALUES (?, ?, ?, ?)`
	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(h.AttendeeID, h.Result, h.Seat, now)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("No record updated")
	}
	return nil
}

func (m *MatchHistoryModel) CurrentTournamentHistory(tournamentID []uint8) ([]MatchHistory, error) {
	q := `SELECT
			mh.id AS history_id, 
			mh.result, 
			mh.seat, 
			mh.created_at, 
			a.id AS attendee_id, 
			a.tournament_id, 
			a.player_id, 
			a.current_seat,
			p.name,
			p.discord_id
		FROM attendees a
		LEFT JOIN match_histories mh ON mh.attendee_id = a.id
		LEFT JOIN players p ON p.id = a.player_id
		WHERE a.tournament_id = ? AND a.current_seat IS NOT NULL;
		`
	rows, err := m.DB.Query(q, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	historyMap := make(map[int]*MatchHistory)

	for rows.Next() {
		var (
			historyID       sql.NullInt64
			result          sql.NullInt64
			seat            sql.NullInt64
			createdAt       sql.NullInt64
			attendeeID      int
			tournamentID    string
			playerID        []uint8
			currentSeat     int
			playerName      string
			playerDiscordID string
		)

		if err := rows.Scan(
			&historyID, &result, &seat, &createdAt, &attendeeID, &tournamentID,
			&playerID, &currentSeat, &playerName, &playerDiscordID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := historyMap[attendeeID]; !exists {
			historyMap[attendeeID] = &MatchHistory{
				Attendee: Attendee{
					Id:           attendeeID,
					TournamentID: tournamentID,
					PlayerID:     string(playerID),
					CurrentSeat:  sql.NullInt64{Int64: int64(currentSeat), Valid: true},
					Player: Player{
						ID:        playerID,
						Name:      playerName,
						DiscordID: playerDiscordID,
					},
				},
				Histories: []History{},
			}
		}
		if historyID.Valid {
			historyMap[attendeeID].Histories = append(historyMap[attendeeID].Histories, History{
				ID:     int(historyID.Int64),
				Result: int(result.Int64),
				Seat: sql.NullInt64{
					Int64: seat.Int64,
					Valid: seat.Valid,
				},
				CreatedAt: sql.NullInt64{
					Int64: createdAt.Int64,
					Valid: createdAt.Valid,
				},
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	var matchHistories []MatchHistory
	for _, matchHistory := range historyMap {
		matchHistories = append(matchHistories, *matchHistory)
	}

	return matchHistories, nil
}
