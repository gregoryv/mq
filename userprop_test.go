package mq

import (
	"fmt"
	"testing"
)

func ExampleUserProperties_AddUserProp() {
	var u UserProperties
	u.AddUserProp(
		"size", "large",
		"color", "red",
	)
	fmt.Println(u)
	// output:
	// [size:large color:red]
}

func TestUserProperties(t *testing.T) {
	var u UserProperties
	u.AddUserProp(
		"key", "val",
		"color", "red",
	)
	if len(u) != 2 {
		t.Error(u)
	}
}
