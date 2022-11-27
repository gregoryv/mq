package mq

import (
	"fmt"
	"testing"
)

func ExampleTopicFilter_String() {
	fmt.Println(NewTopicFilter("gopher/pink/#", OptQoS1|OptRetain1))
	// output:
	// gopher/pink/# --r1---1
}

func TestTopicFilter(t *testing.T) {

	var f TopicFilter
	eq(t, f.SetFilter, f.Filter, "a/b")
	eq(t, f.SetOptions, f.Options, OptQoS1)

	cases := []struct {
		f   TopicFilter
		exp string
	}{
		{
			f:   NewTopicFilter("a/b", OptQoS3),
			exp: "a/b --r0--!!",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS1|OptRetain1),
			exp: "a/b --r1---1",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS2|OptRetain2),
			exp: "a/b --r2--2-",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS2|OptRetain3),
			exp: "a/b --!!--2-",
		},
	}

	for _, c := range cases {
		if v := c.f.String(); v != c.exp {
			t.Error("got", v, "expected", c.exp)
		}
	}
}
