package tournament

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/bracket/seeds"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/models"
)

type SeedHandler struct {
	Base *base.BaseAdmin
	db   *sql.DB
}

func (h *SeedHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "seed",
		Description: "Define tournament seeds by order ascending order",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "players",
				Description: "Player(s) to be registered",
				Required:    true,
			},
		},
	}
}

func (h *SeedHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		base.Respond("Insufficent permission to seed other players to this tournament.", s, i, true)
		return
	}

	h.db = database.GetDB()
	tm := models.NewTournamentsModel(h.db)
	pm := models.NewPlayerModel(h.db)
	am := models.NewAttendeeModel(h.db)

	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		base.Respond("You need atleast seed 1 player to this tournament", s, i, true)
		return
	}

	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		base.Respond(base.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	t, err := tm.GetById(string(tournamentId))
	if err != nil {
		base.Respond(err.Error(), s, i, true)
		return
	}

	tSize, err := strconv.Atoi(t.TournamentType.Size)
	if err != nil {
		log.Fatal(err)
	}

	fields := strings.Fields(data.Options[0].StringValue())

	maxlen := len(fields)
	if len(fields) > tSize {
		maxlen = tSize
	}

	players := fields[0:maxlen]

	bracket, err := bracket.GenerateFromTemplate(tSize)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("error starting transaction %v", err)
		return
	}
	defer tx.Rollback()

	var errMsg string
	defer func() {
		if errMsg != "" {
			base.Respond("Something went wrong while seeding players", s, i, true)
		}
	}()

	if err := am.ResetSeatPos(string(tournamentId)); err != nil {
		errMsg = fmt.Sprintf("error while resetting seat pos %v", err)
		tx.Rollback()
		return
	}

	var playersInterface []interface{}
	for _, player := range players {
		playersInterface = append(playersInterface, player)
	}
	seeds, err := seeds.NewSeeds(playersInterface, seeds.BEST_AGAINST_WORST, tSize)

	var countSuccess int
	for i, seed := range seeds {
		if seed == nil {
			continue
		}
		player := seed.(string)
		discordId := player
		if strings.HasPrefix(player, "<@") {
			discordId = player[2 : len(player)-1]
		}
		p, err := pm.FindByDiscordId(discordId)
		if err != nil {
			errMsg = fmt.Sprintf("error finding discord id of %v", discordId)
			return
		}

		a, err := am.FindById(string(tournamentId), string(p.ID))
		if err != nil {
			errMsg = fmt.Sprintf("error finding attendee id of this tournament %v", err)
			return
		}

		err = am.StartingSeat(a.Id, bracket.StartingSeats[i])
		if err != nil {
			errMsg = fmt.Sprintf("error updating seat position %v", err)
			return
		}
		countSuccess++
	}

	if err := tx.Commit(); err != nil {
		errMsg = err.Error()
		return
	}

	base.Respond(fmt.Sprintf("Successfully seeded %d players", countSuccess), s, i, true)
}
