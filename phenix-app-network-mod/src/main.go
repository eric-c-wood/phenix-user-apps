package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"	
	"os"		
	"strings"
	"strconv"
	"net"
	"log"
	"regexp"
	"github.com/mitchellh/mapstructure"
)

var (
	vlansUsed 	   map[string]bool	
	addressesUsed  map[uint32]bool
	ifaceRe = regexp.MustCompile(`(?i)([a-z]+)(\d+)`)	
)

type NetworkMod struct {
	Action  	string  `json:"action" mapstructure:"action"`
	Network 	string  `json:"network" mapstructure:"network"`
	VLAN		int 	`json:"vlan" mapstructure:"vlan"`
	Alias		string 	`json:"alias" mapstructure:"alias"`
	Prefix      string  `json:"prefix" mapstructure:"prefix"`
	Type		string	`json:"type" mapstructure:"type"`
	Gateway		string	`json:"gateway" mapstructure:"gateway"`
	Hosts  	   []string `json:"hosts" mapstructure:"hosts"`
}

type NetworkMods struct {
	Mods  []*NetworkMod  `json:"modifications" mapstructure:"modifications"`
}

var logger *log.Logger

func main() {

	out := os.Stderr

	if env, ok := os.LookupEnv("PHENIX_LOG_FILE"); ok {
		var err error

		out, err = os.OpenFile(env, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("unable to open phenix log file for writing")
		}

		defer out.Close()
	}

	logger = log.New(out, " network-mod ", log.Ldate|log.Ltime|log.Lmsgprefix)


	if len(os.Args) != 2 {
		logger.Fatal("incorrect amount of args provided")
	}

	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logger.Fatal("unable to read JSON from STDIN")
	}

	stage := os.Args[1]
	

	if stage != "configure" {
		fmt.Print(string(body))
		return
	}

	var exp Experiment

	err = json.Unmarshal(body,&exp)
	if err != nil {
		logger.Fatalf("decoding experiment: %v", err)
	}

	switch stage {
	case "configure":
		if err := configure(&exp); err != nil {
			logger.Fatalf("failed to execute configure stage: %v", err)
		}	
	}

	body, err = json.Marshal(exp)
	if err != nil {
		logger.Fatalf("unable to convert experiment to JSON")
	}

	//logger.Printf("Output:%v",string(body))
	fmt.Print(string(body))
}

func configure(exp *Experiment) error {	

	// Get any network modifications that are defined
	networkModifications := getNetworkMods(exp)

	// If no network modifications were found,
	// just return
	if networkModifications == nil {
		logger.Print("No network modifications found")
		return nil
	}

	initUsedTables(exp.Spec.Topology)

	// Apply the modifications one at a time
	for _,mod := range networkModifications {

		// Add any defaults
		addDefaults(mod)

		// Add any defined VLAN aliases
		if mod.VLAN != 0 {
			addVLANAlias(mod,exp.Spec.VLANs)
		}

		// Make sure the gateway is in the same subnet
		if len(mod.Gateway) > 0 {
			if check,err := isInSubnet(mod.Network,mod.Gateway); !check {
				if err != nil {
					return err
				}
				return fmt.Errorf("%s is not in %s",mod.Gateway,mod.Network)
				
			}

		}


		if err := applyModification(mod,exp.Spec.Topology) ; err != nil {
			return err
		}
	}

	return nil
}

func initUsedTables(topology *TopologySpec) {

	nodes := topology.Nodes
	vlansUsed = make(map[string]bool)
	addressesUsed = make(map[uint32]bool)

	for _,node := range nodes {

		for _,iface := range node.Network.Interfaces {			
			if _,ok := vlansUsed[iface.VLAN]; !ok {
				vlansUsed[iface.VLAN] = true
			}

			decAddress := addressToBinary(iface.Address)

			if _,ok := addressesUsed[decAddress]; !ok {
				addressesUsed[decAddress] = true
			}
		}
	}
}

