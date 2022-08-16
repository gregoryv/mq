package mqtt

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func TestConnect(t *testing.T) {
	//
	c := NewConnect()
	c.SetKeepAlive(299)
	c.SetClientID("macy")
	c.SetUsername("john.doe")
	c.SetPassword([]byte("123"))
	c.SetSessionExpiryInterval(30)
	c.AddUserProp("color", "red")
	c.SetAuthMethod("digest")
	c.SetAuthData([]byte("secret"))
	c.SetMaxPacketSize(4096)
	c.SetTopicAliasMax(128)
	c.SetRequestResponseInfo(true)
	c.SetRequestProblemInfo(true)
	c.SetResponseTopic("response/to/macy")
	c.SetCorrelationData([]byte("perhaps a uuid"))

	c.SetWillRetain(true)
	c.SetWillTopic("topic/dead/clients")
	c.SetWillPayload([]byte(`{"clientID": "macy"}`))
	c.SetWillContentType("application/json")
	c.SetWillDelayInterval(111)
	c.SetWillPayloadFormat(true)
	c.SetWillMessageExpiryInterval(100)
	c.AddWillProp("connected", "2022-01-01 14:44:32")

	c.SetUsername("") // unset toggles flag
	c.SetPassword(nil)

	var buf bytes.Buffer
	c.WriteTo(&buf)
	dump := hex.Dump(buf.Bytes())

	t.Logf("\n\n%s\n\n%s\n\n%v bytes\n\n", c, dump, buf.Len())
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
			our = NewConnect()
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
			their = packets.NewControlPacket(packets.CONNECT)
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
			our = NewConnect()
			our.SetKeepAlive(alive)
			our.SetClientID(cid)
			our.SetUsername(user)
			our.SetPassword(pwd)
			our.SetSessionExpiryInterval(sExpiry)
		}
	})
	b.Run("their", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			their = packets.NewControlPacket(packets.CONNECT)
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

var our *Connect
var their *packets.ControlPacket

func init() {
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
}
