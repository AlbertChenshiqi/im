package snowflake

import (
	"sync"
	"time"
)

const (
	epoch          = int64(1704067200000) // 2024-01-01
	workerBits     = 10
	sequenceBits   = 12
	maxWorker      = -1 ^ (-1 << workerBits)
	maxSequence    = -1 ^ (-1 << sequenceBits)
	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Generator struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastStamp int64
}

func New(workerID int64) *Generator {
	if workerID < 0 || workerID > maxWorker {
		workerID = 1
	}
	return &Generator{workerID: workerID}
}

func (g *Generator) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now().UnixMilli()
	if now == g.lastStamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}
	g.lastStamp = now
	return ((now - epoch) << timestampShift) | (g.workerID << workerShift) | g.sequence
}
