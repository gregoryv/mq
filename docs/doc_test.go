package docs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/eclipse/paho.golang/packets"
	"github.com/gregoryv/align"
	"github.com/gregoryv/draw/design"
	"github.com/gregoryv/draw/shape"
	"github.com/gregoryv/mq"
	. "github.com/gregoryv/web"
	"github.com/gregoryv/web/files"
	"github.com/gregoryv/web/theme"
	"github.com/gregoryv/web/toc"
)

func TestIndex(t *testing.T) {
	if err := NewIndex().SaveAs("index.html"); err != nil {
		t.Fatal(err)
	}
}

func NewIndex() *Page {
	nav := Nav(
		"Table of contents",
	)
	article := Article(
		H1("mqtt-v5 Packet Examples"),
		//NewPacketsDiagram().Inline(),

		nav,

		H2("mq.Publish"),
		docPacket(
			"doc_test.go", "examplePublish", examplePublish(),
		),

		H2("packets.Publish"),
		docPacket(
			"doc_test.go", "examplePublishTheir", examplePublishTheir(),
		),

		Pre(alignPackets(examplePublish(), examplePublishTheir())),
	)
	toc.MakeTOC(nav, article, "h2")

	return NewPage(
		Html(
			Head(
				Style(
					theme.GoldenSpace(),
					theme.GoishColors(),
					docTheme(),
				),
			),
			Body(
				article,
			),
		),
	)
}

func alignPackets(a mq.ControlPacket, b *packets.ControlPacket) string {
	var abuf bytes.Buffer
	a.WriteTo(&abuf)

	var bbuf bytes.Buffer
	b.WriteTo(&bbuf)

	result := align.NeedlemanWunsch(
		[]rune(
			hex.EncodeToString(abuf.Bytes()),
		),
		[]rune(
			hex.EncodeToString(bbuf.Bytes()),
		),
	)
	var buf bytes.Buffer
	result.PrintAlignment(&buf)
	return buf.String()
}

func docPacket(file, fn string, p io.WriterTo) *Element {
	example := files.MustLoadFunc(file, fn)
	var buf bytes.Buffer
	p.WriteTo(&buf)
	return Table(
		Tr(
			Td(
				// Content of the func, without signature and final return
				Code(Pre(stripFirstTab(sublines(example, 1, -2)))),
			),
			Td(
				Attr("rowspan", "2"),
				Pre(
					func() string {
						if p, ok := p.(fmt.Stringer); ok {
							return p.String()
						}
						return ""
					}(),

					Br(), Br(), func() string {
						lines := make([]string, buf.Len())
						for i, b := range buf.Bytes() {
							lines[i] = fmt.Sprintf("%2v %02x", i+1, b)
						}
						return strings.Join(lines, "\n")
					}()),
			),
		),
		Tr(
			Td(
				B("Dump"),
				Pre(func() string {
					if p, ok := p.(mq.ControlPacket); ok {
						var buf bytes.Buffer
						mq.Dump(&buf, p)
						return buf.String()
					}
					return "no dump"
				}()),
			),
		),
		Tr(
			Td(
				Attr("colspan", "2"),
				Pre(hex.Dump(buf.Bytes())),
			),
		),
	)
}

func docTheme() *CSS {
	css := NewCSS()
	css.Style("td",
		"vertical-align: top",
	)
	return css
}

var dropFirst = regexp.MustCompile("^\t")

func stripFirstTab(block string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = dropFirst.ReplaceAllString(line, "")
	}
	return strings.Join(lines, "\n")
}

func sublines(block string, fromStart, fromEnd int) string {
	lines := strings.Split(block, "\n")
	return strings.Join(lines[fromStart:len(lines)+fromEnd], "\n")
}

func examplePublish() *mq.Publish {
	p := mq.NewPublish()
	p.SetTopicName("gopher/pink")
	p.SetPayload([]byte("hug"))
	p.SetPayloadFormat(true) // utf-8
	return p
}

func examplePublishTheir() *packets.ControlPacket {
	p := packets.NewControlPacket(packets.PUBLISH)
	c := p.Content.(*packets.Publish)
	c.Topic = "gopher/pink"
	c.Properties = &packets.Properties{}
	pformat := byte(1)
	c.Properties.PayloadFormat = &pformat
	c.Payload = []byte("hug")
	return p
}

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
			connect, connack, auth, disconnect,
			pingreq, pingresp,
			publish, puback, pubrec, pubrel, pubcomp,
			subscribe, suback, unsubscribe, unsuback,
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

	// note with indexed packets
	var buf bytes.Buffer
	buf.WriteString("Control Packets\n\n")
	for i, p := range all {
		v := strings.ReplaceAll(p.Title, " struct", "")
		v = strings.ReplaceAll(v, "mq.", "")
		fmt.Fprintf(&buf, "%d. %s\n", i+1, v)
	}
	d.Note(strings.TrimSpace(buf.String())).At(750, 30)
	return d
}
