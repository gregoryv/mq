package mqtt

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func TestConnect(t *testing.T) {
	var (
		alive   = uint16(30)
		cid     = "macy"
		user    = "john.doe"
		pwd     = []byte("secret")
		sExpiry = uint32(30)
	)

	// our packet
	our := NewConnect()
	our.SetKeepAlive(alive)
	our.SetClientID(cid)
	our.SetUsername(user)
	our.SetPassword(pwd)
	our.SetSessionExpiryInterval(sExpiry)

	var ourData bytes.Buffer
	our.WriteTo(&ourData)

	// their packet
	their := packets.NewControlPacket(packets.CONNECT)
	c := their.Content.(*packets.Connect)
	c.KeepAlive = alive
	c.ClientID = cid
	c.UsernameFlag = true
	c.Username = user
	c.PasswordFlag = true
	c.Password = pwd
	c.Properties.SessionExpiryInterval = &sExpiry
	var theirData bytes.Buffer
	their.WriteTo(&theirData)

	a := hex.Dump(theirData.Bytes())
	b := hex.Dump(ourData.Bytes())

	if a != b {
		t.Logf("\n\ntheir %v bytes\n%s\n\n", theirData.Len(), a)
		t.Logf("\n\nour %v bytes\n%s\n\n", ourData.Len(), b)
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
	if f.Has(WillFlag) || !f.Has(Reserved) {
		t.Errorf("Has %08b", f)
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
