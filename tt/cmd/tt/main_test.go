package main

import (
	"testing"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/cmdline/clitest"
)

func Test_main(t *testing.T) {
	cmdline.DefaultShell = clitest.NewShellT("test", "-h")
	main()
}
