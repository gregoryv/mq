package main

import (
	"context"
	"fmt"

	"github.com/gregoryv/cmdline"
)

type Pub struct {
	server string
}

func (p *Pub) ExtraOptions(cli *cmdline.Parser) {
	p.server = cli.Option("-s, --server").String("localhost:1883")
}

func (p *Pub) Run(ctx context.Context) error {
	return fmt.Errorf("pub: todo")
}
