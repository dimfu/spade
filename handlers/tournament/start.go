package tournament

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/bracket/seeds"
	"github.com/dimfu/spade/bracket/templates"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/discord/components"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/handlers/queue"
	"github.com/dimfu/spade/models"
)

type StartHandler struct {
	Base          *base.BaseAdmin
	MatchQueue    *queue.MatchQueue
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
	tSize, _ := strconv.Atoi(tournament.TournamentType.Size)

	// if tournament has been already started before, it should skip all checks below.
	if tournament.Starting_At.Valid {
		bracket, err := bracket.GenerateFromTemplate(tSize)
		if err != nil {
			log.Printf("error while generating bracket %v\n", err)
			base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
			return
		}
		err = h.start(tournamentId, bracket, func(match models.Match) {
			h.buildEmbed(s, i, match)
		})
		if err != nil {
			log.Printf("Error when starting tournament %v\n", err)
			base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
			return
		}
		base.Respond("Tournament has already been started, resuming with previous result.", s, i, true)
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

	err = h.start(tournamentId, bracket, func(match models.Match) {
		h.buildEmbed(s, i, match)
	})
	if err != nil {
		log.Printf("Error when starting tournament %v\n", err)
		base.Respond(base.ERR_INTERNAL_ERROR, s, i, true)
		return
	}

	base.Respond("Tournament is now started", s, i, false)
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
	seeds = seeds[0:len(bracket.StartingSeats)] // make sure we limit seed to bracket size

	for seat, seed := range seeds {
		if seed == nil {
			continue
		}

		attendee, ok := seed.(models.Attendee)
		if !ok {
			log.Printf("%v is not valid attendee model\n", attendee)
			continue
		}

		err = h.attendeeModel.StartingSeat(attendee.Id, bracket.StartingSeats[seat])
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *StartHandler) assertPayload(node *bracket.Node) (*models.AttendeeWithResult, error) {
	payload, ok := node.Payload.(models.AttendeeWithResult)
	if !ok {
		return nil, errors.New("payload must be AttendeeWithStatus")
	}
	return &payload, nil
}

func (h *StartHandler) generateMatches(b *bracket.BracketTree, r int) ([]*models.Match, error) {
	nodes, err := b.NodesInRound(r)
	if err != nil {
		return nil, err
	}
	matches := make([]*models.Match, 0, len(nodes)/2)
	for i := 0; i < len(nodes)-1; i += 2 {
		var completed bool
		match := models.Match{P1: &bracket.Node{}, P2: &bracket.Node{}}

		if p1, err := b.Search(nodes[i].Position); err == nil {
			if payload, err := h.assertPayload(p1); err == nil {
				if payload.Completed {
					completed = true
				}
			}
			match.P1 = p1
		}

		if p2, err := b.Search(nodes[i+1].Position); err == nil {
			if payload, err := h.assertPayload(p2); err == nil {
				if payload.Completed {
					completed = true
				}
			}
			match.P2 = p2
		}

		if completed {
			continue
		}
		matches = append(matches, &match)
	}
	return matches, nil
}

func (h *StartHandler) InsertPayload(bt *bracket.BracketTree, s int, a models.Attendee, result int, completed bool) (*models.AttendeeWithResult, error) {
	node, err := bt.Search(s)
	if err != nil {
		return nil, err
	}
	attendee := models.AttendeeWithResult{Attendee: a, Result: result, Completed: completed}
	node.Payload = attendee
	return &attendee, nil
}

func (h *StartHandler) start(tournamentId []uint8, bracket *bracket.BracketTree, callback func(match models.Match)) error {
	now := time.Now().Unix()
	_, err := h.db.Exec("UPDATE tournaments SET starting_at = IFNULL(starting_at, ?) WHERE id = ?", now, tournamentId)
	if err != nil {
		return err
	}

	mhm := models.NewMatchHistoryModel(h.db)
	currentTournament, err := mhm.CurrentTournamentHistory(tournamentId)
	if err != nil {
		return err
	}

	attendeeWithStatus := make([]models.AttendeeWithResult, 0, len(currentTournament))
	for _, c := range currentTournament {
		// insert current position
		attendee, err := h.InsertPayload(bracket, int(c.Attendee.CurrentSeat.Int64), c.Attendee, 0, false)
		if err != nil {
			return err
		}
		attendeeWithStatus = append(attendeeWithStatus, *attendee)

		// insert previous positions to the bracket if exists
		for _, history := range c.Histories {
			attendee, err := h.InsertPayload(bracket, int(history.Seat.Int64), c.Attendee, history.Result, true)
			if err != nil {
				return err
			}
			attendeeWithStatus = append(attendeeWithStatus, *attendee)
		}
	}

	r := 1
	matches := []*models.Match{}
	for {
		matchesInRound, err := h.generateMatches(bracket, r)
		if err != nil {
			return err
		}
		matches = append(matches, matchesInRound...)

		if r == len(bracket.SeatRoundPos) {
			break
		}

		r++
	}

	go h.MatchQueue.Start(string(tournamentId), bracket, matches, h.ctx, func(match models.Match) {
		callback(match)
	})
	return nil
}

func (h *StartHandler) buildEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, m models.Match) {
	var p1, p2 models.AttendeeWithResult
	pairs := make([]models.AttendeeWithResult, 0, 2)

	if m.P1 != nil {
		if payload, ok := m.P1.Payload.(models.AttendeeWithResult); ok {
			p1 = payload
			pairs = append(pairs, payload)
		}
	} else {
		p1.Player.Name = "N/A"
	}

	if m.P2 != nil {
		if payload, ok := m.P2.Payload.(models.AttendeeWithResult); ok {
			p2 = payload
			pairs = append(pairs, payload)
		}
	} else {
		p2.Player.Name = "N/A"
	}

	buttons := make([]discordgo.MessageComponent, 0, len(pairs))
	for _, payload := range pairs {
		buttons = append(buttons, discordgo.Button{
			Emoji: &discordgo.ComponentEmoji{
				Name: "✅",
			},
			Label:    fmt.Sprintf("%v Wins", payload.Player.Name),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("tournament_processresult_%s_%d_%d", payload.TournamentID, payload.Attendee.Id, payload.CurrentSeat.Int64),
		})
	}
	_, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embed: components.MatchupEmbed(components.MatchupPayload{
			P1: p1.Player.Name,
			P2: p2.Player.Name,
		}),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: buttons,
			},
		},
	})

	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
