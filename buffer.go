package mqtt

func NewBuffer(v wireType) *buffer {
	buf := &buffer{
		data: _LENGTH,
	}
	width := buf.fill(true, v)
	buf.data = make([]byte, width)
	buf.fill(true, v)

	return buf
}

// buffer is similar to bytes.Buffer but it's limited to a fixed byte
// slice and cannot grow. It's used for building wire encoded streams
// of data minimizing number of allocations.
type buffer struct {
	data  []byte
	index int
	err   error
}

func (b *buffer) fill(do bool, v ...wireType) int {
	if !do {
		return b.index
	}
	n := b.index
	for _, v := range v {
		b.index += v.fill(b.data, b.index)
	}
	return b.index - n
}
