package load

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

/*
Because we use a map to remove duplicate addresses, this function makes no guarantees the addresses
returned will be in the same order as they are in the slice passed in. (Map iteration order is not
guaranteed in Go)

This is only a minor headache in tests; for the actual LB behavior it seems preferable to always
defaulting to the first address in case of a tied number of connections. Rather than add a performance
hit to sort the slice here I figure it's better to leave it as is.
*/
func validateAndRemoveDuplicateAddresses(addresses []*net.TCPAddr) ([]*net.TCPAddr, error) {
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
