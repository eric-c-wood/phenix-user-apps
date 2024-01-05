
% for i in range(0,len(node['network']['interfaces'])):
<% interface = node['network']['interfaces'][i] %>
% if i == 0:
netsh interface ipv4 set address name="Local Area Connection" static ${interface['address']} ${cidr_to_netmask(interface['mask'])}
% else:
<% counter = i + 1 %>
netsh interface ipv4 set address name="Local Area Connection ${counter}" static ${interface['address']} ${mask_from_interface(interface)}
% endif
% endfor


% for i in range(0,len(node['network']['routes'])):
<% route = node['network']['routes'][i] %>
route -p add ${network_address(route['destination'])} mask ${mask_from_destination(route['destination'])} ${route['next']} metric ${route['cost']}
% endfor

<%!
    import ipaddress

    counter = 0
    interface = None
    route = None

    def mask_from_interface(interface): 
        network = "{}/{}".format(interface['address'],interface['mask'])
        return ipaddress.IPv4Network(network,strict=False).netmask

    def mask_from_destination(destination):         
        return ipaddress.IPv4Network(destination,strict=False).netmask

    def network_address(destination):

        return ipaddress.IPv4Network(destination,strict=False).network_address

%>