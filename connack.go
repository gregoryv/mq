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

func (c *ConnAck) HasFlag(v byte) bool { return c.flags.Has(v) }

func (c *ConnAck) SetSessionPresent(v bool) { c.flags.toggle(1, true) }
func (c *ConnAck) SessionPresent() bool     { return c.flags.Has(1) }

func (c *ConnAck) SetSessionExpiryInterval(v uint32) { c.sessionExpiryInterval = wuint32(v) }
func (c *ConnAck) SessionExpiryInterval() uint32     { return uint32(c.sessionExpiryInterval) }

func (c *ConnAck) SetReceiveMax(v uint16) { c.receiveMax = wuint16(v) }
func (c *ConnAck) ReceiveMax() uint16     { return uint16(c.receiveMax) }

func (c *ConnAck) SetMaxQoS(v uint8) { c.maxQoS = wuint8(v) }
func (c *ConnAck) MaxQoS() uint8     { return uint8(c.maxQoS) }

func (c *ConnAck) SetRetainAvailable(v bool) { c.retainAvailable = wbool(v) }
func (c *ConnAck) RetainAvailable() bool     { return bool(c.retainAvailable) }

func (c *ConnAck) SetMaxPacketSize(v uint32) { c.maxPacketSize = wuint32(v) }
func (c *ConnAck) MaxPacketSize() uint32     { return uint32(c.maxPacketSize) }

func (c *ConnAck) SetAssignedClientID(v string) { c.assignedClientID = wstring(v) }
func (c *ConnAck) AssignedClientID() string     { return string(c.assignedClientID) }

func (c *ConnAck) SetTopicAliasMax(v uint16) { c.topicAliasMax = wuint16(v) }
func (c *ConnAck) TopicAliasMax() uint16     { return uint16(c.topicAliasMax) }

func (c *ConnAck) SetReasonCode(v ReasonCode) { c.reasonCode = wuint8(v) }
func (c *ConnAck) ReasonCode() ReasonCode     { return ReasonCode(c.reasonCode) }

func (c *ConnAck) SetReasonString(v string) { c.reasonString = wstring(v) }
func (c *ConnAck) ReasonString() string     { return string(c.reasonString) }

func (c *ConnAck) SetWildcardSubAvailable(v bool) { c.wildcardSubAvailable = wbool(v) }
func (c *ConnAck) WildcardSubAvailable() bool     { return bool(c.wildcardSubAvailable) }

func (c *ConnAck) SetSubIdentifiersAvailable(v bool) { c.subIdentifiersAvailable = wbool(v) }
func (c *ConnAck) SubIdentifiersAvailable() bool     { return bool(c.subIdentifiersAvailable) }

func (c *ConnAck) SetSharedSubAvailable(v bool) { c.sharedSubAvailable = wbool(v) }
func (c *ConnAck) SharedSubAvailable() bool     { return bool(c.sharedSubAvailable) }

func (c *ConnAck) SetServerKeepAlive(v uint16) { c.serverKeepAlive = wuint16(v) }
func (c *ConnAck) ServerKeepAlive() uint16     { return uint16(c.serverKeepAlive) }

func (c *ConnAck) SetResponseInformation(v string) { c.responseInformation = wstring(v) }
func (c *ConnAck) ResponseInformation() string     { return string(c.responseInformation) }

func (c *ConnAck) SetServerReference(v string) { c.serverReference = wstring(v) }
func (c *ConnAck) ServerReference() string     { return string(c.serverReference) }

func (c *ConnAck) SetAuthMethod(v string) { c.authMethod = wstring(v) }
func (c *ConnAck) AuthMethod() string     { return string(c.authMethod) }

func (c *ConnAck) SetAuthData(v []byte) { c.authData = bindata(v) }
func (c *ConnAck) AuthData() []byte     { return []byte(c.authData) }

// end settings
// ----------------------------------------

func (c *ConnAck) String() string {
	return fmt.Sprintf("%s %s %s%s %v bytes",
		firstByte(c.fixed).String(),
		connAckFlags(c.flags),
		c.assignedClientID,
		func() string {
			if c.ReasonCode() >= 0x80 {
				return " " + c.ReasonCode().String()
			}
			return ""
		}(),
		c.width(),
	)
}

