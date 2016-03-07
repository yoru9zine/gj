package execute

import "io"

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
