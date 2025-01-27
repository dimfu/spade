package tournament

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/models"
)

type TournamentDeleteHandler struct {
	Base *base.BaseAdmin
}

func (h *TournamentDeleteHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "delete",
		Description: "Delete current tournament",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "Tournament id (optional)",
				Required:    false,
			},
		},
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

	var tId []uint8
	var targetChannel string

	data := i.ApplicationCommandData()
	providedTID := data.Options[0].StringValue()

	if len(providedTID) == 0 {
		tId, err = tm.GetTournamentIDInThread(i.ChannelID)
		if err != nil {
			base.Respond(base.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
			return
		}
		targetChannel = i.ChannelID
	} else {
		tId = []uint8(providedTID)
		t, err := tm.GetById(providedTID)
		if err != nil {
			log.Println(err.Error())
			return
		}
		targetChannel = t.Thread_ID.String
	}

	_, err = tm.Delete(string(tId))
	if err != nil {
		base.Respond(err.Error(), s, i, true)
		return
	}

	if len(targetChannel) > 0 {
		_, err = s.ChannelDelete(targetChannel)
		if err != nil {
			base.Respond(err.Error(), s, i, true)
			return
		}
	}

	// TODO: need to delete all attendees
	base.Respond("Tournament successfully deleted", s, i, true)
}
