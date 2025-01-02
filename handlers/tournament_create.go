package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/database"
)

type TournamentCreateHandler struct {
	base BaseAdminHandler
}

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

var (
	db *sql.DB

	ERR_CREATING_TOURNAMENT = "Something went wrong while creating tournament"
	ERR_GENERATE_BRACKET    = "Error occured when generating tournament bracket template"

	tournamentChoices = []tournamentChoice{}
	tournamentTypes   = []tournamentType{}
)

func (h *TournamentCreateHandler) Command() *discordgo.ApplicationCommand {
	db = database.GetDB()
	q := "SELECT id, size, bracket_type, has_third_winner FROM tournament_types"
	rows, err := db.Query(q)

	if err != nil {
		log.Printf("error querying tournament types, ERROR: %v", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var tt tournamentType
		err := rows.Scan(&tt.id, &tt.size, &tt.bracketType, &tt.hasThirdWinner)
		if err != nil {
			log.Printf("error scanning rows, ERROR: %v", err)
		}
		tournamentChoices = append(tournamentChoices,
			tournamentChoice{
				name: fmt.Sprintf("%v (%v players)", tt.bracketType, tt.size), value: tt.id,
			},
		)
		tournamentTypes = append(tournamentTypes, tt)
	}

	discordChoices := make([]*discordgo.ApplicationCommandOptionChoice, len(tournamentChoices))

	for i, choice := range tournamentChoices {
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

func (p *TournamentCreateHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respond := func(r string) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: r,
			},
		})
	}
	err := p.base.HasPermit(s, i)
	if err != nil {
		respond(err.Error())
		return
	}

	data := i.ApplicationCommandData()
	payload := data.Options[0].IntValue()

	var tt tournamentType
	for _, t := range tournamentTypes {
		if t.id == int(payload) {
			tt = t
		}
	}

	/**
	TODO: add buttons to embed:
		- register/release tournament and create the channel for it
		- unpublish tournament
		- edit
		- remove
	*/

	sizeInt, err := strconv.Atoi(tt.size)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT)
		return
	}

	t, err := bracket.GenerateFromTemplate(sizeInt)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_GENERATE_BRACKET)
		return
	}

	query := `
        INSERT INTO tournaments (id, name, tournament_types_id, starting_at, created_at) 
        VALUES (UUID(), ?, ?, ?, ?)`

	createdAt := time.Now().Unix()

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT)
		return
	}
	defer stmt.Close()

	tName := "New Tournament"

	_, err = stmt.Exec(tName, tt.id, nil, createdAt)
	if err != nil {
		log.Println(err.Error())
		respond(ERR_CREATING_TOURNAMENT)
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
						{Name: "Best Of", Value: "1"},
						{Name: "Player Cap", Value: strconv.Itoa(len(t.StartingSeats))},
						{Name: "Bracket Type", Value: "Single Elimination"},
					},
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}
