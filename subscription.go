package mq

func NewSubscription(filter string, h HandlerFunc) *Subscription {
	p := NewSubscribe()
	p.AddFilter(filter, 0)

	return &Subscription{
		Subscribe:   &p,
		HandlerFunc: h,
	}
}
