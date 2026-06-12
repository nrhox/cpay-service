package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"sync"
	"time"
)

type ReferenceCode struct {
	mu        sync.Mutex
	lastStamp int64
}

func NewReferenceCode() *ReferenceCode {
	return &ReferenceCode{}
}

// Next generates a unique, 14-character transaction reference string.
// It requires a 2-character prefix (txType) denoting the operation type (e.g., "TF", "TP").
//
// The output follows the format: [txType][Y][DDD][HH][6-digit random]
// Example: "TF615722084921"
func (g *ReferenceCode) Next(txType string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()

	// Generate cryptographically secure random number to prevent enumeration attacks
	randomCrypto, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	// Pre-allocate fixed 14 bytes on stack
	buf := make([]byte, 0, 14)

	// transaction type prefix (2 bytes)
	buf = append(buf, txType...)

	// last digit of current year (1 byte)
	buf = strconv.AppendInt(buf, int64(now.Year()%10), 10)

	// ordinal day of the year / Julian day (3 bytes with padding)
	jd := now.YearDay()
	if jd < 100 {
		buf = append(buf, '0')
		if jd < 10 {
			buf = append(buf, '0')
		}
	}
	buf = strconv.AppendInt(buf, int64(jd), 10)

	// current hour (2 bytes with padding)
	hr := now.Hour()
	if hr < 10 {
		buf = append(buf, '0')
	}
	buf = strconv.AppendInt(buf, int64(hr), 10)

	// random sequence (6 bytes with padding)
	seq := randomCrypto.Int64()
	startIdx := len(buf)
	buf = strconv.AppendInt(buf, seq, 10)
	currLen := len(buf) - startIdx

	if currLen < 6 {
		padding := 6 - currLen
		buf = append(buf, make([]byte, padding)...)
		copy(buf[startIdx+padding:], buf[startIdx:])
		for i := 0; i < padding; i++ {
			buf[startIdx+i] = '0'
		}
	}

	return string(buf), nil
}
