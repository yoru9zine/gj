package execute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type processLogWriter struct {
	out io.WriteCloser
	enc *json.Encoder
	err error
}

func newProcessLogWriter(out io.WriteCloser) (*processLogWriter, error) {
	l := &processLogWriter{}
	l.out = out
	l.enc = json.NewEncoder(out)
	return l, nil
}

func (w *processLogWriter) Close() error {
	l := logline{EOF: true}
	for _, t := range []string{"stdout", "stderr", "stdin"} {
		l.Type = t
		err := w.enc.Encode(l)
		if err != nil {
			return err
		}
	}
	return w.out.Close()
}

func (w *processLogWriter) Write(line []byte, logtype string) error {
	return w.enc.Encode(&logline{Type: logtype, Data: line})
}

type logline struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
	EOF  bool   `json:"eof"`
}

//ProcessLogReader represents reader for process log
type ProcessLogReader struct {
	f      io.ReadCloser
	err    error
	Stdout *logBuffer
	Stderr *logBuffer
	Stdin  *logBuffer
}

//NewProcessLogReader returns new ProcessLogReader
func NewProcessLogReader(opt *ProcessOption) (*ProcessLogReader, error) {
	r := &ProcessLogReader{}
	f, err := opt.readCloser()
	if err != nil {
		return nil, err
	}
	r.f = f
	r.Stdout = newLogBuffer(128)
	r.Stderr = newLogBuffer(128)
	r.Stdin = newLogBuffer(128)
	return r, nil
}

// Start starts read process
func (r *ProcessLogReader) Start() {
	rr := bufio.NewReader(r.f)
	line := logline{}
	var (
		buf    bytes.Buffer
		closed int
	)
	for {
		if closed == 3 {
			break
		}
		l, isPrefix, err := rr.ReadLine()
		if err != nil {
			if err == io.EOF {
				continue
			}
			r.err = err
			return
		}
		buf.Write(l)
		if isPrefix {
			continue
		}
		b := buf.Bytes()
		buf.Reset()
		if err := json.Unmarshal(b, &line); err != nil {
			r.err = err
			return
		}
		var w *logBuffer
		switch line.Type {
		case "stdout":
			w = r.Stdout
		case "stderr":
			w = r.Stderr
		case "stdin":
			w = r.Stdin
		}
		w.Write(line.Data)
		if line.EOF {
			closed++
			w.Close()
		}
	}
}

// Close closes opened files
func (r *ProcessLogReader) Close() error {
	return r.f.Close()
}

type logBuffer struct {
	c chan []byte
}

func newLogBuffer(bufsize int) *logBuffer {
	return &logBuffer{c: make(chan []byte, bufsize)}
}

func (l *logBuffer) Read(p []byte) (int, error) {
	for l := range l.c {
		copy(p, l)
		return len(l), nil
	}
	return 0, io.EOF
}

func (l *logBuffer) Write(p []byte) (int, error) {
	l.c <- p
	return len(p), nil
}

func (l *logBuffer) Close() {
	close(l.c)
}
