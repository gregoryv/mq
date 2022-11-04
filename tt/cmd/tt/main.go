package main

import (
	"context"

	"github.com/gregoryv/cmdline"
)

func main() {
	var (
		cli = cmdline.NewBasicParser()
		// shared options

		// sub commands
		commands = cli.Group("Commands", "COMMAND")

		_ = commands.New("pub", &Pub{})
		_ = commands.New("sub", &Sub{})
		_ = commands.New("serve", &Serve{})

		cmd = commands.Selected()
	)
	u := cli.Usage()
	u.Preface(
		"mqtt-v5 server and client by Gregory Vinčić",
	)
	cli.Parse()

	sh := cmdline.DefaultShell
	if err := cmd.(Command).Run(context.Background()); err != nil {
		sh.Fatal(err)
	}
	sh.Exit(0)
}

type Command interface {
	Run(context.Context) error
}
