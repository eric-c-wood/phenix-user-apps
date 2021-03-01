package main

type Experiment struct {
	Metadata ConfigMetadata    		 `json:"metadata" yaml:"metadata"` // experiment configuration metadata
	Spec     ExperimentSpec   		 `json:"spec" yaml:"spec"`         // reference to latest versioned experiment spec
	Status   ExperimentStatus 		 `json:"status" yaml:"status"`     // reference to latest versioned experiment status
}

type ExperimentSpec struct {
	ExperimentName 	string            `json:"experimentName" yaml:"experimentName" structs:"experimentName" mapstructure:"experimentName"`
	BaseDir        	string            `json:"baseDir" yaml:"baseDir" structs:"baseDir" mapstructure:"baseDir"`
	Topology      	*TopologySpec     `json:"topology" yaml:"topology" structs:"topology" mapstructure:"topology"`
	Scenario      	*ScenarioSpec     `json:"scenario" yaml:"scenario" structs:"scenario" mapstructure:"scenario"`
	VLANs          	*VLANSpec         `json:"vlans" yaml:"vlans" structs:"vlans" mapstructure:"vlans"`
	Schedules      	map[string]string `json:"schedules" yaml:"schedules" structs:"schedules" mapstructure:"schedules"`
	RunLocal       	bool              `json:"runLocal" yaml:"runLocal" structs:"runLocal" mapstructure:"runLocal"`
}


type ExperimentStatus struct {
	StartTime string                 `json:"startTime" yaml:"startTime" structs:"startTime" mapstructure:"startTime"`
	Schedules map[string]string      `json:"schedules" yaml:"schedules" structs:"schedules" mapstructure:"schedules"`
	Apps      map[string]interface{} `json:"apps" yaml:"apps" structs:"apps" mapstructure:"apps"`
	VLANs     map[string]int         `json:"vlans" yaml:"vlans" structs:"vlans" mapstructure:"vlans"`
}

type VLANSpec struct {
	Aliases map[string]int `json:"aliases" yaml:"aliases" structs:"aliases" mapstructure:"aliases"`
	Min     int            `json:"min" yaml:"min" structs:"min" mapstructure:"min"`
	Max     int            `json:"max" yaml:"max" structs:"max" mapstructure:"max"`
}

type (
	Configs     []Config
	Annotations map[string]string
)

type Config struct {
	Version  string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind     string                 `json:"kind" yaml:"kind"`
	Metadata ConfigMetadata         `json:"metadata" yaml:"metadata"`
	Spec     map[string]interface{} `json:"spec" yaml:"spec"`
	Status   map[string]interface{} `json:"status,omitempty" yaml:"status,omitempty"`
}

type ConfigMetadata struct {
	Name        string      `json:"name" yaml:"name"`
	Created     string      `json:"created" yaml:"created"`
	Updated     string      `json:"updated" yaml:"updated"`
	Annotations Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}


type ScenarioSpec struct {
	Apps []ScenarioApp `json:"apps" yaml:"apps" structs:"apps" mapstructure:"apps"`
}

type ScenarioApp struct {
	Name     string                 `json:"name" yaml:"name" structs:"name" mapstructure:"name"`
	AssetDir string                 `json:"assetDir" yaml:"assetDir" structs:"assetDir" mapstructure:"assetDir"`
	Metadata map[string]interface{}	`json:"metadata" yaml:"metadata" structs:"metadata" mapstructure:"metadata"`
	Hosts    []ScenarioAppHost      `json:"hosts" yaml:"hosts" structs:"hosts" mapstructure:"hosts"`
}

type ScenarioAppHost struct {
	Hostname string                 `json:"hostname" yaml:"hostname" structs:"hostname" mapstructure:"hostname"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata" structs:"metadata" mapstructure:"metadata"`
}

type Host struct {
	Hostname string                 `json:"hostname" yaml:"hostname" structs:"hostname" mapstructure:"hostname"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata" structs:"metadata" mapstructure:"metadata"`
}

