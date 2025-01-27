package tournament

import (
	"database/sql"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/base"
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
			base.Respond(err.Error(), s, i, true)
			return
		}

		t.Name = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		rules := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		fields := []*discordgo.MessageEmbedField{
			{Name: "Name", Value: t.Name},
		}

		t.Description = sql.NullString{
			String: description,
			Valid:  len(description) != 0,
		}
		if len(description) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Description",
				Value: t.Description.String,
			})
		}

		t.Rules = sql.NullString{
			String: rules,
			Valid:  len(rules) != 0,
		}
		if len(rules) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Rules",
				Value: t.Rules.String,
			})
		}

		fields = append(fields,
			&discordgo.MessageEmbedField{Name: "Best Of", Value: "1"},
			&discordgo.MessageEmbedField{Name: "Player Cap", Value: t.TournamentType.Size},
			&discordgo.MessageEmbedField{Name: "Bracket Type", Value: "Single Elimination"},
		)

		if err := tm.Update(t); err != nil {
			base.Respond(err.Error(), s, i, true)
			return
		}

		editEmbed := func(chId, msgId string) {
			s.ChannelMessageEditEmbed(chId, msgId, &discordgo.MessageEmbed{
				Title:       "Configuration",
				Description: "Available configuration for your tournament",
				Fields:      fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: string(t.ID),
				},
			})
		}

		// update tournament embed inside the published channel
		if t.Published {
			msgs, err := s.ChannelMessagesPinned(t.Thread_ID.String)
			if err != nil {
				log.Println(err.Error())
				return
			}
			var embedID string
			for _, msg := range msgs {
				if msg.Author.ID == s.State.User.ID && len(msg.Embeds) > 0 {
					embedID = msg.ID
					break
				}
			}
			editEmbed(t.Thread_ID.String, embedID)
		}

		editEmbed(i.ChannelID, i.Message.ID)
	}
}
