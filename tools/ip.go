package tools

import (
	"fmt"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"log"
	"net/netip"
)

func SubnetIps(subnet string) ([]string, error) {
	var addressString = ipaddr.NewIPAddressString(subnet)
	if !addressString.IsValid() {
		return nil, fmt.Errorf("subnet %s is not valid", subnet)
	}

	var ips = make([]string, 0)
	for i := addressString.GetAddress().Iterator(); i.HasNext(); {
		ips = append(ips, i.Next().ToCanonicalWildcardString())
	}
	return ips, nil
}

func Subnets(cidrAddress string) {
	prefix, err := netip.ParsePrefix(cidrAddress)
	if err != nil {
		log.Fatalf("invalid cidr: %s, error %v", cidrAddress, err)
	}
	var maskedPrefix = prefix.Masked()

	for addr := maskedPrefix.Addr(); maskedPrefix.Contains(addr); addr = addr.Next() {
		fmt.Println(addr)
	}
}

func listSubnets(cidrAddress string, newPrefixLen int) {
	address := ipaddr.NewIPAddressString(cidrAddress)
	if !address.IsValid() {
		log.Fatalf("%s is not valid", cidrAddress)
	}
	subnet := address.GetAddress()
	fmt.Println(subnet)
	fmt.Println(subnet.ToCanonicalWildcardString())
	subnetPrefixLen := address.GetNetworkPrefixLen()
	if subnetPrefixLen.Len() < newPrefixLen {
		iterator := subnet.SetPrefixLen(newPrefixLen).PrefixIterator()
		for iterator.HasNext() {
			fmt.Println(iterator.Next())
		}
		return
	}

	fmt.Println(subnet.ToCanonicalWildcardString())
}
