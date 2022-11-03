package mq

type UserProperties []UserProp

// AddUserProp adds key value pair user properties. The same key is is
// allowed to appear more than once.
func (p *UserProperties) AddUserProp(kvPair ...string) {
	for i := 0; i < len(kvPair); i += 2 {
		p.AddUserProperty(UserProp{kvPair[i], kvPair[i+1]})
	}
}

func (p *UserProperties) AddUserProperty(prop UserProp) {
	*p = append(*p, prop)
}

func (p *UserProperties) properties(b []byte, i int) int {
	n := i
	for _, v := range *p {
		i += v.fillProp(b, i, UserProperty)
	}
	return i - n
}
