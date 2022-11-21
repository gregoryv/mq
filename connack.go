package mq

import (
	"bytes"
	"fmt"
	"io"
)

func NewConnAck() *ConnAck {
	return &ConnAck{
		fixed: bits(CONNACK),
	}
}

type ConnAck struct {
	fixed      bits
	flags      bits // sessionPresent as 7-1 are reserved
	reasonCode wuint8

	// properties
	sessionExpiryInterval wuint32
	receiveMax            wuint16
	maxQoS                wuint8 // 0 or 1, 2
	retainAvailable       wbool
	maxPacketSize         wuint32
	assignedClientID      wstring
	topicAliasMax         wuint16
	reasonString          wstring

	UserProperties
	wildcardSubAvailable    wbool
	subIdentifiersAvailable wbool
	sharedSubAvailable      wbool
	serverKeepAlive         wuint16
	responseInformation     wstring
	serverReference         wstring
	authMethod              wstring
	authData                bindata
}

func (p *ConnAck) HasFlag(v byte) bool { return p.flags.Has(v) }

func (p *ConnAck) SetSessionPresent(v bool) { p.flags.toggle(1, true) }
func (p *ConnAck) SessionPresent() bool     { return p.flags.Has(1) }

func (p *ConnAck) SetSessionExpiryInterval(v uint32) { p.sessionExpiryInterval = wuint32(v) }
func (p *ConnAck) SessionExpiryInterval() uint32     { return uint32(p.sessionExpiryInterval) }

func (p *ConnAck) SetReceiveMax(v uint16) { p.receiveMax = wuint16(v) }
func (p *ConnAck) ReceiveMax() uint16     { return uint16(p.receiveMax) }

func (p *ConnAck) SetMaxQoS(v uint8) { p.maxQoS = wuint8(v) }
func (p *ConnAck) MaxQoS() uint8     { return uint8(p.maxQoS) }

func (p *ConnAck) SetRetainAvailable(v bool) { p.retainAvailable = wbool(v) }
func (p *ConnAck) RetainAvailable() bool     { return bool(p.retainAvailable) }

func (p *ConnAck) SetMaxPacketSize(v uint32) { p.maxPacketSize = wuint32(v) }
func (p *ConnAck) MaxPacketSize() uint32     { return uint32(p.maxPacketSize) }

func (p *ConnAck) SetAssignedClientID(v string) { p.assignedClientID = wstring(v) }
func (p *ConnAck) AssignedClientID() string     { return string(p.assignedClientID) }

func (p *ConnAck) SetTopicAliasMax(v uint16) { p.topicAliasMax = wuint16(v) }
func (p *ConnAck) TopicAliasMax() uint16     { return uint16(p.topicAliasMax) }

func (p *ConnAck) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *ConnAck) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *ConnAck) SetReasonString(v string) { p.reasonString = wstring(v) }
func (p *ConnAck) ReasonString() string     { return string(p.reasonString) }

func (p *ConnAck) SetWildcardSubAvailable(v bool) { p.wildcardSubAvailable = wbool(v) }
func (p *ConnAck) WildcardSubAvailable() bool     { return bool(p.wildcardSubAvailable) }

func (p *ConnAck) SetSubIdentifiersAvailable(v bool) { p.subIdentifiersAvailable = wbool(v) }
func (p *ConnAck) SubIdentifiersAvailable() bool     { return bool(p.subIdentifiersAvailable) }

func (p *ConnAck) SetSharedSubAvailable(v bool) { p.sharedSubAvailable = wbool(v) }
func (p *ConnAck) SharedSubAvailable() bool     { return bool(p.sharedSubAvailable) }

func (p *ConnAck) SetServerKeepAlive(v uint16) { p.serverKeepAlive = wuint16(v) }
func (p *ConnAck) ServerKeepAlive() uint16     { return uint16(p.serverKeepAlive) }

func (p *ConnAck) SetResponseInformation(v string) { p.responseInformation = wstring(v) }
func (p *ConnAck) ResponseInformation() string     { return string(p.responseInformation) }

func (p *ConnAck) SetServerReference(v string) { p.serverReference = wstring(v) }
func (p *ConnAck) ServerReference() string     { return string(p.serverReference) }

func (p *ConnAck) SetAuthMethod(v string) { p.authMethod = wstring(v) }
func (p *ConnAck) AuthMethod() string     { return string(p.authMethod) }

func (p *ConnAck) SetAuthData(v []byte) { p.authData = bindata(v) }
func (p *ConnAck) AuthData() []byte     { return []byte(p.authData) }

// end settings
// ----------------------------------------

func (p *ConnAck) String() string {
	return fmt.Sprintf("%s %s %s%s %v bytes",
		firstByte(p.fixed).String(),
		connAckFlags(p.flags),
		p.assignedClientID,
		func() string {
			if p.ReasonCode() >= 0x80 {
				return " " + p.ReasonCode().String()
			}
			return ""
		}(),
		p.width(),
	)
}

