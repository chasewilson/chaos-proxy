package chaos

import (
	"math/rand"
	"time"
)

type Curse struct {
	DropConnections bool
	StartDelay      time.Duration
}

type Ritual struct {
	DropRate  float64
	LatencyMs int
}

func NewCurse(ritual Ritual) Curse {
	curse := Curse{}

	if ritual.DropRate > 0 && rand.Float64() < ritual.DropRate {
		curse.DropConnections = true
	}

	if ritual.LatencyMs > 0 {
		curse.StartDelay = time.Duration(ritual.LatencyMs) * time.Millisecond
	}

	return curse
}