type TopologySpec struct {
	Nodes []*Node `json:"nodes" yaml:"nodes" structs:"nodes" mapstructure:"nodes"`
}

type Node struct {
	Labels    	map[string]string `json:"labels" yaml:"labels" structs:"labels" mapstructure:"labels"`
	Type      	string            `json:"type" yaml:"type" structs:"type" mapstructure:"type"`
	General   	*General          `json:"general" yaml:"general" structs:"general" mapstructure:"general"`
	Hardware   	*Hardware         `json:"hardware" yaml:"hardware" structs:"hardware" mapstructure:"hardware"`
	Network    	*Network          `json:"network" yaml:"network" structs:"network" mapstructure:"network"`
	Injections 	[]*Injection      `json:"injections" yaml:"injections" structs:"injections" mapstructure:"injections"`
}

type General struct {
	Hostname    string `json:"hostname" yaml:"hostname" structs:"hostname" mapstructure:"hostname"`
	Description string `json:"description" yaml:"description" structs:"description" mapstructure:"description"`
	VMType      string `json:"vm_type" yaml:"vm_type" structs:"vm_type" mapstructure:"vm_type"`
	Snapshot    *bool  `json:"snapshot" yaml:"snapshot" structs:"snapshot" mapstructure:"snapshot"`
	DoNotBoot   *bool  `json:"do_not_boot" yaml:"do_not_boot" structs:"do_not_boot" mapstructure:"do_not_boot"`
}

type Hardware struct {
	CPU    string   `json:"cpu" yaml:"cpu" structs:"cpu" mapstructure:"cpu"`
	VCPU   int      `json:"vcpus" yaml:"vcpus" structs:"vcpus" mapstructure:"vcpus"`
	Memory int      `json:"memory" yaml:"memory" structs:"memory" mapstructure:"memory"`
	OSType string   `json:"os_type" yaml:"os_type" structs:"os_type" mapstructure:"os_type"`
	Drives []*Drive `json:"drives" yaml:"drives" structs:"drives" mapstructure:"drives"`
}

type Drive struct {
	Image           string `json:"image" yaml:"image" structs:"image" mapstructure:"image"`
	Iface           string `json:"interface" yaml:"interface" structs:"interface" mapstructure:"interface"`
	CacheMode       string `json:"cache_mode" yaml:"cache_mode" structs:"cache_mode" mapstructure:"cache_mode"`
	InjectPartition *int   `json:"inject_partition" yaml:"inject_partition" structs:"inject_partition" mapstructure:"inject_partition"`
}

type Injection struct {
	Src         string `json:"src" yaml:"src" structs:"src" mapstructure:"src"`
	Dst         string `json:"dst" yaml:"dst" structs:"dst" mapstructure:"dst"`
	Description string `json:"description" yaml:"description" structs:"description" mapstructure:"description"`
	Permissions string `json:"permissions" yaml:"permissions" structs:"permissions" mapstructure:"permissions"`
}

type Network struct {
	Interfaces []Interface `json:"interfaces" yaml:"interfaces" structs:"interfaces" mapstructure:"interfaces"`
	Routes     []Route     `json:"routes" yaml:"routes" structs:"routes" mapstructure:"routes"`
	OSPF       *OSPF       `json:"ospf" yaml:"ospf" structs:"ospf" mapstructure:"ospf"`
	Rulesets   []Ruleset   `json:"rulesets" yaml:"rulesets" structs:"rulesets" mapstructure:"rulesets"`
}

