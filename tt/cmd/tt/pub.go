package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gregoryv/cmdline"
)

type Pub struct {
	server *url.URL
}

func (p *Pub) ExtraOptions(cli *cmdline.Parser) {
	p.server = cli.Option("-s, --server").Url("localhost:1883")
}

func (p *Pub) Run(ctx context.Context) error {
	return fmt.Errorf("pub: todo")
}
