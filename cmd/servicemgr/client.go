package main

import (
	"io"
	"sync/atomic"
)

var id uint32

type Client struct {
	id uint32
	io.ReadWriteCloser
}

func NewClient(rwc io.ReadWriteCloser) *Client {
	return &Client{
		id:              atomic.AddUint32(&id, 1),
		ReadWriteCloser: rwc,
	}
}

func (c *Client) Id() uint32 {
	return c.id
}
