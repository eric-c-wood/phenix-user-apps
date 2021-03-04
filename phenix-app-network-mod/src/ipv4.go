package main

import (
	"fmt"
	"math"
	"math/bits"
	"net"
	"regexp"
	"strings"
)

const (
	bitsPerByte   int = 8
	ipv4BitLength int = 32
)

var (
	allOnes = uint32(math.Pow(2, 32) - 1)
	ipv4Re  = regexp.MustCompile(`(?:\d{1,3}[.]){3}\d{1,3}`)
)

type ipv4Network struct {
	address         uint32
	cidr            int
	netmask         uint32
	network         uint32
	broadcast       uint32
	lastAddressUsed uint32
}

func newIPv4Network(address string) *ipv4Network {

	addr, refNet, err := net.ParseCIDR(address)

	if err != nil {
		fmt.Printf("Can not create network %s\n", address)
		return nil
	}

	//fmt.Printf("Address:%v\n",addr)
	cidr, _ := refNet.Mask.Size()
	uintAddress, err := addressToUint(addr.String())

	if err != nil {
		return nil
	}

	return &ipv4Network{
		address:         uintAddress,
		cidr:            cidr,
		netmask:         maskFromCIDR(cidr),
		network:         network(cidr, uintAddress),
		broadcast:       broadcast(cidr, uintAddress),
		lastAddressUsed: network(cidr, uintAddress),
	}

}

func (this *ipv4Network) printRange() string {

	return fmt.Sprintf("%s - %s", uintToAddress(this.network), uintToAddress(this.broadcast))
}

func (this *ipv4Network) printShort() string {

	return fmt.Sprintf("%s/%d", uintToAddress(this.address), this.cidr)

}

func (this *ipv4Network) printLong() string {

	return fmt.Sprintf("%s %s", uintToAddress(this.address), uintToAddress(this.netmask))

}

func (this *ipv4Network) wildCardMask() string {
	return uintToAddress(invertAddress(this.netmask))
}

func (this *ipv4Network) getNextAddress(addressesUsed map[uint32]bool) string {

	addressStart := this.lastAddressUsed
	// Add one to the starting address until
	// an address is found that is not currently used
	// Do allow the network or broadcast addresses to
	// be used
	for {
		addressStart += 1

		if addressStart <= this.network {
			return ""
		}

		if addressStart >= this.broadcast {
			return ""
		}

		if _, ok := addressesUsed[addressStart]; !ok {
			this.lastAddressUsed = addressStart
			return uintToAddress(addressStart)
		}

	}

	return ""
}

func (this *ipv4Network) getUsableHostCount() int {

	// The usable addreses will exclude the network and
	// broadcast addresses
	return int((this.broadcast - this.network) - 1)

}

func (this *ipv4Network) containsAddress(address string) (bool, error) {

	uintAddress, err := addressToUint(address)

	if err != nil {
		return false, err
	}

	// The address is contained if the address is between the
	// network and broadcast addresses or equal to the network or
	// broadcast address
	return this.network <= uintAddress && uintAddress <= this.broadcast, nil

}

func (this *ipv4Network) containsUintAddress(uintAddress uint32) bool {

	// The address is contained if the address is between the
	// network and broadcast addresses or equal to the network or
	// broadcast address
	return this.network <= uintAddress && uintAddress <= this.broadcast

}

func (this *ipv4Network) containsNetwork(network string) (bool, error) {

	ipv4Net := newIPv4Network(network)

	if ipv4Net == nil {
		return false, fmt.Errorf("network is invalid")
	}

	// The address is contained if the address is between the
	// network and broadcast addresses or equal to the network or
	// broadcast address
	return this.containsUintAddress(ipv4Net.network) && this.containsUintAddress(ipv4Net.broadcast), nil

}

func broadcast(cidr int, address uint32) uint32 {

	return network(cidr, address) | (allOnes >> cidr)
}

func network(cidr int, address uint32) uint32 {

	mask := maskFromCIDR(cidr)
	return address & mask
}

func maskFromCIDR(cidr int) uint32 {

	return allOnes << (ipv4BitLength - cidr)
}

func cidrFromMask(netmask uint32) int {

	return bits.OnesCount32(netmask)
}

func addressToUint(address string) (uint32, error) {

	uintAddress := make([]uint32, 4)

	octets := net.ParseIP(address)

	if octets == nil {
		return 0, fmt.Errorf("invalid address %v", address)
	}

	offset := 12

	for i := offset; i < offset+4; i++ {

		uintOctet := uint32(octets[i]) << (ipv4BitLength - ((i - offset + 1) * bitsPerByte))
		uintAddress[i-offset] = uintOctet
	}

	return uintAddress[0] | uintAddress[1] | uintAddress[2] | uintAddress[3], nil
}

func uintToAddress(uintAddress uint32) string {

	output := make([]string, 4)

	for i := 0; i < 4; i++ {

		shiftCount := (ipv4BitLength - ((i + 1) * bitsPerByte))
		bitMask := uint32(255 << shiftCount)
		decByte := (uintAddress & bitMask) >> shiftCount
		output[i] = fmt.Sprintf("%d", decByte)

	}

	return strings.Join(output, ".")

}

func invertAddress(address uint32) uint32 {

	return address ^ allOnes
}
