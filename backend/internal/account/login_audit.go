package account

import (
	"net/netip"
	"time"
)

type LoginAudit struct {
	at time.Time
	ip netip.Addr
}

func NewLoginAudit(at time.Time, ip netip.Addr) *LoginAudit {
	return &LoginAudit{
		at: at,
		ip: ip,
	}
}

func (a *LoginAudit) Update(at time.Time, addr netip.Addr) {
	a.at = at
	a.ip = addr
}
