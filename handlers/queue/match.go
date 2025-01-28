package queue

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/models"
)

type MatchQueue struct {
	cond     *sync.Cond
	brackets map[string]*bracket.BracketTree
	matches  map[string][]*models.Match
	queue    map[string][]*models.Match
	running  map[string]context.CancelFunc
	mutex    sync.Mutex
}

var (
	instance *MatchQueue
	once     sync.Once
)

func GetMatchQueue() *MatchQueue {
	once.Do(func() {
		instance = &MatchQueue{
			cond:     sync.NewCond(&sync.Mutex{}),
			brackets: make(map[string]*bracket.BracketTree),
			matches:  make(map[string][]*models.Match),
			queue:    make(map[string][]*models.Match),
			running:  map[string]context.CancelFunc{},
		}
	})
	return instance
}

func (q *MatchQueue) WaitWithContext(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		q.cond.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *MatchQueue) ClearQueue(tournamentID string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	_, ok := q.queue[tournamentID]
	if !ok {
		return errors.New("cannot find tournament key in queue")
	}

	if cancelFunc, running := q.running[tournamentID]; running {
		cancelFunc()
		delete(q.running, tournamentID)
	}

	q.queue[tournamentID] = []*models.Match{}
	return nil
}

func (q *MatchQueue) Remove(tournamentID string, winnerID int) error {
	now := time.Now().Unix()
	db := database.GetDB()
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	items, ok := q.queue[tournamentID]
	if !ok && len(items) == 0 {
		return errors.New("cannot find tournament key in queue")
	}

	popped := items[0]
	match := []*bracket.Node{popped.P1, popped.P2}
	for _, p := range match {
		attendee, ok := p.Payload.(models.AttendeeWithResult)
		if !ok {
			return errors.New("payload is not AttendeeWithResult")
		}
		if attendee.Id != winnerID {
			q := `INSERT INTO match_histories (attendee_id, result, seat, created_at) VALUES (?, ?, ?, ?)`
			_, err := db.Exec(q, attendee.Id, 0, attendee.CurrentSeat.Int64, now)
			if err != nil {
				return err
			}
			break
		}
	}

	q.queue[tournamentID] = q.queue[tournamentID][1:]
	q.cond.Signal()
	return nil
}

func (q *MatchQueue) Start(
	tournamentID string, bracket *bracket.BracketTree,
	matches []*models.Match, ctx context.Context,
	post func(models.Match),
) error {
	q.mutex.Lock()

	if cancelFunc, running := q.running[tournamentID]; running {
		cancelFunc()
		q.ClearQueue(tournamentID)
	}

	newCtx, cancel := context.WithCancel(ctx)
	q.running[tournamentID] = cancel

	q.matches[tournamentID] = matches
	q.brackets[tournamentID] = bracket

	q.mutex.Unlock()

	go func() {
		for _, match := range q.matches[tournamentID] {
			q.cond.L.Lock()
			if len(q.queue[tournamentID]) == 1 {
				if err := q.WaitWithContext(ctx); err != nil {
					q.cond.L.Unlock()
					if newCtx.Err() == context.Canceled {
						log.Println(err)
						return
					}
					panic(err)
				}
			}
			q.queue[tournamentID] = append(q.queue[tournamentID], match)
			q.cond.L.Unlock()
			post(*match)
		}
		q.mutex.Lock()
		delete(q.running, tournamentID)
		q.mutex.Unlock()
	}()

	return nil
}
