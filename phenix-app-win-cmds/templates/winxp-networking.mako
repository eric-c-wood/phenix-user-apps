
% for i in range(0,len(node['network']['interfaces'])):
<% interface = node['network']['interfaces'][i] %>
% if i == 0:
netsh interface ip set address name="Local Area Connection" static ${interface['address']} ${cidr_to_netmask(interface['mask'])}
% else:
<% counter = i + 1 %>
netsh interface ip set address name="Local Area Connection ${counter}" static ${interface['address']} ${cidr_to_netmask(interface)}
% endif
% endfor


% for i in range(0,len(node['network']['routes'])):
<% route = node['network']['routes'][i] %>
route -p add ${network_address(route['destination'])} mask ${get_mask(route['destination'])} ${route['next']} metric ${route['cost']}
% endfor

<%!
    import ipaddress

    counter = 0
    interface = None
    route = None

    def cidr_to_netmask(interface): 
        network = "{}/{}".format(interface['address'],interface['mask'])
        return ipaddress.IPv4Network(network,strict=False).netmask

    def get_mask(destination):         
        return ipaddress.IPv4Network(destination,strict=False).netmask

    def network_address(destination):

        return ipaddress.IPv4Network(destination,strict=False).network_address

%>