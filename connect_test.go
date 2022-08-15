package mqtt

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"
	"unsafe"

	"github.com/eclipse/paho.golang/packets"
)

func TestCompareConnect(t *testing.T) {
	// our packet
	our := NewConnect()
	our.SetKeepAlive(299)
	our.SetClientID("macy")
	our.SetUsername("john.doe")
	our.SetPassword([]byte("123"))
	our.SetSessionExpiryInterval(30)
	our.AddUserProp("color", "red")
	our.SetAuthMethod("digest")
	our.SetAuthData([]byte("secret"))
	our.SetWillFlag(true) // would be nice not to have to think about this one
	our.SetWillTopic("topic/dead/clients")
	our.SetWillPayload([]byte("goodbye"))
	// These fields yield different result in paho.golang
	//
	// our.SetWillContentType("application/json") (maybe bug in Properties.Pack)
	// our.SetPayloadFormat(true)

	// their packet
	their := packets.NewControlPacket(packets.CONNECT)
	c := their.Content.(*packets.Connect)
	c.KeepAlive = our.KeepAlive()
	c.ClientID = our.ClientID()
	c.UsernameFlag = true
	c.Username = our.Username()
	c.PasswordFlag = true
	c.Password = our.Password()
	c.WillFlag = true
	c.WillTopic = "topic/dead/clients"
	c.WillMessage = []byte("goodbye")

	var wp packets.Properties // will properties
	c.WillProperties = &wp
	// set here but has no affect, (bug in Properties.Pack)
	wp.ContentType = "application/json"

	p := c.Properties
	var se uint32 = 30
	p.SessionExpiryInterval = &se
	p.User = append(p.User, packets.User{"color", "red"})
	p.AuthMethod = "digest"
	p.AuthData = []byte("secret")

	// dump the data
	var ourData, theirData bytes.Buffer
	our.WriteTo(&ourData)
	their.WriteTo(&theirData)

	a := hex.Dump(ourData.Bytes())
	b := hex.Dump(theirData.Bytes())

	if a != b {
		t.Logf("\n\nour %v bytes\n%s\n\n", ourData.Len(), a)
		t.Logf("\n\ntheir %v bytes\n%s\n\n", theirData.Len(), b)
	} else {
		t.Logf("their size of %T %v bytes", their, unsafe.Sizeof(their))
		t.Logf("our size of %T %v bytes", our, unsafe.Sizeof(our))
		t.Logf("\n\n%s\n\n%s\n\n%v bytes\n\n", our, a, ourData.Len())
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
