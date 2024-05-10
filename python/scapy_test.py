from scapy.all import *

class Test(Packet):
	name = "Test"
	fields_desc=[ XByteField("id",1),
                 XByteField("id2",1),
				 FieldLenField("len", None, length_of="data"),
				 FieldListField("data",[0,2,3,20,ord('A')],XByteField('val',0), 
				 length_from=lambda pkt: pkt.len),
                  ]

if __name__ == '__main__':
	t = Test(data=[1,20,30])
	p = IP(src="127.0.0.1",dst="127.0.0.1")/UDP(sport=43553,dport=52500)/t
	p.show2()
	raw(p)
	#print("Packet:{}".format(p.show()))
	send(p)