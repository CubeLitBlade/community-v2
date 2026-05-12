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
