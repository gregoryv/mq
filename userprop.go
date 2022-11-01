package mq

type UserProperties []property

// AddUserProp adds a user property. The User Property is allowed to
// appear multiple times to represent multiple name, value pairs. The
// same name is allowed to appear more than once.
func (p *UserProperties) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}

func (p *UserProperties) AddUserProperty(prop property) {
	*p = append(*p, prop)
}

func (p *UserProperties) properties(b []byte, i int) int {
	n := i
	for _, v := range *p {
		i += v.fillProp(b, i, UserProperty)
	}
	return i - n
}
