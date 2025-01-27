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
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/models"
)

type StartHandler struct {
	Base          base.BaseAdmin
	db            *sql.DB
	attendeeModel *models.AttendeeModel
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
		base.Respond(err.Error(), s, i, true)
		return
	}

	h.db = database.GetDB()
	tm := models.NewTournamentsModel(h.db)
	ttm := models.NewTournamentTypesModel(h.db)
	h.attendeeModel = models.NewAttendeeModel(h.db)

	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		log.Println(err)
		base.Respond(base.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	tournament, err := tm.GetById(string(tournamentId))
	if err != nil {
		log.Println(err)
		base.Respond(base.ERR_GET_TOURNAMENT, s, i, true)
		return
	}

	attendees, err := h.attendeeModel.List(string(tournamentId), false)
	if err != nil {
		log.Println(err)
		base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
		return
	}

	if len(attendees) == 0 {
		base.Respond("Not enough seed to start the tournament", s, i, true)
		return
	}

	var randomize bool
	seatedAttendees := make([]models.Attendee, 0, len(attendees))
	for _, a := range attendees {
		if a.CurrentSeat.Valid {
			seatedAttendees = append(seatedAttendees, a)
		}
	}

	// if no seeds provided it should seed with random strategy
	if len(seatedAttendees) == 0 {
		randomize = true
		seatedAttendees = attendees
	}

	// TODO: seed strategy should be coming from the current tournament field
	strategy := seeds.BEST_AGAINST_WORST
	if randomize {
		strategy = seeds.RANDOM
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
		gap := size - len(seatedAttendees)
		sg[size] = int(gap)
		prevSize = size
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	var shouldReseed bool
	bracketSize := tSize

	// handle if attendees count is way lower than the current used bracket size
	if nextMinSize > len(seatedAttendees) {
		shouldReseed = true // should reseed because we change the bracket size
		minGap := math.MaxInt64
		minSize := math.MaxInt64
		for size, gap := range sg {
			if gap >= 0 && gap < minGap {
				// select the bracket size that closest above the current attendees count
				minGap = gap
				bracketSize = size
			}
			minSize = int(math.Min(float64(minSize), float64(size)))
		}
		if minSize > len(seatedAttendees) {
			base.Respond(fmt.Sprintf("Can't start tournament, you need at least %d seeded players to start", bracketSize), s, i, true)
			return
		}

		tournamentTypes, err := ttm.List()
		if err != nil {
			base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
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

		tournament.Tournament_Types_ID = newTType
		if err = tm.Update(tournament); err != nil {
			base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
			log.Printf("error while updating tournament %v\n", err)
			return
		}
	}

	// re-adjust the seat positions according to new bracket size if needed
	if shouldReseed || randomize {
		if err = h.reseed(attendees, strategy, bracketSize); err != nil {
			base.Respond("Something went wrong when re-adjusting seat position", s, i, true)
			log.Println(err)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
		log.Println(err)
		return
	}

	base.Respond(fmt.Sprintf("%d size is suited", bracketSize), s, i, true)
	// TODO: post current match embed maybe? also make a queue for the next matches for this tournament so after once match is finish we immediately post the next embed without looking up to db
}

func (h *StartHandler) reseed(attendees []models.Attendee, strategy seeds.Stragies, bracketSize int) error {
	bracket, err := bracket.GenerateFromTemplate(bracketSize)
	if err != nil {
		return err
	}

	var attendeesInterface []interface{}
	for _, a := range attendees {
		attendeesInterface = append(attendeesInterface, a)
	}

	seeds, err := seeds.NewSeeds(attendeesInterface, strategy, bracketSize)
	if err != nil {
		return err
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

		err = h.attendeeModel.UpdateSeat(attendee.Id, bracket.StartingSeats[seat])
		if err != nil {
			return err
		}
	}
	return nil
}
