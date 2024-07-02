package types

import (
	"bhasherbel/peeruptime/utils"
	"time"
)

type Check struct {
	Time       time.Time
	PingDelay  float64 // ms
	PongDelay  float64 // ms
	LocalDelay float64 // ms
	Success    bool
}

type Checks struct {
	Entries []*Check
	Size    int
	Score   float64
	Average float64
}

var scoreCnt = 0

func (cs *Checks) Append(c Check, factor float64) {
	cs.Entries = append(cs.Entries, &c)
	cs.Average = (cs.Average*float64(cs.Size) + utils.BoolToFloat(c.Success)) / float64(cs.Size+1)

	reverseFactor := 1. / factor
	cs.Score = cs.Average*(1.-reverseFactor) + utils.BoolToFloat(c.Success)*reverseFactor
	cs.Size++
}

func (cs Checks) AmortizedScore() float64 {
	score := float64(scoreCnt)
	scoreCnt++
	//score += float64(config.VariableScoreFactor) * math.Abs(cs.Score-0.5)
	//score += float64(config.VariableScoreFactor) * rand.NormFloat64()
	return score
}

func (cs Checks) Last() *bool {
	if cs.Size == 0 {
		return nil
	}
	return &cs.Entries[cs.Size-1].Success
}