func (c *ConnAck) dump(w io.Writer) {
	fmt.Fprintf(w, "AssignedClientID: %q\n", c.AssignedClientID())
	fmt.Fprintf(w, "AuthData: %q\n", string(c.AuthData()))
	fmt.Fprintf(w, "AuthMethod: %q\n", c.AuthMethod())
	fmt.Fprintf(w, "MaxPacketSize: %v\n", c.MaxPacketSize())
	fmt.Fprintf(w, "MaxQoS: %v\n", c.MaxQoS())
	fmt.Fprintf(w, "ReasonCode: %v\n", c.ReasonCode())
	fmt.Fprintf(w, "ReasonString: %q\n", c.ReasonString())
	fmt.Fprintf(w, "ReceiveMax: %v\n", c.ReceiveMax())
	fmt.Fprintf(w, "ResponseInformation: %q\n", c.ResponseInformation())
	fmt.Fprintf(w, "RetainAvailable: %v\n", c.RetainAvailable())
	fmt.Fprintf(w, "ServerKeepAlive: %v\n", c.ServerKeepAlive())
	fmt.Fprintf(w, "ServerReference: %q\n", c.ServerReference())
	fmt.Fprintf(w, "SessionExpiryInterval: %v\n", c.SessionExpiryInterval())
	fmt.Fprintf(w, "SessionPresent: %v\n", c.SessionPresent())
	fmt.Fprintf(w, "SharedSubAvailable: %v\n", c.SharedSubAvailable())
	fmt.Fprintf(w, "SubIdentifiersAvailable: %v\n", c.SubIdentifiersAvailable())
	fmt.Fprintf(w, "TopicAliasMax: %v\n", c.TopicAliasMax())
	fmt.Fprintf(w, "WildcardSubAvailable: %v\n", c.WildcardSubAvailable())
	c.UserProperties.dump(w)
}

// ---------------------------------------- protocol

// WriteTo writes this connect control packet in wire format to the
// given writer.
func (c *ConnAck) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, c.fill(_LEN, 0))
	c.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (c *ConnAck) fill(b []byte, i int) int {
	i += c.fixed.fill(b, i)                          // firstByte header
	i += vbint(c.variableHeader(_LEN, 0)).fill(b, i) // remaining length
	i += c.variableHeader(b, i)                      // variable header
	return i
}

func (c *ConnAck) variableHeader(b []byte, i int) int {
	n := i
	i += c.flags.fill(b, i) // acknowledge flags
	i += c.reasonCode.fill(b, i)
	i += vbint(c.properties(_LEN, 0)).fill(b, i) // Properties len
	i += c.properties(b, i)
	return i - n
}

func (c *ConnAck) properties(b []byte, i int) int {
	n := i
	i += c.receiveMax.fillProp(b, i, ReceiveMax)
	i += c.sessionExpiryInterval.fillProp(b, i, SessionExpiryInterval)
	i += c.maxQoS.fillProp(b, i, MaxQoS)
	i += c.retainAvailable.fillProp(b, i, RetainAvailable)
	i += c.maxPacketSize.fillProp(b, i, MaxPacketSize)
	i += c.assignedClientID.fillProp(b, i, AssignedClientID)
	i += c.topicAliasMax.fillProp(b, i, TopicAliasMax)
	i += c.reasonString.fillProp(b, i, ReasonString)
	i += c.wildcardSubAvailable.fillProp(b, i, WildcardSubAvailable)
	i += c.subIdentifiersAvailable.fillProp(b, i, SubIDsAvailable)
	i += c.sharedSubAvailable.fillProp(b, i, SharedSubAvailable)
	i += c.serverKeepAlive.fillProp(b, i, ServerKeepAlive)
	i += c.responseInformation.fillProp(b, i, ResponseInformation)
	i += c.serverReference.fillProp(b, i, ServerReference)
	i += c.authMethod.fillProp(b, i, AuthMethod)
	i += c.authData.fillProp(b, i, AuthData)
	i += c.UserProperties.properties(b, i)
	return i - n
}

func (c *ConnAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&c.flags)
	b.get(&c.reasonCode)
	b.getAny(c.propertyMap(), c.appendUserProperty)
	return b.err
}

func (c *ConnAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReceiveMax:            &c.receiveMax,
		SessionExpiryInterval: &c.sessionExpiryInterval,
		MaxQoS:                &c.maxQoS,
		RetainAvailable:       &c.retainAvailable,
		MaxPacketSize:         &c.maxPacketSize,
		AssignedClientID:      &c.assignedClientID,
		TopicAliasMax:         &c.topicAliasMax,
		ReasonString:          &c.reasonString,
		WildcardSubAvailable:  &c.wildcardSubAvailable,
		SubIDsAvailable:       &c.subIdentifiersAvailable,
		SharedSubAvailable:    &c.sharedSubAvailable,
		ServerKeepAlive:       &c.serverKeepAlive,
		ResponseInformation:   &c.responseInformation,
		ServerReference:       &c.serverReference,
		AuthMethod:            &c.authMethod,
		AuthData:              &c.authData,
	}
}

func (c *ConnAck) width() int {
	return c.fill(_LEN, 0)
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
