package main

import (
	"testing"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/cmdline/clitest"
)

func Test_main(t *testing.T) {

	for _, cmd := range []string{"", "pub", "sub", "serve"} {
		cmdline.DefaultShell = clitest.NewShellT("test", cmd, "-h")
		main()
	}
}