func (p *ConnAck) dump(w io.Writer) {
	fmt.Fprintf(w, "AssignedClientID: %q\n", p.AssignedClientID())
	fmt.Fprintf(w, "AuthData: %q\n", string(p.AuthData()))
	fmt.Fprintf(w, "AuthMethod: %q\n", p.AuthMethod())
	fmt.Fprintf(w, "MaxPacketSize: %v\n", p.MaxPacketSize())
	fmt.Fprintf(w, "MaxQoS: %v\n", p.MaxQoS())
	fmt.Fprintf(w, "ReasonCode: %v\n", p.ReasonCode())
	fmt.Fprintf(w, "ReasonString: %q\n", p.ReasonString())
	fmt.Fprintf(w, "ReceiveMax: %v\n", p.ReceiveMax())
	fmt.Fprintf(w, "ResponseInformation: %q\n", p.ResponseInformation())
	fmt.Fprintf(w, "RetainAvailable: %v\n", p.RetainAvailable())
	fmt.Fprintf(w, "ServerKeepAlive: %v\n", p.ServerKeepAlive())
	fmt.Fprintf(w, "ServerReference: %q\n", p.ServerReference())
	fmt.Fprintf(w, "SessionExpiryInterval: %v\n", p.SessionExpiryInterval())
	fmt.Fprintf(w, "SessionPresent: %v\n", p.SessionPresent())
	fmt.Fprintf(w, "SharedSubAvailable: %v\n", p.SharedSubAvailable())
	fmt.Fprintf(w, "SubIdentifiersAvailable: %v\n", p.SubIdentifiersAvailable())
	fmt.Fprintf(w, "TopicAliasMax: %v\n", p.TopicAliasMax())
	fmt.Fprintf(w, "WildcardSubAvailable: %v\n", p.WildcardSubAvailable())
	p.UserProperties.dump(w)
}

// ---------------------------------------- protocol

// WriteTo writes this connect control packet in wire format to the
// given writer.
func (p *ConnAck) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *ConnAck) fill(b []byte, i int) int {
	i += p.fixed.fill(b, i)                          // firstByte header
	i += vbint(p.variableHeader(_LEN, 0)).fill(b, i) // remaining length
	i += p.variableHeader(b, i)                      // variable header
	return i
}

func (p *ConnAck) variableHeader(b []byte, i int) int {
	n := i
	i += p.flags.fill(b, i) // acknowledge flags
	i += p.reasonCode.fill(b, i)
	i += vbint(p.properties(_LEN, 0)).fill(b, i) // Properties len
	i += p.properties(b, i)
	return i - n
}

func (p *ConnAck) properties(b []byte, i int) int {
	n := i
	i += p.receiveMax.fillProp(b, i, ReceiveMax)
	i += p.sessionExpiryInterval.fillProp(b, i, SessionExpiryInterval)
	i += p.maxQoS.fillProp(b, i, MaxQoS)
	i += p.retainAvailable.fillProp(b, i, RetainAvailable)
	i += p.maxPacketSize.fillProp(b, i, MaxPacketSize)
	i += p.assignedClientID.fillProp(b, i, AssignedClientID)
	i += p.topicAliasMax.fillProp(b, i, TopicAliasMax)
	i += p.reasonString.fillProp(b, i, ReasonString)
	i += p.wildcardSubAvailable.fillProp(b, i, WildcardSubAvailable)
	i += p.subIdentifiersAvailable.fillProp(b, i, SubIDsAvailable)
	i += p.sharedSubAvailable.fillProp(b, i, SharedSubAvailable)
	i += p.serverKeepAlive.fillProp(b, i, ServerKeepAlive)
	i += p.responseInformation.fillProp(b, i, ResponseInformation)
	i += p.serverReference.fillProp(b, i, ServerReference)
	i += p.authMethod.fillProp(b, i, AuthMethod)
	i += p.authData.fillProp(b, i, AuthData)
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *ConnAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.flags)
	b.get(&p.reasonCode)
	b.getAny(p.propertyMap(), p.appendUserProperty)
	return b.err
}

func (p *ConnAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReceiveMax:            &p.receiveMax,
		SessionExpiryInterval: &p.sessionExpiryInterval,
		MaxQoS:                &p.maxQoS,
		RetainAvailable:       &p.retainAvailable,
		MaxPacketSize:         &p.maxPacketSize,
		AssignedClientID:      &p.assignedClientID,
		TopicAliasMax:         &p.topicAliasMax,
		ReasonString:          &p.reasonString,
		WildcardSubAvailable:  &p.wildcardSubAvailable,
		SubIDsAvailable:       &p.subIdentifiersAvailable,
		SharedSubAvailable:    &p.sharedSubAvailable,
		ServerKeepAlive:       &p.serverKeepAlive,
		ResponseInformation:   &p.responseInformation,
		ServerReference:       &p.serverReference,
		AuthMethod:            &p.authMethod,
		AuthData:              &p.authData,
	}
}

func (p *ConnAck) width() int {
	return p.fill(_LEN, 0)
}

// ----------------------------------------

type connAckFlags byte

// String returns flags represented with a letter.
// Improper flags are marked with '!' and unset are marked with '-'.
//
//	SessionPresent s
//	Reserved      !
func (c connAckFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 8)

	mark := func(i int, flag byte, v byte) {
		if !bits(c).Has(flag) {
			return
		}
		flags[i] = v
	}
	mark(0, 1<<7, '!')
	mark(1, 1<<6, '!')
	mark(2, 1<<5, '!')
	mark(3, 1<<4, '!')
	mark(4, 1<<3, '!')
	mark(5, 1<<2, '!')
	mark(6, 1<<1, '!')
	mark(7, 1<<0, 's')

	return string(flags) // + fmt.Sprintf(" %08b", c)
}

// ----------------------------------------

const (
	SessionPresent uint8 = 1
)
