package loadbalance

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	emptyOrNilAddressesMessage string = "slice of addresses passed was empty or nil"
	onlyNilAddressesMessage    string = "slice of addresses contained no non-nil entries"
)

// TODO: this function name is bad
func cleanAddresses(addresses []*net.TCPAddr) ([]*net.TCPAddr, error) {
	cleaned := make([]*net.TCPAddr, 0, len(addresses))
	if len(addresses) == 0 {
		return cleaned, errors.New(emptyOrNilAddressesMessage)
	}

	uniqueAddresses := make(map[string]*net.TCPAddr)

	for _, address := range addresses {
		if address == nil {
			continue
		}
		mapKey := getMapKeyForAddress(address)
		uniqueAddresses[mapKey] = address
	}

	for _, address := range uniqueAddresses {
		cleaned = append(cleaned, address)
	}

	if len(cleaned) == 0 {
		return []*net.TCPAddr{}, errors.New(onlyNilAddressesMessage)
	}

	return cleaned, nil
}

func getMapKeyForAddress(address *net.TCPAddr) string {
	return strings.Join([]string{address.Zone, address.IP.String(), strconv.Itoa(address.Port)}, "-")
}

// Assumes a and b are not nil
func areTCPAddressesEqual(a, b *net.TCPAddr) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port && a.Zone == b.Zone
}
