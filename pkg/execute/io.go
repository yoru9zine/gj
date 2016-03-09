package execute

import (
	"io"
	"time"
)

type multiIO struct {
	ioObjects []interface{}
}

func (m multiIO) Read(p []byte) (int, error) {
	var (
		n   int
		err error
	)
	for _, o := range m.ioObjects {
		if r, ok := o.(io.Reader); ok {
			n, err = r.Read(p)
		}
	}
	return n, err
}

func (m multiIO) Write(p []byte) (int, error) {
	var (
		n   int
		err error
	)
	for _, o := range m.ioObjects {
		if w, ok := o.(io.Writer); ok {
			n, err = w.Write(p)
		}
	}
	return n, err
}
func (m multiIO) Close() error {
	var err error
	for _, o := range m.ioObjects {
		if c, ok := o.(io.Closer); ok {
			err = c.Close()
		}
	}
	return err
}

type reader2chan struct {
	reader  io.Reader
	Channel chan []byte
	fin     chan struct{}
}

func newReader2Chan(r io.Reader) *reader2chan {
	return &reader2chan{
		reader:  r,
		Channel: make(chan []byte),
		fin:     make(chan struct{}),
	}
}

func (r *reader2chan) Start() {
	for {
		select {
		case <-r.fin:
			close(r.Channel)
			return
		case <-time.After(100 * time.Millisecond):
			buf := make([]byte, 1024)
			n, err := r.reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					continue
				}
				close(r.Channel)
				return
			}
			r.Channel <- buf[:n]
		}
	}
}

func (r *reader2chan) Stop() {
	r.fin <- struct{}{}
}
