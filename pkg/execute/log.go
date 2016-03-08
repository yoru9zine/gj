package execute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
)

type ProcessLogWriter struct {
	f   *os.File
	enc *json.Encoder
	err error
}

func newProcessLogWriter(filePath string) (*ProcessLogWriter, error) {
	l := &ProcessLogWriter{}
	dir := path.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create `%s`: %s", dir, err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create ``")
	}
	l.f = f
	l.enc = json.NewEncoder(f)
	return l, nil
}

func (w *ProcessLogWriter) Close() error {
	l := logline{EOF: true}
	for _, t := range []string{"stdout", "stderr", "stdin"} {
		l.Type = t
		err := w.enc.Encode(l)
		if err != nil {
			return err
		}
	}
	return w.f.Close()
}

func (w *ProcessLogWriter) WriteOutput2(line []byte, logtype string) {
	w.enc.Encode(&logline{Type: logtype, Data: line})
}

type logline struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
	EOF  bool   `json:"eof"`
}

type ProcessLogReader struct {
	f      *os.File
	err    error
	Stdout *logBuffer
	Stderr *logBuffer
	Stdin  *logBuffer
}

func NewProcessLogReader(opt *ProcessOption) (*ProcessLogReader, error) {
	r := &ProcessLogReader{}
	f, err := os.Open(opt.logFile())
	if err != nil {
		return nil, err
	}
	r.f = f
	r.Stdout = newLogBuffer(128)
	r.Stderr = newLogBuffer(128)
	r.Stdin = newLogBuffer(128)
	return r, nil
}

func (r *ProcessLogReader) Start() {
	rr := bufio.NewReader(r.f)
	line := logline{}
	var buf bytes.Buffer
	for {
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
			w.Close()
		}
	}
}

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
