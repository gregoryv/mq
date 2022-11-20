package mq

import (
	"fmt"
	"io"
)

func NewAuth() *Auth {
	return &Auth{fixed: bits(AUTH)}
}

type Auth struct {
	fixed        bits
	reasonCode   wuint8
	reasonString wstring
	authMethod   wstring
	authData     bindata
	UserProperties
}

func (p *Auth) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *Auth) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *Auth) SetAuthMethod(v string) { p.authMethod = wstring(v) }
func (p *Auth) AuthMethod() string     { return string(p.authMethod) }

func (p *Auth) SetAuthData(v []byte) { p.authData = bindata(v) }
func (p *Auth) AuthData() []byte     { return []byte(p.authData) }

func (p *Auth) SetReasonString(v string) { p.reasonString = wstring(v) }
func (p *Auth) ReasonString() string     { return string(p.reasonString) }

func (p *Auth) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(),
		p.width(),
	)
}

func (p *Auth) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *Auth) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.reasonCode)
	b.getAny(p.propertyMap(), p.appendUserProperty)
	return b.err
}

func (p *Auth) width() int {
	return p.fill(_LEN, 0)
}

func (p *Auth) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)

	return i
}

func (p *Auth) variableHeader(b []byte, i int) int {
	n := i
	proplen := p.properties(_LEN, 0)
	if p.reasonCode == 0 && proplen == 0 {
		return 0
	}
	i += p.reasonCode.fill(b, i)
	i += vbint(proplen).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *Auth) properties(b []byte, i int) int {
	n := i

	// using c.propertyMap is slow compared to direct field access
	i += p.authMethod.fillProp(b, i, AuthMethod)
	i += p.authData.fillProp(b, i, AuthData)
	i += p.reasonString.fillProp(b, i, ReasonString)
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *Auth) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		AuthMethod:   &p.authMethod,
		AuthData:     &p.authData,
		ReasonString: &p.reasonString,
	}
}
