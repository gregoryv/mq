package mq

func NewSubscription(filter string, r Receiver) *Subscription {
	p := NewSubscribe()
	p.AddFilter(filter, 0)

	return &Subscription{
		Subscribe: &p,
		Receiver:  r,
	}
}
