package account

import (
	"net/netip"
	"time"
)

type LastLogin struct {
	Time   time.Time
	IPAddr netip.Addr
}

func NewLastLogin(now time.Time, ipAddr netip.Addr) *LastLogin {
	return &LastLogin{
		Time:   now,
		IPAddr: ipAddr,
	}
}

func (a *LastLogin) Update(now time.Time, ipAddr netip.Addr) {
	a.Time = now
	a.IPAddr = ipAddr
}
