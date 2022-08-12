package mqtt

import (
	"bytes"
	"io"

	"github.com/gregoryv/nexus"
)

// If we want to be able to handle large packets each must implement
// io.ReaderFrom This allows a client decide if it should read in all
// the data in one slice and wrap it in a reader or not.

// The other direction is also important to be able to write out large
// packets without loading everything into memory each packet must
// implement io.WriterTo.

// NewConnect returns an empty MQTT v5 connect packet.
func NewConnect() *Connect {
	return &Connect{
		fixed:           CONNECT,
		protocolName:    "MQTT",
		protocolVersion: 5,
	}
}

type Connect struct {
	// fields are ordered to minimize memory allocation
	fixed byte // 1
	flags byte // 1

	protocolVersion uint8 // 1
	protocolName    string

	payload *limitedReader
}

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	p, err := nexus.NewPrinter(w)

	// variable header
	p.Write([]byte{c.fixed})
	vbint(c.width()).WriteTo(p)
	u8str(c.protocolName).WriteTo(p)
	p.Write([]byte{c.protocolVersion, c.flags})

	c.payload.WriteTo(p)

	return p.Written, *err
}

// width returns the remaining length
func (p *Connect) width() int {
	n := 10 // always there
	n += 0  // todo width of properties

	if p.payload != nil {
		n += p.payload.width
	}
	return n
}

func (p *Connect) String() string {
	return Fixed(p.fixed).String()
}

// ---------------------------------------------------------------------
// 3.1.2.3 Connect Flags
// ---------------------------------------------------------------------

const (
	Reserved byte = 1 << iota
	CleanStart
	WillFlag
	WillQoS1
	WillQoS2
	WillRetain
	PasswordFlag
	UsernameFlag
)

type ConnectFlags byte

// String returns flags represented with a letter.
// Improper flags are marked with '!' and unset are marked with '-'.
//
//   UsernameFlag  u
//   PasswordFlag  p
//   WillRetain    r
//   WillQoS       1, 2 or !
//   WillFlag      2
//   CleanStart    s
//   Reserved      !
func (c ConnectFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 7)

	mark := func(i int, flag byte, v byte) {
		if !c.Has(flag) {
			return
		}
		flags[i] = v
	}
	mark(0, UsernameFlag, 'u')
	mark(1, PasswordFlag, 'p')
	mark(2, WillRetain, 'r')
	mark(3, WillQoS1, '1')
	mark(3, WillQoS2, '2')
	mark(3, WillQoS1|WillQoS2, '!')
	mark(4, WillFlag, 'w')
	mark(5, CleanStart, 's')
	mark(6, Reserved, '!')

	return string(flags)
}

func (c ConnectFlags) Has(f byte) bool { return bits(c).Has(f) }

// limitedReader is a reader with a known size. This is needed to
// calculate the remaining length of a control packet without loading
// everything into memory.
type limitedReader struct {
	src io.ReadSeeker

	// width is the number of bytes the above reader will ever read
	// before returning EOF. Similar to io.LimitedReader, though it's
	// not updated after each read.
	width int
}
