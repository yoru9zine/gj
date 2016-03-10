package gj

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/kr/pty"
)

type Command struct {
	Name string
	Args []string
}

func (c *Command) Start(opt *CommandOption) (io.Writer, *exec.Cmd, error) {
	cmd := exec.Command(c.Name, c.Args...)
	f, err := pty.Start(cmd)
	cl := newCommandLogger(opt.LogWriter)
	cl.setStdout(f)
	cl.setStderr(nil)
	return f, cmd, err
}

type CommandOption struct {
	PTY       bool
	LogWriter io.Writer
}

type commandLog struct {
	Type string `json:"type"`
	Data []byte `json:"data,omitempty"`
	EOF  bool   `json:"eof,omitempty"`
}
type commandLogger struct {
	log     io.Writer
	encoder *json.Encoder
	readers map[string]io.Reader
	errors  []error
}

func newCommandLogger(w io.Writer) *commandLogger {
	cl := &commandLogger{
		log:     w,
		encoder: json.NewEncoder(w),
		readers: make(map[string]io.Reader, 3),
	}
	return cl
}

func (cl *commandLogger) startReader(r io.Reader, logtype string) {
	p := make([]byte, 1024)
	var l *commandLog
	for {
		if r != nil {
			n, err := r.Read(p)
			if err != nil {
				if err != io.EOF {
					cl.errors = append(cl.errors, fmt.Errorf("read error at %s: %s", logtype, err))
					break
				}
				l = &commandLog{Type: logtype, EOF: true}
			} else {
				l = &commandLog{Type: logtype, Data: p[:n]}
			}
		} else {
			l = &commandLog{Type: logtype, EOF: true}
		}
		if err := cl.encoder.Encode(l); err != nil {
			cl.errors = append(cl.errors, fmt.Errorf("encode error at %s: %s, `%+v`", logtype, err, l))
			break
		}
		if l.EOF {
			break
		}
	}

}
func (cl *commandLogger) setStdout(r io.Reader) {
	go cl.startReader(r, "stdout")
}

func (cl *commandLogger) setStderr(r io.Reader) {
	go cl.startReader(r, "stderr")
}

type commandLogReader struct {
	r      io.Reader
	Stdout io.Reader
	Stderr io.Reader
	err    error
}

func newCommandLogReader(r io.Reader) *commandLogReader {
	cl := &commandLogReader{
		r:      r,
		Stdout: newChanReader(),
		Stderr: newChanReader(),
	}
	go cl.readLoop()
	return cl
}

func (c *commandLogReader) readLoop() {
	buf := &bytes.Buffer{}
	br := bufio.NewReader(c.r)
	for {
		var l commandLog
		line, prefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			c.err = err
			break
		}
		buf.Write(line)
		if prefix {
			continue
		}
		if err := json.Unmarshal(buf.Bytes(), &l); err != nil {
			c.err = err
			break
		}
		buf.Reset()
		c.dispatch(&l)
	}
}

func (c *commandLogReader) dispatch(l *commandLog) {
	var r *chanReader
	switch l.Type {
	case "stdout":
		r = c.Stdout.(*chanReader)
	case "stderr":
		r = c.Stderr.(*chanReader)
	}
	if l.EOF {
		close(r.Chan)
	} else {
		r.Chan <- l.Data
	}
}

type chanReader struct {
	Chan chan []byte
}

func newChanReader() *chanReader {
	return &chanReader{
		Chan: make(chan []byte),
	}
}

func (c *chanReader) Read(p []byte) (int, error) {
	b, ok := <-c.Chan
	if !ok {
		return 0, io.EOF
	}
	copy(p, b)
	return len(b), nil
}
