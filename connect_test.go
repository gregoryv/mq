package mqtt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func ExampleConnect() {
	c := NewConnect()
	c.SetClientID("macy")
	c.SetKeepAlive(299)
	c.SetUsername("john.doe")
	c.SetPassword([]byte("123"))

	fmt.Print(c)
	// output:
	// CONNECT ---- up----- MQTT5 4m59s 34 bytes
}

func TestConnect(t *testing.T) {
	//
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

	c.AddUserProp("color", "red")

	eq(t, c.SetMaxPacketSize, c.MaxPacketSize, 4096)
	eq(t, c.SetTopicAliasMax, c.TopicAliasMax, 128)
	eq(t, c.SetRequestResponseInfo, c.RequestResponseInfo, true)
	eq(t, c.SetRequestProblemInfo, c.RequestProblemInfo, true)
	eq(t, c.SetResponseTopic, c.ResponseTopic, "response/to/macy")
	eq(t, c.SetCorrelationData, c.CorrelationData, []byte("perhaps a uuid"))

	eq(t, c.SetWillRetain, c.WillRetain, true)
	eq(t, c.SetWillTopic, c.WillTopic, "topic/dead/clients")
	eq(t, c.SetWillPayload, c.WillPayload, []byte(`{"clientID": "macy"}`))
	eq(t, c.SetWillContentType, c.WillContentType, "application/json")
	eq(t, c.SetWillDelayInterval, c.WillDelayInterval, 111)
	eq(t, c.SetWillPayloadFormat, c.WillPayloadFormat, true)
	eq(t, c.SetWillMessageExpiryInterval, c.WillMessageExpiryInterval, 100)
	c.AddWillProp("connected", "2022-01-01 14:44:32")

	var buf bytes.Buffer
	c.WriteTo(&buf)
	//t.Logf("\n\n%s\n\n%s\n\n", c, hex.Dump(buf.Bytes()))

	c.SetUsername("") // unset toggles flag
	c.SetPassword(nil)

	if c.Flags().Has(UsernameFlag) {
		t.Error("still has", UsernameFlag)
	}

	if got := c.String(); !strings.Contains(got, "CONNECT") {
		t.Error(got)
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

func TestconnectFlags(t *testing.T) {
	f := connectFlags(0b11110110)
	// QoS2
	if got, exp := f.String(), "upr2ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	// QoS1
	f = connectFlags(0b11101110)
	if got, exp := f.String(), "upr1ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	f = connectFlags(0b00000001)
	if got, exp := f.String(), "------!"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
}

func BenchmarkCreateAndWriteTo(b *testing.B) {
	var (
		alive   = uint16(30)
		cid     = "macy"
		user    = "john.doe"
		pwd     = []byte("secret")
		sExpiry = uint32(30)
	)
	b.Run("our", func(b *testing.B) {

		for n := 0; n < b.N; n++ {

			// our packet
			our := NewConnect()
			our.SetKeepAlive(alive)
			our.SetClientID(cid)
			our.SetUsername(user)
			our.SetPassword(pwd)
			our.SetSessionExpiryInterval(sExpiry)
			our.WriteTo(ioutil.Discard)
		}
	})
	b.Run("their", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			their := packets.NewControlPacket(packets.CONNECT)
			c := their.Content.(*packets.Connect)
			c.KeepAlive = alive
			c.ClientID = cid
			c.UsernameFlag = true
			c.Username = user
			c.PasswordFlag = true
			c.Password = pwd
			c.Properties.SessionExpiryInterval = &sExpiry
			their.WriteTo(ioutil.Discard)
		}
	})

}

func BenchmarkNewConnect(b *testing.B) {
	var (
		alive   = uint16(30)
		cid     = "macy"
		user    = "john.doe"
		pwd     = []byte("secret")
		sExpiry = uint32(30)
	)
	b.Run("our", func(b *testing.B) {

		for n := 0; n < b.N; n++ {

			// our packet
			our := NewConnect()
			our.SetKeepAlive(alive)
			our.SetClientID(cid)
			our.SetUsername(user)
			our.SetPassword(pwd)
			our.SetSessionExpiryInterval(sExpiry)
		}
	})
	b.Run("their", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			their := packets.NewControlPacket(packets.CONNECT)
			c := their.Content.(*packets.Connect)
			c.KeepAlive = alive
			c.ClientID = cid
			c.UsernameFlag = true
			c.Username = user
			c.PasswordFlag = true
			c.Password = pwd
			c.Properties.SessionExpiryInterval = &sExpiry
		}
	})

}

func BenchmarkConnect_WriteTo(b *testing.B) {

	var our *Connect
	var their *packets.ControlPacket

	var (
		alive   = uint16(30)
		cid     = "macy"
		user    = "john.doe"
		pwd     = []byte("secret")
		sExpiry = uint32(30)
	)

	// our packet
	our = NewConnect()
	our.SetKeepAlive(alive)
	our.SetClientID(cid)
	our.SetUsername(user)
	our.SetPassword(pwd)
	our.SetSessionExpiryInterval(sExpiry)

	// their packet
	their = packets.NewControlPacket(packets.CONNECT)
	c := their.Content.(*packets.Connect)
	c.KeepAlive = alive
	c.ClientID = cid
	c.UsernameFlag = true
	c.Username = user
	c.PasswordFlag = true
	c.Password = pwd
	c.Properties.SessionExpiryInterval = &sExpiry

	b.Run("our", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			our.WriteTo(ioutil.Discard)
		}
	})

	b.Run("their", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			their.WriteTo(ioutil.Discard)
		}
	})
}
