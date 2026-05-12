package account

import (
	"net/netip"
	"time"
)

type Audit struct {
	createdAt time.Time
	updatedAt time.Time
	lastLogin *LoginAudit
}

type LoginAudit struct {
	at time.Time
	ip netip.Addr
}

func NewAudit(now time.Time) Audit {
	return Audit{
		createdAt: now,
		updatedAt: now,
		lastLogin: nil,
	}
}
