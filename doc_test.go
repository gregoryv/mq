package mq

func DocumentFlags(p Packet) string {
	switch p.(type) {
	case *Connect:
		return `        3210 76543210 ProtocolVersion ClientID KeepAlive Size

3-0 reserved

7 u   User Name Flag
6 p   Password Flag
5 r   Will Retain
4 2|! Will QoS
3 1|! Will QoS
2 w   Will Flag
1 s   Clean Start
0     reserved
`
	case *ConnAck:
		return `        3210 76543210 AssignedClientID Size

3-0 reserved

7-1 reserved
0 s Session present
`
	default:
		return p.String()
	}
}
