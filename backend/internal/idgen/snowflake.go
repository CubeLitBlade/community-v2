// Package idgen provides utilities for generating unique Snowflake IDs.
package idgen

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var (
	// ErrWorkerIDOutOfRange indicates that the provided worker ID exceeds the
	// allowed limit.
	ErrWorkerIDOutOfRange = errors.New("worker id out of range")

	// ErrClockMovedBack indicates that the system clock moved backwards.
	ErrClockMovedBack = errors.New("clock moved back")
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

// Snowflake represents a Snowflake ID generator instance.
type Snowflake struct {
	mu            sync.Mutex
	workerID      int64
	sequence      int64
	lastTimestamp int64
}

// NewSnowflake creates a new Snowflake ID generator with the given worker ID.
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
		mu:            sync.Mutex{},
		workerID:      workerID,
		sequence:      0,
		lastTimestamp: -1,
	}, nil
}

// NextID generates and returns a unique Snowflake ID.
func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentTimestampMilli := time.Now().UnixMilli()

	if currentTimestampMilli < s.lastTimestamp {
		return 0, fmt.Errorf(
			"%w: current timestamp %d is before last timestamp %d",
			ErrClockMovedBack,
			currentTimestampMilli,
			s.lastTimestamp,
		)
	}

	var seq int64

	if currentTimestampMilli == s.lastTimestamp {
		seq = (s.sequence + 1) & maxSequence

		if seq == 0 {
			currentTimestampMilli = waitNextMillis(s.lastTimestamp)
		}

		s.sequence = seq
	} else {
		seq = 0
		s.sequence = 0
	}

	s.lastTimestamp = currentTimestampMilli

	id := ((currentTimestampMilli - epochMillis) << timestampShift) |
		(s.workerID << workerIDShift) |
		seq

	return id, nil
}

func waitNextMillis(lastTimestamp int64) int64 {
	timestampMilli := time.Now().UnixMilli()

	for timestampMilli <= lastTimestamp {
		runtime.Gosched()

		timestampMilli = time.Now().UnixMilli()
	}

	return timestampMilli
}
