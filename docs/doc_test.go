package docs

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gregoryv/draw/design"
	"github.com/gregoryv/draw/shape"
	"github.com/gregoryv/mq"
)

func TestDesignDiagram(t *testing.T) {
	NewPacketsDiagram().SaveAs("packets_diagram.svg")
}

func NewPacketsDiagram() *design.ClassDiagram {
	var (
		d      = design.NewClassDiagram()
		packet = d.Interface((*mq.Packet)(nil))

		connect     = d.Struct(mq.Connect{})
		connack     = d.Struct(mq.ConnAck{})
		auth        = d.Struct(mq.Auth{})
		disconnect  = d.Struct(mq.Disconnect{})
		publish     = d.Struct(mq.Publish{})
		puback      = d.Struct(mq.PubAck{})
		pubrec      = d.Struct(mq.PubRec{})
		pubrel      = d.Struct(mq.PubRel{})
		pubcomp     = d.Struct(mq.PubComp{})
		pingreq     = d.Struct(mq.PingReq{})
		pingresp    = d.Struct(mq.PingResp{})
		subscribe   = d.Struct(mq.Subscribe{})
		unsubscribe = d.Struct(mq.Unsubscribe{})
		unsuback    = d.Struct(mq.UnsubAck{})
		suback      = d.Struct(mq.SubAck{})

		all = []design.VRecord{
			connect, connack, auth, publish, disconnect,
			pingreq, pingresp,
			subscribe, suback, unsubscribe, unsuback,
			puback, pubrec, pubrel, pubcomp,
		}
	)
	d.Style.Spacing = 40

	d.HideRealizations()
	for _, p := range all {
		p.HideMethods()
	}
	d.Place(packet).At(240, 200)

	// connect cluster
	d.Place(connect).Above(packet).Move(-200, -40)
	d.Place(connack, auth, disconnect).RightOf(connect, 20)

	// publish cluster
	d.Place(
		publish, puback, pubrec, pubrel,
	).Below(connect, 20).Move(0, 110)
	shape.Move(pubrel, 30, 0)

	d.Place(pubcomp).RightOf(pubrel, 20)

	// ping cluster
	d.Place(pingreq, pingresp).Below(disconnect, 20).Move(120, 40)

	// subscribe cluster
	d.Place(subscribe, suback, unsubscribe).Below(pingresp, 20).Move(30, 40)
	d.Place(unsuback).LeftOf(unsubscribe, 20)
	shape.Move(suback, 30, 0)

	var buf bytes.Buffer
	for i, p := range all {
		v := strings.ReplaceAll(p.Title, " struct", "")
		v = strings.ReplaceAll(v, "mq.", "")
		fmt.Fprintf(&buf, "%d. %s\n", i+1, v)
	}
	d.Note(strings.TrimSpace(buf.String())).At(750, 30)
	return d
}