func addDefaults(mod *NetworkMod) {

	// Only apply defaults for
	// add actions
	if mod.Action != "add" {
		return
	}

	// If no network is specified, use
	// a large private classB
	if len(mod.Network) == 0 {
		mod.Network = "172.16.0.0/16"
	}

	// If no VLAN alias is provied, try
	// to find an available alias
	if len(mod.Alias) == 0 {
		mod.Alias = findAvailableAlias()
	}

	// TODO - Perhaps try to infer interface
	// prefix from existing interfaces
	if len(mod.Prefix) == 0 {
		mod.Prefix = "eth"
	}

	if len(mod.Type) == 0 {
		mod.Type = "ethernet"
	}


}

func addVLANAlias(mod *NetworkMod,vlans *VLANSpec) {

	// Add the alias if it does not already exist
	if _,ok := vlans.Aliases[mod.Alias]; !ok {	
			vlans.Aliases[mod.Alias] = mod.VLAN
		
	}

}

func findAvailableAlias() string {

	aliasPrefix := "network"	
	counter := 1

	// Loop until a suitable alias is 
	// found
	for {
		testAlias := fmt.Sprintf("%s%d",aliasPrefix,counter)

		// Limit the infinite loop to 10000
		// iterations
		if counter >= 10000 {
			return ""
		}

		// Alias already exists, try the next name
		if _,ok := vlansUsed[testAlias]; ok {
			counter += 1
			continue
		}

		return testAlias

	}

	return ""
}

func findAvailableName(prefix string, ifaces []Interface) string {

	ifaceMap := make(map[string]bool)
	lastIndex := 0
	var prefixesFound []string


	for _,iface := range ifaces {
		if _,ok := ifaceMap[iface.Name]; !ok {
			ifaceMap[iface.Name] = true

			// Try to extract the index and prefix
			matches := ifaceRe.FindAllStringSubmatch(iface.Name,-1)

			if len(matches[0]) == 3 {
				tmp,_ := strconv.Atoi(matches[0][2])

				prefixesFound = append(prefixesFound,matches[0][1])
				

				if tmp > lastIndex {
					lastIndex = tmp
				}			


			}
		}
	}


	counter := lastIndex

	// If all the prefixes are the same, then
	// it is probably safe to use the prefix
	if len(prefixesFound) == len(ifaces) {
		prefix = prefixesFound[0]
	}

	// Loop until an available name can
	// be found
	for {

		// Limit the infinite loop to 5000
		// as most devices will not have 5000 interfaces

		if counter >= 5000 {
			return ""
		}

		testName := fmt.Sprintf("%s%d",prefix,counter)
		if _,ok := ifaceMap[testName]; ok {
			counter += 1
			continue
		}

		return testName
		
	}

	return ""
}

func getNetworkMods(exp *Experiment) []*NetworkMod {

	var modifications NetworkMods
	
	// Check for any network modifications defined
	for _,app := range exp.Spec.Scenario.Apps {
		
		if app.Name != "network-mod" {
			continue
		}

		if err := mapstructure.Decode(app.Metadata,&modifications); err != nil {
			logger.Printf("mapsructure can't decode %v",app.Metadata)
		}
		
		break
	}

	return modifications.Mods

}

func applyModification(mod *NetworkMod, topology *TopologySpec) error {

	hostsMap := make(map[string]bool)
	refNet := newIPv4Network(mod.Network)

	if refNet == nil {
		//logger.Printf("Invalid address %v",mod.Network)
		return fmt.Errorf("Invalid address %v",mod.Network)
	}

	// Make sure that there are enough
	// available addresses
	if strings.ToLower(mod.Action) == "add" {
		totalAddresses := refNet.getUsableHostCount()
		usedAddresses,err := getAddressesUsedCount(mod.Network,topology.Nodes)		

		if err != nil {
			return err
		}

		if len(mod.Hosts) > (totalAddresses - usedAddresses) {
			//logger.Printf("Not enough addresses in %s",mod.Network)
			return fmt.Errorf("Not enough addresses in %s",mod.Network)
		}

	}

	// Put the hosts in a hash table for
	// easy lookup
	for _,host := range mod.Hosts {
		if _,ok := hostsMap[host]; !ok{
			hostsMap[host]=true
		}
	}

	nodes := topology.Nodes

	for _,node := range nodes {
		// Skip hosts that are not the target 
		// of the modification
		if len(mod.Hosts) > 0 {
			if _,ok := hostsMap[node.General.Hostname]; !ok {
				continue
			}
		}

		switch strings.ToLower(mod.Action) {
			case "add": {
				// Do not add the network/alias if
				// it already exists
				if exists,_ := interfaceMatch(node,mod); exists {
					continue
				}
				
				addInterface(node,mod,refNet)

			}
			case "delete":{		
				
				// Skip delete actions when both an alias and subnet
				// were not specified
				if len(mod.Alias) == 0 && len(mod.Network) == 0{
					logger.Printf("A subnet and alias were not specified")
					continue
				}
				

				// Do not attempt to remove a network/alias
				// if it does not already exist
				exists,index := interfaceMatch(node,mod)
				if !exists {
					logger.Printf("%s does not exist on %s",mod.Network,node.General.Hostname)
					continue
				}

				node.Network.Interfaces = removeInterface(node.Network.Interfaces,index)

			}
		}
	}

	return nil
}

