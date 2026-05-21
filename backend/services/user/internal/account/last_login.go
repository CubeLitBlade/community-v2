package account

import (
	"net/netip"
	"time"
)

// LastLogin represents the timestamp and IP address of a user's last login.
type LastLogin struct {
	Time   time.Time  // Time is the time of the last login.
	IPAddr netip.Addr // IPAddr is the IP address used during the last login.
}

// NewLastLogin creates a new LastLogin instance with the given time and IP address.
func NewLastLogin(now time.Time, ipAddr netip.Addr) *LastLogin {
	return &LastLogin{
		Time:   now,
		IPAddr: ipAddr,
	}
}

// Update sets the last login time and IP address to the provided values.
func (a *LastLogin) Update(now time.Time, ipAddr netip.Addr) {
	a.Time = now
	a.IPAddr = ipAddr
}
