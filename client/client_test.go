package client

import (
	"testing"

	"github.com/gregoryv/mqtt"
)

func TestClient(t *testing.T) {
	c := mqtt.NewConnect()
	c.Flags().Has(mqtt.CleanStart)
	// above needs knowledge of the flags and the protocol
	//
	// nicer, not having to know how it's stored and where
	// c.IsCleanStart()
}
