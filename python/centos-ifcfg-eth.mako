Type=Ethernet
BOOTPROTO=static
NM_CONTROLLED=no
Device=${name}
Name=${name}
ONBOOT=yes
HWADDR=${interface['mac']}
IPADDR=${interface['address']}
NETMASK=${cidr_to_netmask(interface['mask'])}
GATEWAY=${interface['gateway']}
DEFROUTE=yes
IPV6INIT=no
IPV6_AUTOCONF=no

<%!
    import socket
    import struct 

    def cidr_to_netmask(cidr):
        cidr = int(cidr)
        bits = 0xffffffff ^ (1 << 32 - cidr) - 1

        return socket.inet_ntoa(struct.pack('>I',bits))
%>
