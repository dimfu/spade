package tournament

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/handlers/queue"
	"github.com/dimfu/spade/models"
)

type RestartTournamentHandler struct {
	Base       *base.BaseAdmin
	MatchQueue *queue.MatchQueue
	db         *sql.DB
}

func (h *RestartTournamentHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "restart",
		Description: "Restart the tournament",
	}
}

func (h *RestartTournamentHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		base.Respond(err.Error(), s, i, true)
		return
	}

	h.db = database.GetDB()
	tm := models.NewTournamentsModel(h.db)
	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		base.SendError(err, s, i)
		return
	}
	_, err = tm.GetById(string(tournamentId))
	if err != nil {
		base.SendError(err, s, i)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		base.SendError(err, s, i)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			base.Respond("Something went wrong while restarting the tournament.", s, i, true)
		}
	}()

	err = h.restart(tx, string(tournamentId))
	if err != nil {
		base.SendError(err, s, i)
		return
	}
	if err = h.MatchQueue.ClearQueue(string(tournamentId)); err != nil {
		base.SendError(err, s, i)
		return
	}

	base.Respond("Tournament has been restarted, use /start to start again.", s, i, false)
}

func (h *RestartTournamentHandler) restart(tx *sql.Tx, tournamentID string) error {
	s := make([]string, 0)
	s = append(s, "UPDATE attendees SET current_seat = starting_seat WHERE tournament_id = ?")
	s = append(s, "UPDATE tournaments SET starting_at = NULL WHERE id = ?")
	s = append(s, "DELETE FROM match_histories WHERE attendee_id IN (SELECT id FROM attendees WHERE tournament_id = ?)")

	for _, q := range s {
		_, err := tx.Exec(q, tournamentID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
