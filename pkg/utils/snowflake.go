package utils

import (
	"fmt"
	"sync"
	"time"
)

type Snowflake struct {
	mu        sync.Mutex
	epoch     int64
	sequence  int64
	lastStamp int64
}

func NewSnowflake(epochMillis int64) *Snowflake {
	return &Snowflake{
		epoch: epochMillis,
	}
}

func (g *Snowflake) getTimestampMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func (g *Snowflake) NextID() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	ts := g.getTimestampMillis()
	if ts == g.lastStamp {
		g.sequence = (g.sequence + 1) % 10000
		if g.sequence == 0 {
			for ts <= g.lastStamp {
				ts = g.getTimestampMillis()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastStamp = ts

	timeDelta := (ts - g.epoch) % 100000000

	return fmt.Sprintf("%08d%04d", timeDelta, g.sequence)
}
