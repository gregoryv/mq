package main

import (
	"context"
	"log"

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
		_ = commands.New("serve", &ServeCmd{})

		cmd = commands.Selected()
	)
	u := cli.Usage()
	u.Preface(
		"mqtt-v5 server and client by Gregory Vinčić",
	)
	cli.Parse()

	sh := cmdline.DefaultShell
	if cmd == nil {
		// this shouldn't happen, default should be the first one. When testing it's empty
		log.Println("empty command")
		return
	}

	if err := cmd.(Command).Run(context.Background()); err != nil {
		sh.Fatal(err)
	}
	sh.Exit(0)
}

type Command interface {
	Run(context.Context) error
}
