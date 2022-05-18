package loadbalance

import (
	"net"
)

// Assumes a and b are not nil
func areTCPAddressesEqual(a, b *net.TCPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port && a.Zone == b.Zone
}
