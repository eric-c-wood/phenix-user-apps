% for route in routes:
% if include_address(interface['address'],interface['mask'],route['next']):
${network_address(route['destination'])} via ${route['next']} metric ${route['cost']}
% endif
% endfor

<%!
    import ipaddress

    def include_address(interface_address,mask,next_hop):
    
        if interface_address == next_hop:
            return True

        ipv4 = ipaddress.IPv4Network("{}/{}".format(interface_address,mask),strict=False)
        ipv4_next_hop = ipaddress.IPv4Network("{}/32".format(next_hop),strict=False)
        return ipv4.overlaps(ipv4_next_hop)

    def network_address(destination):

        return ipaddress.IPv4Network(destination,strict=False).exploded
%>