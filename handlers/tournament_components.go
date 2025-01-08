package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/models"
)

type TournamentComponentHandler struct {
	base BaseAdminHandler
}

func (tch *TournamentComponentHandler) Name() string {
	return "tournament"
}

func (tch *TournamentComponentHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := tch.base.HasPermit(s, i)
	if err != nil {
		respond(err.Error(), s, i, true)
		return
	}

	db := database.GetDB()
	tm := models.NewTournamentsModel(db)

	cid := i.MessageComponentData().CustomID

	// first index is the main component name, second is type of action, third is the unique id
	splitcid := strings.Split(cid, "_")
	action := splitcid[1]
	id := splitcid[2]

	switch action {
	/**
	* NOTE: for some reason discord not yet implemented select menu in modals, so in order
			to edit the bracket type you need to create a new tournament, lol.
	*/
	case "edit":
		t, err := tm.GetById(id)
		if err != nil {
			respond(ERR_GET_TOURNAMENT, s, i, true)
			return
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "modals-tournament_edit_" + string(t.ID),
				Title:    "Edit Tournament",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:  "name",
								Label:     "Tournament Name",
								Style:     discordgo.TextInputShort,
								Required:  true,
								MaxLength: 128,
								MinLength: 5,
								Value:     t.Name,
							},
						},
					},
				},
			},
		})

		if err != nil {
			panic(err)
		}

		if err != nil {
			log.Print(err.Error())
		}
	case "delete":
		err := tm.Delete(id)
		if err != nil {
			respond(err.Error(), s, i, true)
			return
		}

		// delete the embed message of this tournament
		err = s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		if err != nil {
			fmt.Println("Error deleting message:", err)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Tournament ID: %s deleted", id),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if err != nil {
			log.Print(err.Error())
		}
	}
}
