package main

import "io"

type combinedIO struct {
	r io.Reader
	w io.Writer
}

func (c *combinedIO) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *combinedIO) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}
