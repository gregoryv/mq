package mq

import "testing"

func Test_ReasonStringCode(t *testing.T) {
	for i := 0; i < 0xFF; i++ {
		if v := ReasonCode(i).String(); v == "" {
			t.Error("ReasonCode", i, "empty")
		}
	}
}
