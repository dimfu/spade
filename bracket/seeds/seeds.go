package seeds

import (
	"errors"
	"math/rand"
	"time"
)

type Stragies = int

const (
	RANDOM Stragies = iota
	BEST_AGAINST_WORST
	SIMILAR_SKILL
)

func NewSeeds(payload []interface{}, strat Stragies) ([]interface{}, error) {
	if len(payload) <= 0 {
		return nil, errors.New("payload cannot be 0 or less")
	}

	switch strat {
	case RANDOM:
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := len(payload) - 1; i >= 0; i-- {
			j := r.Intn(i + 1)
			temp := payload[i]
			payload[i] = payload[j]
			payload[j] = temp
		}
		break
	case BEST_AGAINST_WORST:
		var temp interface{}
		for i := 1; i < len(payload)/2; i += 2 {
			temp = payload[i]
			payload[i] = payload[len(payload)-i]
			payload[len(payload)-i] = temp
		}
		break
	case SIMILAR_SKILL:
		// based on the payload position, first index is best & last index is worst
		break
	default:
		return nil, errors.New("seeding strategies not found")
	}

	return payload, nil
}
