package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/models"
	"github.com/google/uuid"
)

type tournamentType struct {
	id             int
	size           string
	bracketType    string
	hasThirdWinner bool
}

type tournamentChoice struct {
	name  string
	value int
}

type TournamentCreateHandler struct {
	base              BaseAdminHandler
	tournamentChoices []tournamentChoice
	tournamentTypes   []models.TournamentType
}

func (h *TournamentCreateHandler) Command() *discordgo.ApplicationCommand {
	db := database.GetDB()
	ttm := models.NewTournamentTypesModel(db)
	h.tournamentChoices = make([]tournamentChoice, 0)
	h.tournamentTypes = make([]models.TournamentType, 0)

	t_types, err := ttm.List()
	if err != nil {
		log.Printf("error querying tournament types, ERR: %v", err.Error())
		return nil
	}

	for _, tt := range t_types {
		h.tournamentChoices = append(h.tournamentChoices,
			tournamentChoice{
				name: fmt.Sprintf("%v (%v players)", tt.Bracket_Type, tt.Size), value: tt.ID,
			},
		)
		h.tournamentTypes = append(h.tournamentTypes, tt)
	}

	discordChoices := make([]*discordgo.ApplicationCommandOptionChoice, len(h.tournamentChoices))

	for i, choice := range h.tournamentChoices {
		discordChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  choice.name,
			Value: choice.value,
		}
		discordChoices[i] = discordChoice
	}

	return &discordgo.ApplicationCommand{
		Name:        "create",
		Description: "Create a tournament",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "configurations",
				Description: "Select pre configured tournament settings",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Choices:     discordChoices,
				Required:    true,
			},
		},
	}
}

func (h *TournamentCreateHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	db := database.GetDB()
	err := h.base.HasPermit(s, i)
	if err != nil {
		respond(err.Error(), s, i, true)
		return
	}

	data := i.ApplicationCommandData()
	payload := data.Options[0].IntValue()

	var tt models.TournamentType
	for _, t := range h.tournamentTypes {
		if t.ID == int(payload) {
			tt = t
		}
	}

	sizeInt, err := strconv.Atoi(tt.Size)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT, s, i, true)
		return
	}

	t, err := bracket.GenerateFromTemplate(sizeInt)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_GENERATE_BRACKET, s, i, true)
		return
	}

	query := `
        INSERT INTO tournaments (id, name, tournament_types_id, starting_at, created_at) 
        VALUES (?, ?, ?, ?, ?)`

	createdAt := time.Now().Unix()

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT, s, i, true)
		return
	}
	defer stmt.Close()

	tName := "New Tournament"

	tId := uuid.New().String()
	_, err = stmt.Exec(tId, tName, tt.ID, nil, createdAt)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT, s, i, true)
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Configuration",
					Description: "Available configuration for your tournament",
					Fields: []*discordgo.MessageEmbedField{
						{Name: "Name", Value: tName},
						{Name: "Best Of", Value: "1"}, // TODO: Use dynamic value instead of hard coded value
						{Name: "Player Cap", Value: strconv.Itoa(len(t.StartingSeats))},
						{Name: "Bracket Type", Value: "Single Elimination"},
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: "▶️",
							},
							Label:    "Start",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("tournament_start_%s", tId),
						},
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: "✍",
							},
							Label:    "Edit",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("tournament_edit_%s", tId),
						},
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: "🗑️",
							},
							Label:    "Delete",
							Style:    discordgo.DangerButton,
							CustomID: fmt.Sprintf("tournament_delete_%s", tId),
						},
					},
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}
