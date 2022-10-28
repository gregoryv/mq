package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gregoryv/cmdline"
)

type Serve struct {
	bind *url.URL
}

func (s *Serve) ExtraOptions(cli *cmdline.Parser) {
	s.bind = cli.Option("-b, --bind").Url(":1883")
}

func (s *Serve) Run(ctx context.Context) error {
	return fmt.Errorf("serve: todo")
}
