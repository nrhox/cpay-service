package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
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

	ts := max(g.getTimestampMillis(), g.lastStamp)
	g.lastStamp = ts

	timeDelta := (ts - g.epoch) % 100000000

	randomCrypto, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return ""
	}
	suffixRandom := randomCrypto.Int64()

	return fmt.Sprintf("%08d%04d", timeDelta, suffixRandom)
}
