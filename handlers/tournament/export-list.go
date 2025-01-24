package tournament

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
)

type ExportListHandler struct {
	Base *handler.BaseAdmin
	db   *sql.DB
}

type extractType = int64

const (
	PARTICIANTS extractType = iota
	SEEDS
)

func (h *ExportListHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "export",
		Description: "Create a tournament",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "type",
				Description: "Type of list you want to export",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "participants",
						Value: 0,
					},
					{
						Name:  "seed",
						Value: 1,
					},
				},
			},
		},
	}
}

func (h *ExportListHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.db = database.GetDB()
	err := h.Base.HasPermit(s, i)
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	tm := models.NewTournamentsModel(h.db)
	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	am := models.NewAttendeeModel(h.db)

	cdata := i.ApplicationCommandData()
	listType := PARTICIANTS
	if len(cdata.Options) != 0 {
		listType = cdata.Options[0].IntValue()
	}

	var (
		seeded bool
		data   [][]string
	)

	// csv header
	data = append(data, []string{"name", "discord_id", "current_seat"})

	if listType == SEEDS {
		seeded = true
	}

	list, err := am.List(string(tournamentId), seeded)
	if err != nil {
		log.Println(err.Error())
		handler.Respond("Cannot extract participants list", s, i, true)
		return
	}

	if len(list) == 0 {
		handler.Respond("No records to print", s, i, true)
		return
	}

	for _, l := range list {
		currSeat := strconv.Itoa(int(l.CurrentSeat.Int64))
		data = append(data, []string{l.Player.Name, fmt.Sprintf("<@%s>", l.Player.DiscordID), currSeat})
	}

	buf, err := toCSV(data)
	if err != nil {
		panic(err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{Name: fmt.Sprintf("%s.csv", tournamentId), ContentType: "text/csv", Reader: buf},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

func toCSV(data [][]string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	defer writer.Flush()

	for _, value := range data {
		if err := writer.Write(value); err != nil {
			return &buffer, err
		}
	}

	return &buffer, nil
}
