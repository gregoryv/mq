package mq

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/eclipse/paho.golang/packets"
	"github.com/gregoryv/asserter"
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
	p.SetWill(Pub(1, "client/gone", "pink"), 3)

	fmt.Println(p.String())
	fmt.Println(DocumentFlags(p))
	// output:
	// CONNECT ---- up--1w-- MQTT5 pink 0s 58 bytes
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
	c.AddUserProp("color", "red")

	if got := c.String(); !strings.Contains(got, "CONNECT") {
		t.Error(got)
	}
	if v, _ := c.Will(); v != nil {
		t.Error("no will was set but got", v)
	}
	testControlPacket(t, c)

	c.SetWill(Pub(1, "client/gone", "pink"), 3)
	testControlPacket(t, c)

	// clears it
	if c.SetUsername(""); c.HasFlag(UsernameFlag) {
		t.Error("username flag still set")
	}
	if c.SetPassword(nil); c.HasFlag(PasswordFlag) {
		t.Error("password flag still set")
	}
	testControlPacket(t, c)
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

func TestDump_connect(t *testing.T) {
	c := NewConnect()
	c.SetClientID("macy")
	c.SetKeepAlive(299)
	c.SetUsername("john.doe")
	c.SetPassword([]byte("secret"))
	c.SetWill(Pub(0, "client/gone", "macy"), 300)
	c.AddUserProp("color", "red")

	var buf bytes.Buffer
	Dump(&buf, c)

	if v := buf.String(); strings.Contains(v, "john.doe") {
		t.Error("username not masked")
	}
	if v := buf.String(); strings.Contains(v, "secret") {
		t.Error("password not masked")
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

func TestCompareConnect(t *testing.T) {
	our := NewConnect()
	// theirs is divided into a wrapping ControlPacket and content
	their := packets.NewControlPacket(packets.CONNECT)
	the := their.Content.(*packets.Connect)

	our.SetKeepAlive(299)
	the.KeepAlive = our.KeepAlive()

	our.SetClientID("macy")
	the.ClientID = our.ClientID()

	var se uint32 = 30
	our.SetSessionExpiryInterval(se)
	the.Properties.SessionExpiryInterval = &se

	// Username and password
	our.SetUsername("john.doe")
	the.UsernameFlag = our.HasFlag(UsernameFlag)
	the.Username = our.Username()

	our.SetPassword([]byte("123"))
	the.PasswordFlag = our.HasFlag(PasswordFlag)
	the.Password = our.Password()

	// Authentication method and data
	our.SetAuthMethod("digest")
	the.Properties.AuthMethod = "digest"

	our.SetAuthData([]byte("secret"))
	the.Properties.AuthData = []byte("secret")

	// User properties
	our.AddUserProp("color", "red")
	the.Properties.User = append(
		the.Properties.User, packets.User{"color", "red"},
	)

	// Receive maximum
	our.SetReceiveMax(9)
	rm := our.ReceiveMax()
	the.Properties.ReceiveMaximum = &rm

	{
		p := NewPublish()
		p.SetRetain(true)
		p.SetTopicName("topic/dead/clients")
		p.SetPayload([]byte(`{"clientID": "macy", "message": "died"`))
		p.SetQoS(2)
		p.AddUserProp("connected", "2022-01-01 14:44:32")
		//p.SetCorrelationData([]byte("11-22-33")) doesn't work in paho
		our.SetWill(p, 3)
	}
	will, wExp := our.Will()
	the.WillRetain = will.Retain()
	the.WillFlag = our.HasFlag(WillFlag)
	the.WillTopic = will.TopicName()
	the.WillMessage = will.Payload()

	// possible bug in Properties.Pack
	// our.SetWillContentType("application/json")
	the.WillProperties = &packets.Properties{}
	the.WillProperties.ContentType = "application/json" // never written
	the.WillProperties.User = append(the.WillProperties.User, packets.User{
		Key:   "connected",
		Value: "2022-01-01 14:44:32",
	})
	the.WillProperties.WillDelayInterval = &wExp

	// possible bug in Properties.Pack
	// the.WillProperties.CorrelationData = []byte("11-22-33")

	// our.SetPayloadFormat(true)
	// unsupported in paho
	the.WillQOS = will.QoS()

	our.SetCleanStart(true)
	the.CleanStart = our.HasFlag(CleanStart)

	// write our theirs
	var buf bytes.Buffer
	their.WriteTo(&buf)

	got, err := ReadPacket(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, &our) {
		t.Log("our", our.String())
		t.Log("got", got.String())

		a := clean(got)
		b := clean(our)
		assert := asserter.New(t)
		assert().Equals(a, b)
	}
}

var dropRefs = regexp.MustCompile(`0x([0-9|a-z]+)`)

func clean(in any) string {
	v := fmt.Sprintf("%#v", in)
	v = strings.ReplaceAll(v, ", ", ",\n")
	v = dropRefs.ReplaceAllString(v, "0x")
	return v
}

func compare(t *testing.T, our, their io.WriterTo) {
	t.Helper()
	// dump the data
	var ourData, theirData bytes.Buffer
	our.WriteTo(&ourData)
	their.WriteTo(&theirData)

	a := hex.Dump(ourData.Bytes())
	b := hex.Dump(theirData.Bytes())

	f := theirData.Bytes()[0]
	if a != b {
		t.Logf("\n\n%s\n\nour %v bytes\n%s\n\n", our, ourData.Len(), a)
		t.Errorf("\n\n%s %08b\n\ntheir %v bytes\n%s\n\n",
			firstByte(f), f,
			theirData.Len(), b)
	}
}

func BenchmarkConnect(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewConnect()
			p.SetKeepAlive(30)
			p.SetClientID("macy")
			p.SetUsername("john.doe")
			p.SetPassword([]byte("secret"))
			p.SetSessionExpiryInterval(30)
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.CONNECT)
			c := p.Content.(*packets.Connect)
			c.KeepAlive = 30
			c.ClientID = "macy"
			c.UsernameFlag = true
			c.Username = "john.doe"
			c.PasswordFlag = true
			c.Password = []byte("secret")
			sExpiry := uint32(30)
			c.Properties.SessionExpiryInterval = &sExpiry
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}

func BenchmarkConnectWill(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewConnect()
			p.SetKeepAlive(30)
			p.SetClientID("macy")
			p.SetUsername("john.doe")
			p.SetPassword([]byte("secret"))
			p.SetSessionExpiryInterval(30)

			w := NewPublish()
			w.SetRetain(true)
			w.SetQoS(1)
			w.SetTopicName("topic/name")
			w.SetMessageExpiryInterval(199)
			w.SetPayloadFormat(true)
			w.SetCorrelationData([]byte("corr"))
			w.AddUserProp("color", "red")
			w.SetContentType("text/plain")
			w.SetPayload([]byte("gopher"))
			p.SetWill(w, 5)

			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.CONNECT)
			c := p.Content.(*packets.Connect)
			c.KeepAlive = 30
			c.ClientID = "macy"
			c.UsernameFlag = true
			c.Username = "john.doe"
			c.PasswordFlag = true
			c.Password = []byte("secret")
			sExpiry := uint32(30)
			c.Properties.SessionExpiryInterval = &sExpiry
			c.WillFlag = true
			c.WillRetain = true
			c.WillQOS = 1
			c.WillTopic = "topic/name"
			c.WillMessage = []byte("gopher")
			wExpiry := uint32(199)
			pFormat := byte(1)
			c.WillProperties = &packets.Properties{
				MessageExpiry:   &wExpiry,
				PayloadFormat:   &pFormat,
				CorrelationData: []byte("corr"),
				ContentType:     "text/plain",
				User: []packets.User{
					{Key: "color", Value: "red"},
				},
			}

			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
