import socket
import platform

platform = platform.system()
if (platform == "Linux"):    
    if not hasattr(socket, "IP_ADD_SOURCE_MEMBERSHIP"):
        setattr(socket, "IP_ADD_SOURCE_MEMBERSHIP",39)
    

MCAST_GRP = ''
MCAST_PORT = 5007
SSM_SOURCE = ''
LOCAL_IFACE = ''

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
#sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
#sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)


mreq = socket.inet_aton(MCAST_GRP) + socket.inet_aton(LOCAL_IFACE) + socket.inet_aton(SSM_SOURCE)
sock.setsockopt(socket.IPPROTO_IP, socket.IP_ADD_SOURCE_MEMBERSHIP, mreq)
sock.bind((MCAST_GRP, MCAST_PORT))

while True:
    data, addr = sock.recvfrom(1024)
    print("Received {} from {}".format(addr,data))
