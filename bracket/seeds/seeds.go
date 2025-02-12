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

func NewSeeds(payload []interface{}, strat Stragies, slot int) ([]interface{}, error) {
	if len(payload) <= 0 {
		return nil, errors.New("payload cannot be 0 or less")
	}

	switch strat {
	// ? should empty seat be randomized or be grouped together if possible?
	case RANDOM:
		for len(payload) < slot {
			payload = append(payload, nil)
		}
		for i := len(payload) - 1; i >= 0; i-- {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			j := r.Intn(i + 1)
			temp := payload[i]
			payload[i] = payload[j]
			payload[j] = temp
		}
		break
	case BEST_AGAINST_WORST:
		var temp interface{}
		for len(payload) < slot {
			payload = append(payload, nil)
		}

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
