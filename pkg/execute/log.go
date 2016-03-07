package execute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"
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
	m.Lock()
	defer m.Unlock()
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
	m.Lock()
	defer m.Unlock()
	w.enc.Encode(&logline{Type: logtype, Data: line})
}

/*
func (w *ProcessLogWriter) WriteOutput(buf []byte, r io.Reader, logtype string) {
	n, err := r.Read(buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		w.err = err
	}
	w.enc.Encode(&logline{Type: logtype, Data: buf[:n]})
}
*/
type logline struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
	EOF  bool   `json:"eof"`
}

type ProcessLogReader struct {
	f      *os.File
	err    error
	Stdout *LogReadWriter
	Stderr *LogReadWriter
	Stdin  *LogReadWriter
}

func NewProcessLogReader(opt *ProcessOption) (*ProcessLogReader, error) {
	r := &ProcessLogReader{}
	f, err := os.Open(opt.logFile())
	if err != nil {
		return nil, err
	}
	r.f = f
	r.Stdout = &LogReadWriter{}
	r.Stderr = &LogReadWriter{}
	r.Stdin = &LogReadWriter{}
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
		var w *LogReadWriter
		switch line.Type {
		case "stdout":
			w = r.Stdout
		case "stderr":
			w = r.Stderr
		case "stdin":
			w = r.Stdin
		}
		w.Write(line.Data)
		w.setEOF(line.EOF)
	}
}

func (r *ProcessLogReader) Close() error {
	return r.f.Close()
}

type LogReadWriter struct {
	buf bytes.Buffer
	eof bool
}

func (l *LogReadWriter) Read(p []byte) (int, error) {
	for {
		m.Lock()
		bufsize := l.buf.Len()
		if bufsize > 0 {
			defer m.Unlock()
			return l.buf.Read(p)
		}
		if l.eof {
			defer m.Unlock()
			return 0, io.EOF
		}
		m.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func (l *LogReadWriter) Write(p []byte) (int, error) {
	m.Lock()
	defer m.Unlock()
	return l.buf.Write(p)
}

func (l *LogReadWriter) setEOF(b bool) {
	m.Lock()
	defer m.Unlock()
	l.eof = b
}
