package tools

import (
	"fmt"
	"net/netip"
)

// ListIpsInNetwork List all host ips in given cidr network
// example ips in 192.168.0.0/24
// 192.168.0.0
// 192.168.0.1
// .....
// 192.168.0.255
func ListIpsInNetwork(cidrAddress string) ([]string, error) {
	prefix, err := netip.ParsePrefix(cidrAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid cidr: %s, error %v", cidrAddress, err)
	}
	var maskedPrefix = prefix.Masked()

	var ips = make([]string, 0)
	for addr := maskedPrefix.Addr(); maskedPrefix.Contains(addr); addr = addr.Next() {
		ips = append(ips, addr.String())
	}
	return ips, nil
}
