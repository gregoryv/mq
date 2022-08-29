package mqtt

func NewSubscribe() Subscribe {
	return Subscribe{}
}

type Subscribe struct {
	fixed Bits
}
