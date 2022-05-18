package load

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

var (
	emptyOrNilError       = errors.New("slice of addresses passed was empty or nil")
	onlyNilAddressesError = errors.New("slice of addresses contained no non-nil entries")
)

/*
Because we use a map to remove duplicate addresses, this function makes no guarantees the addresses
returned will be in the same order as they are in the slice passed in. (Map iteration order is not
guaranteed in Go)

This is only a minor headache in tests; for the actual LB behavior it seems preferable to always
defaulting to the first address in case of a tied number of connections. Rather than add a performance
hit to sort the slice here I figure it's better to leave it as is.
*/
func validateAndRemoveDuplicateAddresses(addresses []*net.TCPAddr) ([]*net.TCPAddr, error) {
	if len(addresses) == 0 {
		return nil, emptyOrNilError
	}

	uniqueAddresses := make(map[string]*net.TCPAddr)

	for _, address := range addresses {
		if address == nil {
			continue
		}
		mapKey := getMapKeyForAddress(address)
		uniqueAddresses[mapKey] = address
	}

	if len(uniqueAddresses) == 0 {
		return nil, onlyNilAddressesError
	}

	cleaned := make([]*net.TCPAddr, 0, len(uniqueAddresses))

	for _, address := range uniqueAddresses {
		cleaned = append(cleaned, address)
	}

	return cleaned, nil
}

func getMapKeyForAddress(address *net.TCPAddr) string {
	return strings.Join([]string{address.Zone, address.IP.String(), strconv.Itoa(address.Port)}, "-")
}
