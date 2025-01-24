package tournament

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/bracket/seeds"
	"github.com/dimfu/spade/bracket/templates"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
)

type StartHandler struct {
	Base handler.BaseAdmin
	db   *sql.DB
}

func (h *StartHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "start",
		Description: "Start the tournament",
	}
}

func (h *StartHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	h.db = database.GetDB()
	tm := models.NewTournamentsModel(h.db)
	ttm := models.NewTournamentTypesModel(h.db)
	am := models.NewAttendeeModel(h.db)

	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		log.Println(err)
		handler.Respond(handler.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	tournament, err := tm.GetById(string(tournamentId))
	if err != nil {
		log.Println(err)
		handler.Respond(handler.ERR_GET_TOURNAMENT, s, i, true)
		return
	}

	attendees, err := am.List(string(tournamentId), true)
	if err != nil {
		log.Println(err)
		handler.Respond("Cannot get attendees before starting tournament", s, i, true)
		return
	}

	tSize, _ := strconv.Atoi(tournament.TournamentType.Size)
	sizes := make([]int, 0, len(templates.Templates))
	for k := range templates.Templates {
		sizes = append(sizes, k)
	}
	sort.Ints(sizes)

	var prevSize, nextMinSize int
	sg := make(map[int]int)

	for _, size := range sizes {
		if size == tSize {
			nextMinSize = prevSize
		}
		gap := size - len(attendees)
		sg[size] = int(gap)
		prevSize = size
	}

	bracketSize := tSize

	tx, err := h.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	if nextMinSize > len(attendees) {
		minGap := math.MaxInt64
		minSize := math.MaxInt64
		for size, gap := range sg {
			if gap >= 0 && gap < minGap {
				minGap = gap
				bracketSize = size
			}
			minSize = int(math.Min(float64(minSize), float64(size)))
		}
		if minSize > len(attendees) {
			handler.Respond(fmt.Sprintf("Can't start tournament, you need at least %d seeded players to start", bracketSize), s, i, true)
			return
		}

		tournamentTypes, err := ttm.List()
		if err != nil {
			handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
			log.Printf("error while getting tournament types, %v", err)
			return
		}

		// TODO: update current tournament bracket size to be `bracketSize`
		var newTType int
		for _, tt := range tournamentTypes {
			size, _ := strconv.Atoi(tt.Size)
			if bracketSize == size && tournament.TournamentType.Has_Third_Winner == tt.Has_Third_Winner {
				newTType = tt.ID
				break
			}
		}

		// re-adjust the seat positions according to new bracket size
		bracket, err := bracket.GenerateFromTemplate(bracketSize)
		if err != nil {
			handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
			log.Println(err)
			return
		}

		var attendeesInterface []interface{}
		for _, a := range attendees {
			attendeesInterface = append(attendeesInterface, a)
		}
		// TODO: seed strategy should be coming from the current tournament field
		seeds, err := seeds.NewSeeds(attendeesInterface, seeds.BEST_AGAINST_WORST, bracketSize)
		if err != nil {
			handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
			log.Println(err)
			return
		}

		for seat, seed := range seeds {
			if seed == nil {
				continue
			}

			attendee, ok := seed.(models.Attendee)
			if !ok {
				log.Printf("%v is not valid attendee model\n", attendee)
				continue
			}

			err = am.UpdateSeat(attendee.Id, bracket.StartingSeats[seat])
			if err != nil {
				handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
				log.Printf("error while updating seat position, %v", err)
				return
			}
		}

		tournament.Tournament_Types_ID = newTType
		if err = tm.Update(tournament); err != nil {
			handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
			log.Printf("error while updating tournament %v\n", err)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		handler.Respond(handler.ERR_INTERNAL_ERROR, s, i, true)
		log.Println(err)
		return
	}

	handler.Respond(fmt.Sprintf("%d size is suited", bracketSize), s, i, true)
	// TODO: post current match embed maybe? also make a queue for the next matches for this tournament so after once match is finish we immediately post the next embed without looking up to db
}
