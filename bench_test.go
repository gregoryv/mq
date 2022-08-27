package mqtt

import (
	"bytes"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func BenchmarkPublish(b *testing.B) {
	var (
		our   Publish
		their *packets.ControlPacket
	)

	b.Run("create", func(b *testing.B) {
		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				our = NewPublish()
				_ = our // todo fill out with reasonable values
			}

		})
		b.Run("their", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				their = packets.NewControlPacket(packets.PUBLISH)
				c := their.Content.(*packets.Publish)
				_ = c // todo fill out with reasonable values
			}
		})
	})
}

func BenchmarkConnect(b *testing.B) {
	var (
		alive   = uint16(30)
		cid     = "macy"
		user    = "john.doe"
		pwd     = []byte("secret")
		sExpiry = uint32(30)

		our   Connect
		their *packets.ControlPacket
	)

	b.Run("create", func(b *testing.B) {
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
	})

	// this buf is used in the next Unmarshal, our output is used in
	// both as input
	var buf bytes.Buffer
	b.Run("write", func(b *testing.B) {
		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				buf.Reset()
				our.WriteTo(&buf)
			}
		})
		b.Run("their", func(b *testing.B) {
			var buf bytes.Buffer
			for n := 0; n < b.N; n++ {
				buf.Reset()
				their.WriteTo(&buf) // to be similar to our
			}
		})
	})

	b.Run("read", func(b *testing.B) {
		var fh FixedHeader
		fh.ReadFrom(&buf)

		data := make([]byte, buf.Len())
		copy(data, buf.Bytes())

		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if _, err := fh.ReadPacket(&buf); err != nil {
					b.Fatal(err)
				}
				buf.Write(data)
			}
		})

		var (
			their = packets.NewControlPacket(packets.CONNECT)
			the   = their.Content.(*packets.Connect)
		)
		b.Run("their", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if err := the.Unpack(&buf); err != nil {
					b.Fatal(err)
				}
				buf.Write(data)
			}
		})
	})
}
