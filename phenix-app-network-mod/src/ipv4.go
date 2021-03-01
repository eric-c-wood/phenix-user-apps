package main

import (	
	"fmt"	
	"net"		
	"math"
	"math/bits"
	"strings"
	"strconv"
)

const (
	bitsPerByte int = 8
	ipV4BitLength int = 32	
)

var (	
	allOnes = uint32(math.Pow(2,32)-1)
)


type ipV4Network struct {
	address uint32
	cidr int
	netmask uint32
	network uint32
	broadcast uint32
	lastAddressUsed uint32
}


func newIPv4Network(address string) *ipV4Network {

	addr,refNet,err := net.ParseCIDR(address)
	
	if err != nil {
		fmt.Printf("Can not create network %s\n",address)
		return nil
	}
	
	//fmt.Printf("Address:%v\n",addr)
	cidr,_ := refNet.Mask.Size()
	binAddress := addressToBinary(addr.String())
	
	return &ipV4Network{
		address:binAddress,
		cidr:cidr,
		netmask:maskFromCIDR(cidr),
		network:network(cidr,binAddress),
		broadcast:broadcast(cidr,binAddress),
		lastAddressUsed:network(cidr,binAddress),
	
	}

}

func (this *ipV4Network) printRange() string {
	
	return fmt.Sprintf("%s - %s",binaryToAddress(this.network),binaryToAddress(this.broadcast))	
}

func (this *ipV4Network) printShort() string {

	return fmt.Sprintf("%s/%d",binaryToAddress(this.address),this.cidr)
	
}

func (this *ipV4Network) wildCardMask() string {
	return binaryToAddress(invertAddress(this.netmask))	
}

func (this *ipV4Network) getNextAddress(addressesUsed map[uint32]bool) string {
	

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

		if _,ok := addressesUsed[addressStart]; !ok {			
			return binaryToAddress(addressStart)
		}
		
	}
			
	return ""
}

func (this *ipV4Network) getUsableHostCount() int {
	
	// The usable addreses will exclude the network and
	// broadcast addresses
	return int((this.broadcast - this.network) - 1)


}

func broadcast(cidr int, address uint32) uint32 {
	
	return network(cidr,address) | (allOnes >> cidr)	
}

func network(cidr int, address uint32) uint32 {

	mask := maskFromCIDR(cidr)
	return address & mask	
}

func maskFromCIDR(cidr int) uint32 {

	return allOnes << (ipV4BitLength - cidr)	
}

func cidrFromMask(netmask uint32) int {
	
	return bits.OnesCount32(netmask)	
}

func addressToBinary(address string) uint32 {

	var binAddress []uint32
		
	octets := strings.Split(address,".")
	
	for i:=0; i<len(octets);i++ {
	
		intOctet,_ := strconv.Atoi(octets[i])
		binOctet := uint32(intOctet) << (ipV4BitLength - ((i+1)*bitsPerByte))
		binAddress = append(binAddress,binOctet)
	}
	
	return binAddress[0] | binAddress[1] | binAddress[2] | binAddress[3]
}

func binaryToAddress (binaryAddress uint32) string {

	var output []string
	
		
	for i:=0; i<4; i++ {
	
		shiftCount := (ipV4BitLength - ((i+1)*bitsPerByte))
		bitMask := uint32(255 << shiftCount)
		decByte := (binaryAddress & bitMask) >> shiftCount
		output = append(output,fmt.Sprintf("%d",decByte))
	
	}
	
	return strings.Join(output,".")

}

func invertAddress (address uint32) uint32 {

	
	return address ^ allOnes
}

