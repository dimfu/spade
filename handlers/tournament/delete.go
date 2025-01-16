package tournament

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
)

type TournamentDeleteHandler struct {
	Base handler.BaseAdmin
}

func (h *TournamentDeleteHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "delete",
		Description: "Delete current tournament",
	}
}

func (h *TournamentDeleteHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		log.Println(err)
		return
	}

	db := database.GetDB()
	tm := models.NewTournamentsModel(db)

	tId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	_, err = tm.Delete(string(tId))
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	// TODO: need to delete all attendees

	_, err = s.ChannelDelete(i.ChannelID)
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}
}
