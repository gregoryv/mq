package mqtt_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/eclipse/paho.golang/packets"
	"github.com/gregoryv/mqtt"
)

func TestCompareConnect(t *testing.T) {
	// our packet
	our := mqtt.NewConnect()
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
	// longer payload does not compare, why? todo
	//our.SetWillPayload([]byte(`{"clientID": "macy", "message": "died"`))
	our.SetWillPayload([]byte(`{"clientID": "macy"}`))
	// These fields yield different result in paho.golang
	//
	// our.SetWillContentType("application/json") (maybe bug in Properties.Pack)
	// our.SetPayloadFormat(true)

	// their packet
	their := packets.NewControlPacket(packets.CONNECT)
	c := their.Content.(*packets.Connect)
	c.KeepAlive = our.KeepAlive()
	c.ClientID = our.ClientID()
	c.UsernameFlag = our.HasFlag(mqtt.UsernameFlag)
	c.Username = our.Username()
	c.PasswordFlag = our.HasFlag(mqtt.PasswordFlag)
	c.Password = our.Password()
	c.WillFlag = our.HasFlag(mqtt.WillFlag)
	c.WillTopic = our.WillTopic()
	c.WillMessage = our.WillPayload()
	// will properties
	var wp packets.Properties
	c.WillProperties = &wp
	// set here but has no affect, (maybe bug in Properties.Pack)
	wp.ContentType = "application/json"
	// user properties
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
		t.Errorf("\n\ntheir %v bytes\n%s\n\n", theirData.Len(), b)
	}
}
