package mq

import (
	"bytes"
	"strings"
)

func DocumentFlags(p Packet) string {
	var buf bytes.Buffer
	name := p.String()
	i := strings.Index(name, " ")
	buf.WriteString(strings.Repeat(" ", i+1))

	switch p.(type) {

	case *Connect:
		buf.WriteString(`3210 76543210 ProtocolVersion ClientID KeepAlive Size

3-0 reserved

7 u   User Name Flag
6 p   Password Flag
5 r   Will Retain
4 2|! Will QoS
3 1|! Will QoS
2 w   Will Flag
1 s   Clean Start
0     reserved
`)
	case *ConnAck:
		buf.WriteString(`3210 76543210 AssignedClientID Size [ReasonCode and ReasonString if error]

3-0 reserved

7-1 reserved
0 s Session present
`)

	case *PubAck:
		buf.WriteString(`3210 PacketID ReasonString Size [reason text]

3-0 reserved
`)
	case *SubAck:
		buf.WriteString(`3210 PacketID Size

3-0 reserved
`)
	case *Publish:
		buf.WriteString(`3210 PacketID Topic [CorrelationData] Size

3 d   Duplicate
2 2|! QoS
1 1|! QoS
0 r   Retain
`)

	default:
		buf.WriteString(`3210 Size

3-0 reserved
`)
	}
	return buf.String()
}
