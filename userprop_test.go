package mq

import "testing"

func TestUserProperties(t *testing.T) {
	var u UserProperties
	u.AddUserProp("key", "val")
	if len(u) != 1 {
		t.Error(u)
	}
}
