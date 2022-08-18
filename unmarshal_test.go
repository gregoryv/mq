package mqtt

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"testing"
)

func TestConnect_UnmarshalBinary(t *testing.T) {
	// data is taken from TestConnect output
	data := []byte{
		16, 208, 1, 0, 4, 77, 81, 84, 84, 5, 36, 1, 43, 48, 17,
		0, 0, 0, 30, 39, 0, 0, 16, 0, 34, 0, 128, 25, 1, 23,
		1, 21, 0, 6, 100, 105, 103, 101, 115, 116, 22, 0, 6, 115, 101,
		99, 114, 101, 116, 38, 0, 5, 99, 111, 108, 111, 114, 0, 3, 114,
		101, 100, 0, 4, 109, 97, 99, 121, 100, 24, 0, 0, 0, 111, 1,
		1, 2, 0, 0, 0, 100, 3, 0, 16, 97, 112, 112, 108, 105, 99,
		97, 116, 105, 111, 110, 47, 106, 115, 111, 110, 8, 0, 16, 114, 101,
		115, 112, 111, 110, 115, 101, 47, 116, 111, 47, 109, 97, 99, 121, 9,
		0, 14, 112, 101, 114, 104, 97, 112, 115, 32, 97, 32, 117, 117, 105,
		100, 38, 0, 9, 99, 111, 110, 110, 101, 99, 116, 101, 100, 0, 19,
		50, 48, 50, 50, 45, 48, 49, 45, 48, 49, 32, 49, 52, 58, 52,
		52, 58, 51, 50, 0, 18, 116, 111, 112, 105, 99, 47, 100, 101, 97,
		100, 47, 99, 108, 105, 101, 110, 116, 115, 0, 20, 123, 34, 99, 108,
		105, 101, 110, 116, 73, 68, 34, 58, 32, 34, 109, 97, 99, 121, 34,
		125,
	}

	r := bytes.NewReader(data)

	// use fixed buffer for reading to minimize memory consumption
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)

	f := buf[0]   // first byte
	var rem vbint // remaining length
	if err := rem.UnmarshalBinary(buf[1:n]); err != nil {
		t.Fatal(err)
	}

	c := Connect{fixed: Bits(f)}
	if err := c.UnmarshalBinary(buf[1+rem.width() : n]); err != nil {
		t.Log(c.String())
		t.Fatal(err)
	}
	var out bytes.Buffer
	c.WriteTo(&out)
	t.Logf("\n\n%s\n\n", hex.Dump(out.Bytes()))
}

func (c *Connect) UnmarshalBinary(p []byte) error {
	// get guards against errors, it also advances the index
	var i int
	var err error
	get := func(v encoding.BinaryUnmarshaler) {
		if err != nil {
			return
		}
		if err = v.UnmarshalBinary(p[i:]); err != nil {
			return
		}
		i += v.(interface{ width() int }).width()
	}
	get(&c.protocolName)
	get(&c.protocolVersion)
	get(&c.flags)
	get(&c.keepAlive)

	// properties
	var propLen vbint
	get(&propLen)
	end := i + int(propLen)

	var id Ident

	// fields map property code to the correct Connect field
	fields := map[Ident]wireType{
		ReceiveMax:            &c.receiveMax,
		SessionExpiryInterval: &c.sessionExpiryInterval,
		MaxPacketSize:         &c.maxPacketSize,
		TopicAliasMax:         &c.topicAliasMax,
		RequestResponseInfo:   &c.requestResponseInfo,
		RequestProblemInfo:    &c.requestProblemInfo,
		AuthMethod:            &c.authMethod,
		AuthData:              &c.authData,

		// will fields
		WillDelayInterval:      &c.willDelayInterval,
		PayloadFormatIndicator: &c.willPayloadFormat,
		MessageExpiryInterval:  &c.willMessageExpiryInterval,
		ContentType:            &c.willContentType,
		ResponseTopic:          &c.responseTopic,
		CorrelationData:        &c.correlationData,
	}

	for i < end {
		get(&id)
		if field, ok := fields[id]; ok {
			get(field)
			continue
		}
		switch id {
		case UserProperty:
			var p property
			get(&p)
			c.AddUserProperty(p)
		}

	}
	// payload
	get(&c.clientID)
	if Bits(c.flags).Has(WillFlag) {
		var willLen vbint
		get(&willLen)
		end := i + int(willLen)

		for i < end {
			get(&id)
			if field, ok := fields[id]; ok {
				get(field)
				continue
			}
			switch id {
			case UserProperty:
				var p property
				get(&p)
				c.willProp = append(c.willProp, p)
				continue
			}
			break
		}
		// todo continue here
	}
	return err
}

type wireType interface {
	encoding.BinaryUnmarshaler
	fill([]byte, int) int
}
