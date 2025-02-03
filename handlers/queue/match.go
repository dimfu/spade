package queue

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"

	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/models"
)

type MatchQueue struct {
	conds      map[string]*sync.Cond
	brackets   map[string]*bracket.BracketTree
	matches    map[string][]*models.Match
	queue      map[string][]*models.Match
	running    map[string]context.CancelFunc
	wg         map[string]*sync.WaitGroup
	matchCount map[string]int
	mutex      sync.Mutex
}

type MatchResult struct {
	Winner     *models.AttendeeWithResult
	Loser      *models.AttendeeWithResult
	WinnerTo   *int
	MatchCount int
}

var (
	instance *MatchQueue
	once     sync.Once
)

func GetMatchQueue() *MatchQueue {
	once.Do(func() {
		instance = &MatchQueue{
			conds:      make(map[string]*sync.Cond),
			brackets:   make(map[string]*bracket.BracketTree),
			matches:    make(map[string][]*models.Match),
			queue:      make(map[string][]*models.Match),
			matchCount: map[string]int{},
			wg:         map[string]*sync.WaitGroup{},
			running:    map[string]context.CancelFunc{},
		}
	})
	return instance
}

func (q *MatchQueue) cleanup(tournamentID string) {
	delete(q.queue, tournamentID)
	delete(q.wg, tournamentID)
	delete(q.brackets, tournamentID)
}

func (q *MatchQueue) ClearQueue(tournamentID string) error {
	if cancelFunc, exists := q.running[tournamentID]; exists {
		cancelFunc()
	}
	if cond, exists := q.conds[tournamentID]; exists {
		cond.Broadcast()
	}
	q.cleanup(tournamentID)
	return nil
}

func (q *MatchQueue) Move(tournamentId string, a models.AttendeeWithResult, to int) error {
	bracket, ok := q.brackets[tournamentId]
	if !ok {
		return errors.New("could not find tournament bracket")
	}

	node, err := bracket.Search(to)
	if err != nil {
		return err
	}

	attendee := models.AttendeeWithResult{Attendee: a.Attendee, Result: 0, Completed: false}
	node.Payload = attendee
	return nil
}

func (q *MatchQueue) Result(tournamentID string, winnerID int) (*MatchResult, error) {
	result := &MatchResult{}

	items, ok := q.queue[tournamentID]
	if !ok || len(items) == 0 {
		return nil, errors.New("cannot find tournament key in queue")
	}

	wg, ok := q.wg[tournamentID]
	if !ok {
		q.mutex.Unlock()
		return nil, errors.New("tournament already completed")
	}
	defer wg.Done()

	popped := items[0]
	match := make(map[int]*bracket.Node)
	currSeats := []int{}

	if popped.P1 != nil {
		match[popped.P1.Position] = popped.P1
		currSeats = append(currSeats, popped.P1.Position)
	}

	if popped.P2 != nil {
		match[popped.P2.Position] = popped.P2
		currSeats = append(currSeats, popped.P2.Position)
	}

	if len(currSeats) == 0 {
		return nil, errors.New("both P1 and P2 are nil")
	}

	b, ok := q.brackets[tournamentID]
	if !ok {
		return nil, errors.New("cannot find bracket with this tournament id")
	}

	winnerTo := -1
	for _, match := range b.Matches {
		matchSeats := []int{match.Seats[0], match.Seats[1]}
		sort.Ints(matchSeats)
		if reflect.DeepEqual(currSeats, matchSeats) {
			winnerTo = match.WinnerTo
			break
		}
	}

	for _, p := range match {
		attendee, ok := p.Payload.(models.AttendeeWithResult)
		if !ok {
			return nil, errors.New("payload is not AttendeeWithResult")
		}
		if attendee.Id != winnerID {
			result.Loser = &attendee
		} else {
			attendee.CurrentSeat.Int64 = int64(winnerTo)
			result.WinnerTo = &winnerTo
			result.Winner = &attendee
			if err := q.Move(tournamentID, *result.Winner, *result.WinnerTo); err != nil {
				return nil, err
			}
		}
	}

	result.MatchCount = q.matchCount[tournamentID]

	if winnerTo == q.brackets[tournamentID].Root.Position {
		return result, base.ERR_FOUND_TOURNAMENT_WINNER
	}

	q.queue[tournamentID] = q.queue[tournamentID][1:]
	q.conds[tournamentID].Signal()
	return result, nil
}

func (q *MatchQueue) Start(
	tournamentID string, bracket *bracket.BracketTree,
	matches []*models.Match, ctx context.Context,
	post func(models.Match, int),
) error {
	q.mutex.Lock()

	if _, running := q.running[tournamentID]; running {
		q.ClearQueue(tournamentID)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	q.running[tournamentID] = cancel

	q.conds[tournamentID] = sync.NewCond(&sync.Mutex{})

	var wg sync.WaitGroup
	q.wg[tournamentID] = &wg
	q.matches[tournamentID] = matches
	q.brackets[tournamentID] = bracket
	q.wg[tournamentID].Add(len(q.matches[tournamentID]))
	q.mutex.Unlock()

	go func(ctx context.Context) {
		q.matchCount[tournamentID] = 1
		for _, currMatch := range q.matches[tournamentID] {
			q.conds[tournamentID].L.Lock()
			if len(q.queue[tournamentID]) == 1 {
				q.conds[tournamentID].Wait()
			}

			select {
			case <-ctx.Done(): // should early return if the tournament is restarting
				return
			default:
			}

			match := &models.Match{}
			if p1, err := q.brackets[tournamentID].Search(currMatch.P1.Position); err == nil {
				match.P1 = p1
			}
			if p2, err := q.brackets[tournamentID].Search(currMatch.P2.Position); err == nil {
				match.P2 = p2
			}

			q.queue[tournamentID] = append(q.queue[tournamentID], match)
			post(*match, q.matchCount[tournamentID])
			q.matchCount[tournamentID]++
			q.conds[tournamentID].L.Unlock()
		}

		q.wg[tournamentID].Wait()
		q.cleanup(tournamentID)
		delete(q.conds, tournamentID)
	}(cancelCtx)

	return nil
}
