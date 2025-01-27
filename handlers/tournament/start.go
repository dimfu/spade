package tournament

import (
	"context"
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
	"github.com/dimfu/spade/handlers/queue"
	"github.com/dimfu/spade/models"
)

type StartHandler struct {
	Base          base.BaseAdmin
	MatchQueue    queue.MatchQueue
	ctx           context.Context
	db            *sql.DB
	attendeeModel *models.AttendeeModel
}

func (h *StartHandler) WithCtx(ctx context.Context) {
	h.ctx = ctx
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

	prevSize = sizes[0]

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
	if nextMinSize >= len(seatedAttendees) {
		log.Println("readjusting")
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

	bracket, err := bracket.GenerateFromTemplate(bracketSize)
	if err != nil {
		log.Printf("error while generating bracket %v\n", err)
		base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
		return
	}

	// re-adjust the seat positions according to new bracket size if needed
	if shouldReseed || randomize {
		if err = h.reseed(bracket, attendees, strategy); err != nil {
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

	// TODO: should skip all the operations above if the tournament is already started before
	attendees, err = h.attendeeModel.List(string(tournamentId), false)
	if err != nil {
		log.Println(err)
		base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
		return
	}

	matches := h.generateMatches(bracket, attendees)
	go h.MatchQueue.Start(string(tournamentId), bracket, matches, h.ctx, func(match models.Match) {
		// TODO: update winner node to the next bracket based on the template provided
		s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
			Title:       "Match", // TODO: add round count
			Description: fmt.Sprintf("%v vs %v", match.P1.CurrentSeat, match.P2.CurrentSeat),
			// TODO: add buttons to determine which node advances to the next bracket
			// TODO: also clear queue for that tournament_id when clicking the button
		})
	})
}

func (h *StartHandler) reseed(bracket *bracket.BracketTree, attendees []models.Attendee, strategy seeds.Stragies) error {
	var attendeesInterface []interface{}
	for _, a := range attendees {
		attendeesInterface = append(attendeesInterface, a)
	}

	seeds, err := seeds.NewSeeds(attendeesInterface, strategy, len(bracket.StartingSeats))
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

func (h *StartHandler) generateMatches(bracket *bracket.BracketTree, attendees []models.Attendee) []*models.Match {
	matches := make([]*models.Match, 0, len(attendees)/2)
	for i := 0; i < len(bracket.StartingSeats)-1; i += 2 {
		var match models.Match
		for j := 0; j < len(attendees); j++ {
			if bracket.StartingSeats[i] == int(attendees[j].CurrentSeat.Int64) {
				match.P1 = &attendees[j]
			}
			if bracket.StartingSeats[i+1] == int(attendees[j].CurrentSeat.Int64) {
				match.P2 = &attendees[j]
			}
		}
		if match.P1 != nil && match.P2 != nil {
			matches = append(matches, &match)
		}
	}
	return matches
}
