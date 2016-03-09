package execute

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"reflect"
	"testing"
)

type tmpLog struct{ bytes.Buffer }

func (t *tmpLog) Close() error { return nil }

func TestWaitBeforeStart(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{LogWriteCloser: l}
	p, err := ExecutePTY(opt, "ls")
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	if err := p.Wait(); err != ErrProcessNotStarted {
		t.Fatalf("error not returned: got=%s, expected=%s", err, ErrProcessNotStarted)
	}
}

func TestLogging(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{LogWriteCloser: l}
	p, err := Execute(opt, "sh", "-c", "echo 1 && sleep 1 && echo 2 1>&2")
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("failed to start process: %s", err)
	}
	if err := p.Wait(); err != nil {
		t.Fatalf("failed to wait process: %s", err)
	}
	logs := []logline{}
	dec := json.NewDecoder(l)
	for {
		var ll logline
		if err := dec.Decode(&ll); err != nil {
			break
		}
		logs = append(logs, ll)
	}
	expected := []logline{
		{Type: "stdout", Data: []byte("1\n"), EOF: false},
		{Type: "stderr", Data: []byte("2\n"), EOF: false},
		{Type: "stdout", EOF: true},
		{Type: "stderr", EOF: true},
		{Type: "stdin", EOF: true},
	}
	if !reflect.DeepEqual(logs, expected) {
		t.Fatalf("logs mismatch:\ngot=%+v\nexpected=%+v\n", logs, expected)
	}
}

func TestPTYLogging(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{LogWriteCloser: l}
	p, err := ExecutePTY(opt, "sh", "-c", "echo 1 && sleep 1 && echo 2 1>&2")
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("failed to start process: %s", err)
	}
	if err := p.Wait(); err != nil {
		t.Fatalf("failed to wait process: %s", err)
	}
	logs := []logline{}
	dec := json.NewDecoder(l)
	for {
		var ll logline
		if err := dec.Decode(&ll); err != nil {
			break
		}
		logs = append(logs, ll)
	}
	expected := []logline{
		{Type: "stdout", Data: []byte("1\r\n"), EOF: false},
		{Type: "stdout", Data: []byte("2\r\n"), EOF: false},
		{Type: "stdout", EOF: true},
		{Type: "stderr", EOF: true},
		{Type: "stdin", EOF: true},
	}
	if !reflect.DeepEqual(logs, expected) {
		t.Fatalf("logs mismatch:\ngot=%+v\nexpected=%+v\n", logs, expected)
	}
}

func TestPTYInteractiveLog(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{
		LogWriteCloser: l,
		Env:            []string{"PS1=$ "}}
	p, err := ExecutePTY(opt, "sh")
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("failed to start process: %s", err)
	}
	p.Stdin.Write([]byte("echo 1\n"))
	p.Stdin.Write([]byte("exit\n"))
	if err := p.Wait(); err != nil {
		t.Fatalf("failed to wait process: %s", err)
	}
	logs := []logline{}
	dec := json.NewDecoder(l)
	for {
		var ll logline
		if err := dec.Decode(&ll); err != nil {
			break
		}
		logs = append(logs, ll)
	}
	out := []byte{}
	out = append(out, []byte("echo 1\r\n")...)
	out = append(out, []byte("exit\r\n")...)
	out = append(out, []byte("$ echo 1\r\n")...)
	out = append(out, []byte("1\r\n")...)
	out = append(out, []byte("$ exit\r\n")...)
	out = append(out, []byte("exit\r\n")...)
	expected := []logline{
		{Type: "stdout", Data: out, EOF: false},
		{Type: "stdout", EOF: true},
		{Type: "stderr", EOF: true},
		{Type: "stdin", EOF: true},
	}
	if !reflect.DeepEqual(logs, expected) {
		t.Fatalf("logs mismatch:\ngot=%+v\nexpected=%+v\n", logs, expected)
	}
}

func TestHoge(t *testing.T) {
	opt := &ProcessOption{
		Dir:  "./log",
		Name: "test",
	}

	p, err := ExecutePTY(opt, "ping", "-c 5", "127.0.0.1")
	//p, err := ExecutePTY(opt, "ls")
	if err != nil {
		log.Fatalf("failed to setup process: %s", err)
	}
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start process: %s", err)
	}

	r, err := NewProcessLogReader(opt)
	if err != nil {
		log.Fatalf("failed to create log reader: %s", err)
	}
	go r.Start()
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := r.Stdout.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatalf("failed to read stdout: %s", err)
				}
				break
			}
			log.Printf("%s", buf[:n])
		}
	}()
	if err := p.Wait(); err != nil {
		log.Fatalf("failed to wait process: %s", err)
	}

}
