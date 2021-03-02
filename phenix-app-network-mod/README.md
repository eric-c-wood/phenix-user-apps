## network-mod

The `network-mod` app adds and deletes networks. Networks
are added by adding interfaces to VMs with an ipv4 address in the specified
network.  Likewise, networks are deleted by removing interfaces from VMs by specifying
either the vlan alias or the network.  When adding a network which in turn adds interfaces to
the specified set of VMs/hosts, a variety of interface parameters can be specified for the added interfaces.
Listed below are the interface parameters along with the parameters for the network-mod app.  

### network-mod scenario configuration fields
**action**: [add,delete]  
**network**: Optional ipv4 address/cidr  Default is `172.16.0.0/16`  
**vlan**: Optional numeric id  
**alias**: Optional descriptive name for the vlan  Default is `network#` where # is a counter based on the number of networks being added  
**prefix**: Optional interface prefix to use.  Default is
`eth`  
**type**: Optional interface type to use. Default is `ethernet`  
**gateway**: Optional gateway to use for the subnet.  This should be
specified when needing to reach other networks outside the subnet   
**hosts**: Optional list of VMs to apply the action to.  Default is 
an empty list which means that action wil be applied to all hosts.  The `hosts` key should stil be specified in the
scenario configuration file.  

A list of modifications can be specified where
networks are added and deleted in the same scenario.  Gateways 
can be specified when the network needs to reach other networks outside the subnet.

Listed below are several examples of adding/deleting networks.

### Add a network to a set of hosts

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: network_mod
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: network-mod
      metadata:
        modifications:
          - action: add   
            network: 172.168.1.0/24
            vlan: 30
            alias: mgntB 
            hosts: 
              - site_A_server
              - site_B_workstation
              - site_A_router
              - site_B_router
              - internet_router       
```

### Add a network to all hosts
The network can be added to all VMs by 
not specifying any VMs/hosts

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: network_mod_all
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: network-mod
      metadata:
        modifications:
          - action: add   
            network: 172.168.1.0/24
            vlan: 30
            alias: mgntB 
            hosts:    
```

### Remove a network from all matching hosts
A network can be removed from all matching hosts
by specifying either the network/subnet or the 
vlan alias.  

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: network_mod_delete
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: network-mod
      metadata:
        modifications:
          - action: delete   
            network: 192.168.1.0/30            
            hosts:    
```


   
