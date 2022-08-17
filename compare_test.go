package mqtt_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/eclipse/paho.golang/packets"
	"github.com/gregoryv/mqtt"
)

func TestCompareConnect(t *testing.T) {
	our := mqtt.NewConnect()
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
	the.UsernameFlag = our.HasFlag(mqtt.UsernameFlag)
	the.Username = our.Username()

	our.SetPassword([]byte("123"))
	the.PasswordFlag = our.HasFlag(mqtt.PasswordFlag)
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

	our.SetWillRetain(true)
	the.WillRetain = our.HasFlag(mqtt.WillRetain)
	the.WillFlag = our.HasFlag(mqtt.WillFlag)

	our.SetWillTopic("topic/dead/clients")
	the.WillTopic = our.WillTopic()

	our.SetWillPayload([]byte(`{"clientID": "macy", "message": "died"`))
	the.WillMessage = our.WillPayload()

	// possible bug in Properties.Pack
	// our.SetWillContentType("application/json")
	the.WillProperties = &packets.Properties{}
	the.WillProperties.ContentType = "application/json" // never written

	our.AddWillProp("connected", "2022-01-01 14:44:32")
	the.WillProperties.User = append(the.WillProperties.User, packets.User{
		Key:   "connected",
		Value: "2022-01-01 14:44:32",
	})

	// our.SetPayloadFormat(true)
	// unsupported in paho

	our.SetWillQoS(2)
	the.WillQOS = our.WillQoS()

	our.SetCleanStart(true)
	the.CleanStart = our.HasFlag(mqtt.CleanStart)

	// dump the data
	var ourData, theirData bytes.Buffer
	our.WriteTo(&ourData)
	their.WriteTo(&theirData)

	a := hex.Dump(ourData.Bytes())
	b := hex.Dump(theirData.Bytes())

	if a != b {
		t.Logf("\n\n%s\n\nour %v bytes\n%s\n\n", our, ourData.Len(), a)
		t.Errorf("\n\ntheir %v bytes\n%s\n\n", theirData.Len(), b)
	}
}
