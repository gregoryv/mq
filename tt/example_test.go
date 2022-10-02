package tt

import (
	"bytes"
	"context"
)

func Example_newClient() {
	c := NewClient()

	var buf bytes.Buffer // network connection substitute
	c.SetIO(&buf)

	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)

	// use the transceiver...

	// and finally stop it
	cancel()
}
