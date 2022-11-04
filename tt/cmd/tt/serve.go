package main

import (
	"context"
	"fmt"

	"github.com/gregoryv/cmdline"
)

type Serve struct {
	bind string
}

func (s *Serve) ExtraOptions(cli *cmdline.Parser) {
	s.bind = cli.Option("-b, --bind, $BIND").String("localhost:1883")
}

func (s *Serve) Run(ctx context.Context) error {
	return fmt.Errorf("serve: todo")
}
