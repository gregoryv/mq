package tt

import "fmt"

var (
	// Client.IO is closed
	ErrNoConnection = fmt.Errorf("no connection")

	// Client has no receiver configured
	ErrUnsetReceiver = fmt.Errorf("unset receiver")

	// Client is not fully operational, ie. hasn't been started
	ErrNotRunning = fmt.Errorf("not running")

	// Settings cannot be changed
	ErrReadOnly = fmt.Errorf("read only")
)
