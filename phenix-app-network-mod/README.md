## network-mod

The `network-mod` app currently adds or deletes VM interfaces.  When
adding interfaces, the network, vlan, alias, and the set of VMs
can be specified.  Below are several examples showing how to use
the network-mod app.  A list of modifications can be specified where
interfaces are both added and deleted in the same scenario.  Gateways 
can be specified when the network needs to reach other networks outside the subnet.

### Add a network to a set of hosts
**action**:`[add,delete]`  
**network**:`Optional:ipv4 address/cidr`  
**vlan**:`Optional:vlan numeric id`  
**alias**:`Optional:descriptive name for vlan`  
**prefix**:`Optional:interface prefix to use.`  
**type**:`Optional:interface type to use.`  
**gateway**:`Optional:gateway to use for the subnet.  This should be
specified when needing to reach other networks outside the subnet`  
**hosts**:`Optional:list of VMs to apply the action to`  

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


   