package main

import "io"

func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{Reader: r, Writer: w}
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}

func (c *Conn) Close() error {
	if v, ok := c.Reader.(io.Closer); ok {
		if err := v.Close(); err != nil {
			return err
		}
	}
	if v, ok := c.Writer.(io.Closer); ok {
		if err := v.Close(); err != nil {
			return err
		}
	}
	return nil
}
