package tournament

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/models"
	"github.com/google/uuid"
)

type tournamentChoice struct {
	name  string
	value int
}

type TournamentCreateHandler struct {
	Base              *base.BaseAdmin
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
		if tt.Has_Third_Winner {
			h.tournamentChoices = append(h.tournamentChoices,
				tournamentChoice{
					name: fmt.Sprintf("%v (%v players)", tt.Bracket_Type, tt.Size), value: tt.ID,
				},
			)
		}
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
	err := h.Base.HasPermit(s, i)
	if err != nil {
		base.Respond(err.Error(), s, i, true)
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
		base.SendError(base.ERR_CREATING_TOURNAMENT, s, i)
		return
	}

	t, err := bracket.GenerateFromTemplate(sizeInt)
	if err != nil {
		log.Println(err.Error())
		base.SendError(base.ERR_GENERATE_BRACKET, s, i)
		return
	}

	query := `
        INSERT INTO tournaments (id, name, tournament_types_id, starting_at, created_at) 
        VALUES (?, ?, ?, ?, ?)`

	createdAt := time.Now().Unix()

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		base.SendError(base.ERR_CREATING_TOURNAMENT, s, i)
		return
	}
	defer stmt.Close()

	tName := "New Tournament"

	tId := uuid.New().String()
	_, err = stmt.Exec(tId, tName, tt.ID, nil, createdAt)
	if err != nil {
		log.Println(err.Error())
		base.SendError(base.ERR_CREATING_TOURNAMENT, s, i)
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
					Footer: &discordgo.MessageEmbedFooter{
						Text: tId,
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: "📤",
							},
							Label:    "Publish",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("tournament_publish_%s", tId),
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
