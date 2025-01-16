package tournament

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/config"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
)

type TournamentComponentHandler struct {
	Base handler.BaseAdmin
}

func (h *TournamentComponentHandler) Name() string {
	return "tournament"
}

func (h *TournamentComponentHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	db := database.GetDB()
	tm := models.NewTournamentsModel(db)

	cid := i.MessageComponentData().CustomID

	splitcid := strings.Split(cid, "_")
	action := splitcid[1]
	id := splitcid[2]

	switch action {
	case "publish":
		h.publish(s, i, tm, id)
	case "edit":
		t, err := tm.GetById(id)
		if err != nil {
			handler.Respond(handler.ERR_GET_TOURNAMENT, s, i, true)
			return
		}
		h.edit(s, i, t)
	case "delete":
		h.delete(s, i, tm, id)
	}
}

func (h *TournamentComponentHandler) publish(
	s *discordgo.Session, i *discordgo.InteractionCreate, tm *models.TournamentsModel, id string) {
	cfg := config.GetEnv()
	t, err := tm.GetById(id)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT, s, i, true)
		return
	}

	if t.Published {
		handler.Respond("Tournament has already been published", s, i, true)
		return
	}

	thread, err := s.ThreadStartComplex(cfg.TOURNAMENT_CHANNEL_ID, &discordgo.ThreadStart{
		Name: t.Name,
		Type: discordgo.ChannelTypeGuildPublicThread,
	})

	if err != nil {
		log.Println(err)
		handler.Respond("Failed to create thread for the tournament", s, i, true)
		return
	}

	t.Published = true
	t.Thread_ID = sql.NullString{
		String: thread.ID,
		Valid:  thread.ID != "",
	}

	if err = tm.Update(t); err != nil {
		log.Println(err)
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Name", Value: t.Name},
	}

	if len(t.Description.String) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Description",
			Value: t.Description.String,
		})
	}

	if len(t.Rules.String) > 0 {
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

	e, err := s.ChannelMessageSendEmbed(thread.ID, &discordgo.MessageEmbed{
		Title:       "Configuration",
		Description: "Available configuration for your tournament",
		Fields:      fields,
	})

	if err != nil {
		log.Println(err)
		return
	}

	err = s.ChannelMessagePin(thread.ID, e.ID)
	if err != nil {
		log.Println(err)
		return
	}

	handler.Respond(fmt.Sprintf("Tournament <#%s> is published to the public", thread.ID), s, i, true)
}

func (h *TournamentComponentHandler) edit(
	s *discordgo.Session, i *discordgo.InteractionCreate, t *models.Tournament) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "description",
							Placeholder: "Describe what's this tournament all about",
							Label:       "Description",
							Style:       discordgo.TextInputParagraph,
							Required:    false,
							MaxLength:   2000,
							MinLength:   0,
							Value:       t.Description.String,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "rules",
							Placeholder: "Rules if they are applicable",
							Label:       "Rules",
							Style:       discordgo.TextInputParagraph,
							Required:    false,
							MaxLength:   2000,
							MinLength:   0,
							Value:       t.Rules.String,
						},
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

func (h *TournamentComponentHandler) delete(s *discordgo.Session, i *discordgo.InteractionCreate,
	tm *models.TournamentsModel, id string) {
	// delete the embed message of this tournament
	err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		fmt.Println("Error deleting message:", err)
	}

	var threadID sql.NullString
	t, err := tm.Delete(id)
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
	}

	threadID = t.Thread_ID
	if t != nil && threadID.Valid {
		_, err = s.ChannelDelete(threadID.String)
		if err != nil {
			handler.Respond(fmt.Sprintf("Cannot delete tournament channel while deleting tournament: %v", err), s, i, true)
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Tournament successfully deleted",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		var httpErr *discordgo.RESTError
		// ignore if the message is already acknowledged above when handling err
		if errors.As(err, &httpErr) && httpErr.Message.Code == discordgo.ErrCodeInteractionHasAlreadyBeenAcknowledged {
			return
		}
		log.Print(err.Error())
	}
}
