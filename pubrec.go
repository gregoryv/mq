package mqtt

func NewPubRec() *PubRec {
	return &PubRec{}
}

type PubRec struct {
	fixed Bits
}
