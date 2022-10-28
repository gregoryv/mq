package main

import (
	"context"
	"fmt"

	"github.com/gregoryv/cmdline"
)

type Sub struct {
	server string
}

func (s *Sub) ExtraOptions(cli *cmdline.Parser) {
	s.server = cli.Option("-s, --server").String("localhost:1883")
}

func (s *Sub) Run(ctx context.Context) error {
	return fmt.Errorf("sub: todo")
}
