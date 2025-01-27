package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/dimfu/spade/bracket"
	"github.com/dimfu/spade/models"
)

type MatchQueue struct {
	cond     *sync.Cond
	brackets map[string]*bracket.BracketTree
	matches  map[string][]*models.Match
	queue    map[string][]*models.Match
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

func (q *MatchQueue) Remove(tournamentID string) error {
	_, ok := q.queue[tournamentID]
	if !ok {
		return errors.New("cannot find tournament key in queue")
	}
	q.cond.L.Lock()
	q.queue[tournamentID] = q.queue[tournamentID][1:]
	q.cond.L.Unlock()
	q.cond.Signal()
	return nil
}

func (q *MatchQueue) Start(
	tournamentID string, bracket *bracket.BracketTree,
	matches []*models.Match, ctx context.Context,
	post func(models.Match),
) error {
	q.matches[tournamentID] = matches
	q.brackets[tournamentID] = bracket

	for _, match := range q.matches[tournamentID] {
		q.cond.L.Lock()
		if len(q.queue[tournamentID]) == 1 {
			if err := q.WaitWithContext(ctx); err != nil {
				q.cond.L.Unlock()
				panic(err)
			}
		}
		q.queue[tournamentID] = append(q.queue[tournamentID], match)
		q.cond.L.Unlock()
		post(*match)
	}

	return nil
}
