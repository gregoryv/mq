package mq

import (
	"bytes"
	"fmt"
)

func NewTopicFilter(filter string, options Opt) TopicFilter {
	return TopicFilter{
		filter:  wstring(filter),
		options: bits(options),
	}
}

type TopicFilter struct {
	filter  wstring
	options bits
}

func (c *TopicFilter) SetFilter(v string) { c.filter = wstring(v) }
func (c *TopicFilter) Filter() string     { return string(c.filter) }

func (c *TopicFilter) SetOptions(v Opt) { c.options = bits(v) }
func (c *TopicFilter) Options() Opt     { return Opt(c.options) }

func (c TopicFilter) fill(b []byte, i int) int {
	n := i
	i += c.filter.fill(b, i)
	i += c.options.fill(b, i)
	return i - n
}

func (c TopicFilter) String() string {
	flags := bytes.Repeat([]byte("-"), 8)

	mark := func(i int, flag byte, v byte) {
		if !c.options.Has(flag) {
			return
		}
		flags[i] = v
	}

	// QoS
	mark(7, byte(OptQoS1), '1')
	mark(6, byte(OptQoS2), '2')
	if c.options.Has(byte(OptQoS3)) {
		flags[7] = '!'
		flags[6] = '!'
	}
	if c.options.Has(byte(OptNL)) {
		flags[5] = 'n'
	}
	if c.options.Has(byte(OptRAP)) {
		flags[4] = 'p'
	}
	// Retain
	flags[3] = '0'
	flags[2] = 'r'
	if c.options.Has(byte(OptRetain1)) {
		flags[3] = '1'
		flags[2] = 'r'
	}
	if c.options.Has(byte(OptRetain2)) {
		flags[3] = '2'
		flags[2] = 'r'
	}
	if c.options.Has(byte(OptRetain3)) {
		flags[3] = '!'
		flags[2] = '!'
	}

	// Reserved
	mark(1, 1<<6, '!')
	mark(0, 1<<7, '!')

	return fmt.Sprintf("%s %s", c.filter, string(flags))
}
