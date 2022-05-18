package loadbalance

import (
	"net"
	"testing"

	"github.com/KiaFarhang/tcp-load-balancer/internal/assert"
)

const (
	port int    = 5555
	zone string = ""
)

var ip net.IP = net.IPv4(10, 255, 255, 255)

// TODO: update with changed func name
func TestCleanAddresses(t *testing.T) {
	t.Run("Returns an error if the slice of addresses passed is empty", func(t *testing.T) {
		addresses := make([]*net.TCPAddr, 0)
		_, err := cleanAddresses(addresses)
		assert.Equal(t, err.Error(), emptyOrNilAddressesMessage)
	})
	t.Run("Returns an error if the slice of addresses passed is nil", func(t *testing.T) {
		_, err := cleanAddresses(nil)
		assert.Equal(t, err.Error(), emptyOrNilAddressesMessage)
	})
	t.Run("Returns an error if the slice only contains nil addresses", func(t *testing.T) {
		addresses := []*net.TCPAddr{nil}
		_, err := cleanAddresses(addresses)
		assert.Equal(t, err.Error(), onlyNilAddressesMessage)
	})
	t.Run("Removes duplicate addresses from the returned slice", func(t *testing.T) {
		a := &net.TCPAddr{IP: ip, Port: port, Zone: zone}
		b := &net.TCPAddr{IP: ip, Port: port, Zone: zone}

		addresses := []*net.TCPAddr{a, b}
		cleaned, err := cleanAddresses(addresses)
		assert.NoError(t, err)
		assert.Equal(t, len(cleaned), 1)
	})
}

func TestAreTCPAddressesEqual(t *testing.T) {
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
