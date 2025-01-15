package tournament

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
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
	case "start":
		h.start(s, i, tm, id)
	case "edit":
		t, err := tm.GetById(id)
		if err != nil {
			handler.Respond(handler.ERR_GET_TOURNAMENT, s, i, true)
			return
		}
		h.edit(s, i, t)
	case "delete":
		err := tm.Delete(id)
		if err != nil {
			handler.Respond(err.Error(), s, i, true)
			return
		}
		h.delete(s, i, id)
	}
}

func (h *TournamentComponentHandler) start(
	s *discordgo.Session, i *discordgo.InteractionCreate, tm *models.TournamentsModel, id string) {
	t, err := tm.GetById(id)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT, s, i, true)
		return
	}

	if t.Has_Started {
		handler.Respond("Tournament has already been started", s, i, true)
		return
	}

	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		fmt.Println("Error deleting message:", err)
		return
	}

	var lastCount int
	var tChannel *discordgo.Channel

	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildCategory && channel.Name == "tournaments" {
			tChannel = channel
		} else if tChannel != nil && channel.ParentID == tChannel.ID {
			c := strings.Split(channel.Name, "-")
			tnumber, err := strconv.Atoi(c[0])
			if err != nil {
				log.Println("error while converting padded integer string to number")
				continue
			}
			lastCount = tnumber
		}
	}

	if tChannel == nil {
		c, err := s.GuildChannelCreate(i.GuildID, "tournaments", discordgo.ChannelTypeGuildCategory)
		if err != nil {
			log.Println("Error while starting tournament:", err)
			return
		}
		tChannel = c
		lastCount = 0
	}

	t.Has_Started = true
	if err = tm.Update(t); err != nil {
		log.Println(err)
		return
	}

	currTChannel, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:     fmt.Sprintf("%04d-%s", lastCount+1, strings.Replace(t.Name, " ", "-", -1)),
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: tChannel.ID,
		Topic:    string(t.ID),
	})

	if err != nil {
		log.Println("Error while creating channel:", err)
		return
	}

	handler.Respond(fmt.Sprintf("Tournament: %s is now started", t.Name), s, i, true)

	_, err = s.ChannelMessageSendEmbed(currTChannel.ID, &discordgo.MessageEmbed{
		Title:       "Configuration",
		Description: "Available configuration for your tournament",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Name", Value: t.Name},
			{Name: "Best Of", Value: "1"}, // TODO: Use dynamic value instead of hard coded value
			{Name: "Player Cap", Value: t.TournamentType.Size},
			{Name: "Bracket Type", Value: "Single Elimination"},
		},
	})

	if err != nil {
		log.Println(err)
	}
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
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

func (h *TournamentComponentHandler) delete(s *discordgo.Session, i *discordgo.InteractionCreate, id string) {
	// delete the embed message of this tournament
	err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		fmt.Println("Error deleting message:", err)
	}

	var tChannel *discordgo.Channel
	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		fmt.Println("Error deleting message:", err)
		return
	}

	for _, channel := range channels {
		if channel.ParentID == "tournaments" && channel.Topic == id {
			tChannel = channel
			break
		}
	}

	// if there is tournament channel for this tournament, channel should be deleted.
	if tChannel != nil {
		_, err = s.ChannelDelete(tChannel.ID)
		if err != nil {
			handler.Respond(fmt.Sprintf("Cannot delete tournament channel while deleting tournament: %v", err), s, i, true)
			return
		}
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