type Interface struct {
	Name       string `json:"name" yaml:"name" structs:"name" mapstructure:"name"`
	Type       string `json:"type" yaml:"type" structs:"type" mapstructure:"type"`
	Proto      string `json:"proto" yaml:"proto" structs:"proto" mapstructure:"proto"`
	UDPPort    int    `json:"udp_port" yaml:"udp_port" structs:"udp_port" mapstructure:"udp_port"`
	BaudRate   int    `json:"baud_rate" yaml:"baud_rate" structs:"baud_rate" mapstructure:"baud_rate"`
	Device     string `json:"device" yaml:"device" structs:"device" mapstructure:"device"`
	VLAN       string `json:"vlan" yaml:"vlan" structs:"vlan" mapstructure:"vlan"`
	Bridge     string `json:"bridge" yaml:"bridge" structs:"bridge" mapstructure:"bridge"`
	Autostart  bool   `json:"autostart" yaml:"autostart" structs:"autostart" mapstructure:"autostart"`
	MAC        string `json:"mac" yaml:"mac" structs:"mac" mapstructure:"mac"`
	MTU        int    `json:"mtu" yaml:"mtu" structs:"mtu" mapstructure:"mtu"`
	Address    string `json:"address" yaml:"address" structs:"address" mapstructure:"address"`
	Mask       int    `json:"mask" yaml:"mask" structs:"mask" mapstructure:"mask"`
	Gateway    string `json:"gateway" yaml:"gateway" structs:"gateway" mapstructure:"gateway"`
	RulesetIn  string `json:"ruleset_in" yaml:"ruleset_in" structs:"ruleset_in" mapstructure:"ruleset_in"`
	RulesetOut string `json:"ruleset_out" yaml:"ruleset_out" structs:"ruleset_out" mapstructure:"ruleset_out"`
}

type Route struct {
	Destination string `json:"destination" yaml:"destination" structs:"destination" mapstructure:"destination"`
	Next        string `json:"next" yaml:"next" structs:"next" mapstructure:"next"`
	Cost        *int   `json:"cost" yaml:"cost" structs:"cost" mapstructure:"cost"`
}

type OSPF struct {
	RouterID               string `json:"router_id" yaml:"router_id" structs:"router_id" mapstructure:"router_id"`
	Areas                  []Area `json:"areas" yaml:"areas" structs:"areas" mapstructure:"areas"`
	DeadInterval           *int   `json:"dead_interval" yaml:"dead_interval" structs:"dead_interval" mapstructure:"dead_interval"`
	HelloInterval          *int   `json:"hello_interval" yaml:"hello_interval" structs:"hello_interval" mapstructure:"hello_interval"`
	RetransmissionInterval *int   `json:"retransmission_interval" yaml:"retransmission_interval" structs:"retransmission_interval" mapstructure:"retransmission_interval"`
}

type Area struct {
	AreaID       *int          `json:"area_id" yaml:"area_id" structs:"area_id" mapstructure:"area_id"`
	AreaNetworks []AreaNetwork `json:"area_networks" yaml:"area_networks" structs:"area_networks" mapstructure:"area_networks"`
}

type AreaNetwork struct {
	Network string `json:"network" yaml:"network" structs:"network" mapstructure:"network"`
}

type Ruleset struct {
	NameF        string `json:"name" yaml:"name" structs:"name" mapstructure:"name"`
	DescriptionF string `json:"description" yaml:"description" structs:"description" mapstructure:"description"`
	DefaultF     string `json:"default" yaml:"default" structs:"default" mapstructure:"default"`
	RulesF       []Rule `json:"rules" yaml:"rules" structs:"rules" mapstructure:"rules"`
}

type Rule struct {
	ID          int       `json:"id" yaml:"id" structs:"id" mapstructure:"id"`
	Description string    `json:"description" yaml:"description" structs:"description" mapstructure:"description"`
	Action      string    `json:"action" yaml:"action" structs:"action" mapstructure:"action"`
	Protocol    string    `json:"protocol" yaml:"protocol" structs:"protocol" mapstructure:"protocol"`
	Source      *AddrPort `json:"source" yaml:"source" structs:"source" mapstructure:"source"`
	Destination *AddrPort `json:"destination" yaml:"destination" structs:"destination" mapstructure:"destination"`
}

type AddrPort struct {
	Address string `json:"address" yaml:"address" structs:"address" mapstructure:"address"`
	Port    int    `json:"port" yaml:"port" structs:"port" mapstructure:"port"`
}