func addInterface(node *Node,mod *NetworkMod,ipv4 *ipV4Network) error {

	name := findAvailableName(mod.Prefix,node.Network.Interfaces)
	mask := ipv4.cidr

	address := ipv4.getNextAddress(addressesUsed)

	if len(address) == 0 {		
		return fmt.Errorf("unable to obtain an available IPv4 address in %s",ipv4.printShort())
	}

	newInterface := Interface{
		Name:name,
		VLAN:mod.Alias,
		Address:address,
		Mask:mask,
		Proto:"static",
		Gateway:mod.Gateway,
	}

	// Add the interface to the array/slice of interfaces
	node.Network.Interfaces = append(node.Network.Interfaces,newInterface)

	decAddress := addressToBinary(address)
	
	// Add the address to the map of used addresses
	addressesUsed[decAddress] = true

	return nil

}

func removeInterface(ifaces []Interface,index int) []Interface {

	if len(ifaces) == 0 {
		return ifaces
	}	

	// TODO if interfaces is equal to 1, should
	// we allow the last interface to be removed?

	return append(ifaces[:index],ifaces[index+1:]...)
}

func interfaceMatch(node *Node,networkMod *NetworkMod) (bool,int) {
	

	for index, iface := range node.Network.Interfaces {

		// First check the vlan alias
		if len(networkMod.Alias) > 0 {
			if iface.VLAN == networkMod.Alias {
				return true,index
			}
		}

		_,refNet,err := net.ParseCIDR(networkMod.Network)
				
		if err != nil {					
			logger.Printf("Unable to parse network:%v",networkMod.Network)
			continue
		}		
		
		
		
		address := net.ParseIP(iface.Address)		

		if address == nil {
			logger.Printf("Unable to parse address:%v",iface.Address)
			continue
		}

		match := refNet.Contains(address)					
		if match {
			return match,index
		}
		
	}

	return false,-1
}

func getAddressesUsedCount(subnet string,nodes []*Node) (int,error) {

	_,refNet,err := net.ParseCIDR(subnet)
				
	if err != nil {
		return 0,fmt.Errorf("Unable to parse network:%v",subnet)
	}	

	used := make(map[string]bool)

	for _,node := range nodes {
		for _,iface := range node.Network.Interfaces {

			address := net.ParseIP(iface.Address)		

			if address == nil {
				logger.Printf("Unable to parse address:%v",iface.Address)
				continue
			}

			match := refNet.Contains(address)	
			
			// Only looking for addresses contained in the
			// specified subnet
			if !match {
				continue
			}
			

			if _,ok := used[iface.Address]; !ok {
				used[iface.Address] = true
			}

		}
	}

	return len(used),nil


}

func isInSubnet(network string,address string) (bool,error) {

	_,refNet,err := net.ParseCIDR(network)
				
	if err != nil {
		return false,fmt.Errorf("Unable to parse network:%v",network)
	}	

	netAddress := net.ParseIP(address)		

	if netAddress == nil {
		return false,fmt.Errorf("Unable to parse address:%v",netAddress)
		
	}

	return refNet.Contains(netAddress),nil
			
						

}

