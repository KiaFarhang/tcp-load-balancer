package load

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	port int    = 5555
	zone string = ""
)

var ip net.IP = net.IPv4(10, 255, 255, 255)

func TestValidateAndRemoveDuplicateAddresses(t *testing.T) {
	t.Run("Returns an error if the slice of addresses passed is empty", func(t *testing.T) {
		addresses := make([]*net.TCPAddr, 0)
		_, err := validateAndRemoveDuplicateAddresses(addresses)
		assert.ErrorIs(t, err, errEmptyOrNilSlice)
	})
	t.Run("Returns an error if the slice of addresses passed is nil", func(t *testing.T) {
		_, err := validateAndRemoveDuplicateAddresses(nil)
		assert.ErrorIs(t, err, errEmptyOrNilSlice)
	})
	t.Run("Returns an error if the slice only contains nil addresses", func(t *testing.T) {
		addresses := []*net.TCPAddr{nil}
		_, err := validateAndRemoveDuplicateAddresses(addresses)
		assert.ErrorIs(t, err, errOnlyNilAddresses)
	})
	t.Run("Removes duplicate addresses from the returned slice", func(t *testing.T) {
		a := &net.TCPAddr{IP: ip, Port: port, Zone: zone}
		b := &net.TCPAddr{IP: ip, Port: port, Zone: zone}

		addresses := []*net.TCPAddr{a, b}
		cleaned, err := validateAndRemoveDuplicateAddresses(addresses)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(cleaned))
	})
	t.Run("Removes nil addresses from the returned slice", func(t *testing.T) {
		a := &net.TCPAddr{IP: ip, Port: port, Zone: zone}

		addresses := []*net.TCPAddr{a, nil}
		cleaned, err := validateAndRemoveDuplicateAddresses(addresses)
		expected := []*net.TCPAddr{a}

		assert.NoError(t, err)
		assert.Equal(t, expected, cleaned)
	})
}
