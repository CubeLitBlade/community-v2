package account

import (
	"net/netip"
	"time"
)

// Audit tracks account creation, updates, and last login.
type Audit struct {
	createdAt time.Time
	updatedAt time.Time
	lastLogin *LoginAudit
}

// LoginAudit records a single login time and IP.
type LoginAudit struct {
	at time.Time
	ip netip.Addr
}

// NewAudit creates a new Audit with the given time.
func NewAudit(now time.Time) Audit {
	return Audit{
		createdAt: now,
		updatedAt: now,
		lastLogin: nil,
	}
}
