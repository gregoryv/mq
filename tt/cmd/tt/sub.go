package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gregoryv/cmdline"
)

type Sub struct {
	server *url.URL
}

func (s *Sub) ExtraOptions(cli *cmdline.Parser) {
	s.server = cli.Option("-s, --server").Url("localhost:1883")
}

func (s *Sub) Run(ctx context.Context) error {
	return fmt.Errorf("sub: todo")
}
