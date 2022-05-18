package loadbalance

import (
	"net"
	"testing"

	"github.com/KiaFarhang/tcp-load-balancer/internal/assert"
)

func TestIsTCPAdressEqual(t *testing.T) {
	ip := net.IPv4(10, 255, 255, 255)
	port := 5555
	zone := ""
	t.Run("Returns true if all fields on both addresses are equal", func(t *testing.T) {
		a := &net.TCPAddr{IP: ip, Port: port, Zone: zone}
		b := &net.TCPAddr{IP: ip, Port: port, Zone: zone}

		result := areTCPAddressesEqual(a, b)
		assert.Equal(t, result, true)
	})
	t.Run("Returns false if any fields on the addresses are different", func(t *testing.T) {
		a := &net.TCPAddr{IP: ip, Port: port, Zone: zone}
		b := &net.TCPAddr{IP: ip, Port: 6666, Zone: zone}

		result := areTCPAddressesEqual(a, b)
		assert.Equal(t, result, false)
	})
}
