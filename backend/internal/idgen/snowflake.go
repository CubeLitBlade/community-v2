package idgen

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var (
	ErrWorkerIDOutOfRange = errors.New("worker id out of range")
	ErrClockMovedBack     = errors.New("clock moved back")
)

const (
	epochMillis = int64(1773187200000) // 2026-03-11T00:00:00Z

	sequenceBits = int64(12)
	workerIDBits = int64(10)

	maxSequence = int64(-1) ^ (int64(-1) << sequenceBits)
	maxWorkerID = int64(-1) ^ (int64(-1) << workerIDBits)

	workerIDShift  = sequenceBits
	timestampShift = sequenceBits + workerIDBits
)

type Snowflake struct {
	mu            sync.Mutex
	workerID      int64
	sequence      int64
	lastTimestamp int64
}

func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf(
			"%w: worker id must be between 0 and %d, got %d",
			ErrWorkerIDOutOfRange,
			maxWorkerID,
			workerID,
		)
	}

	return &Snowflake{
		workerID:      workerID,
		lastTimestamp: -1,
	}, nil
}

func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentTimestamp := time.Now().UnixMilli()

	if currentTimestamp < s.lastTimestamp {
		return 0, fmt.Errorf(
			"%w: current timestamp %d is before last timestamp %d",
			ErrClockMovedBack,
			currentTimestamp,
			s.lastTimestamp,
		)
	}

	var seq int64

	if currentTimestamp == s.lastTimestamp {
		seq = (s.sequence + 1) & maxSequence

		if seq == 0 {
			currentTimestamp = waitNextMillis(s.lastTimestamp)
		}

		s.sequence = seq
	} else {
		seq = 0
		s.sequence = 0
	}

	s.lastTimestamp = currentTimestamp

	id := ((currentTimestamp - epochMillis) << timestampShift) |
		(s.workerID << workerIDShift) |
		seq

	return id, nil
}

func waitNextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixMilli()

	for timestamp <= lastTimestamp {
		runtime.Gosched()
		timestamp = time.Now().UnixMilli()
	}

	return timestamp
}
