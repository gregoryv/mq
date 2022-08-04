package mqtt

type FixedHeader []byte

func (h FixedHeader) Name() string {
	return controlPacketTypeName[byte(h[0])&0b1111_0000]
}

func (h FixedHeader) Value() byte {
	return byte(h[0]) & 0b1111_0000
}

func (h FixedHeader) Flags() byte {
	return byte(h[0]) & 0b0000_1111
}
