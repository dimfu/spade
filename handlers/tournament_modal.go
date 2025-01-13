package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/models"
)

type TournamentModalHandler struct {
}

func (h *TournamentModalHandler) Name() string {
	return "modals-tournament"
}

func (h *TournamentModalHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// acknowledge modal is submitted so it wont hang
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	db := database.GetDB()
	tm := models.NewTournamentsModel(db)
	data := i.ModalSubmitData()

	// first index is the main component name, second is type of action, third is the unique id
	splitcid := strings.Split(data.CustomID, "_")
	action := splitcid[1]
	id := splitcid[2]

	// TODO: add some kind of gate that prevent the non tournament owner to edit/delete
	switch action {
	case "edit":
		t, err := tm.GetById(id)
		if err != nil {
			respond(err.Error(), s, i, true)
			return
		}

		t.Name = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		if err := tm.Update(t); err != nil {
			respond(err.Error(), s, i, true)
			return
		}

		s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, &discordgo.MessageEmbed{
			Title:       "Configuration",
			Description: "Available configuration for your tournament",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Name", Value: t.Name},
				{Name: "Best Of", Value: "1"},
				{Name: "Player Cap", Value: t.TournamentType.Size},
				{Name: "Bracket Type", Value: "Single Elimination"},
			},
		})
	}
}
