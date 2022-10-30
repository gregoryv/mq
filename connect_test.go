package mq

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func ExampleConnect() {
	c := NewConnect()
	c.SetClientID("macy")
	c.SetKeepAlive(299)
	c.SetUsername("john.doe")
	c.SetPassword([]byte("123"))

	fmt.Print(c.String())
	// output:
	// CONNECT ---- up------ MQTT5 macy 4m59s 34 bytes
}

func ExampleConnect_String() {
	p := NewConnect()
	p.SetClientID("pink")
	p.SetUsername("gopher")
	p.SetPassword([]byte("cute"))
	p.SetWillQoS(1)

	fmt.Println(p.String())
	fmt.Println(DocumentFlags(&p))
	// output:
	// CONNECT ---- up--1--- MQTT5 pink 0s 33 bytes
	//         3210 76543210 ProtocolVersion ClientID KeepAlive Size
	//
	// 3-0 reserved
	//
	// 7 u   User Name Flag
	// 6 p   Password Flag
	// 5 r   Will Retain
	// 4 2|! Will QoS
	// 3 1|! Will QoS
	// 2 w   Will Flag
	// 1 s   Clean Start
	// 0     reserved
}

func TestConnect(t *testing.T) {
	c := NewConnect()

	eq(t, c.SetProtocolVersion, c.ProtocolVersion, 5)
	eq(t, c.SetProtocolName, c.ProtocolName, "MQTT")
	eq(t, c.SetKeepAlive, c.KeepAlive, 299)
	eq(t, c.SetClientID, c.ClientID, "macy")
	eq(t, c.SetSessionExpiryInterval, c.SessionExpiryInterval, 30)
	eq(t, c.SetUsername, c.Username, "john.doe")
	eq(t, c.SetPassword, c.Password, []byte("123"))
	eq(t, c.SetAuthMethod, c.AuthMethod, "digest")
	eq(t, c.SetAuthData, c.AuthData, []byte("secret"))
	eq(t, c.SetMaxPacketSize, c.MaxPacketSize, 4096)
	eq(t, c.SetTopicAliasMax, c.TopicAliasMax, 128)
	eq(t, c.SetRequestResponseInfo, c.RequestResponseInfo, true)
	eq(t, c.SetRequestProblemInfo, c.RequestProblemInfo, true)
	eq(t, c.SetResponseTopic, c.ResponseTopic, "response/to/macy")
	eq(t, c.SetCorrelationData, c.CorrelationData, []byte("perhaps a uuid"))

	c.AddUserProp("color", "red")

	eq(t, c.SetWillRetain, c.WillRetain, true)
	eq(t, c.SetWillQoS, c.WillQoS, 1)
	eq(t, c.SetWillTopic, c.WillTopic, "topic/dead/clients")
	eq(t, c.SetWillPayload, c.WillPayload, []byte(`{"clientID": "macy"}`))
	eq(t, c.SetWillContentType, c.WillContentType, "application/json")
	eq(t, c.SetWillDelayInterval, c.WillDelayInterval, 111)
	eq(t, c.SetWillPayloadFormat, c.WillPayloadFormat, true)
	eq(t, c.SetWillMessageExpiryInterval, c.WillMessageExpiryInterval, 100)
	c.AddWillProp("connected", "2022-01-01 14:44:32")

	if got := c.String(); !strings.Contains(got, "CONNECT") {
		t.Error(got)
	}

	testControlPacket(t, &c)

	// clears it
	if c.SetUsername(""); c.HasFlag(UsernameFlag) {
		t.Error("username flag still set")
	}
	if c.SetPassword(nil); c.HasFlag(PasswordFlag) {
		t.Error("password flag still set")
	}
}

// eq is used to check equality of set and "get" funcs
// Thank you generics.
func eq[T any](t *testing.T, set func(T), get func() T, value T) {
	set(value)
	if got := get(); !reflect.DeepEqual(got, value) {
		t.Helper()
		t.Errorf("got %v, expected %v", got, value)
	}
}

func Test_connectFlags(t *testing.T) {
	f := connectFlags(0b11110110)
	// QoS2
	if got, exp := f.String(), "upr2-ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	// QoS1
	f = connectFlags(0b11101110)
	if got, exp := f.String(), "upr-1ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	f = connectFlags(0b00000001)
	if got, exp := f.String(), "-------!"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
}
