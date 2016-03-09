package execute

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

type tmpLog struct{ bytes.Buffer }

func (t *tmpLog) Close() error { return nil }

func TestWaitBeforeStart(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{LogIO: l, AllocatePTY: true}
	p, err := NewProcess(opt, "ls")
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	if err := p.Wait(); err != ErrProcessNotStarted {
		t.Fatalf("error not returned: got=%s, expected=%s", err, ErrProcessNotStarted)
	}
}

func TestLogging(t *testing.T) {
	l := &tmpLog{}
	opt := &ProcessOption{LogIO: l}
	p, err := NewProcess(opt, "sh", "-c", "echo 1 && sleep 1 && echo 2 1>&2")
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
	opt := &ProcessOption{LogIO: l, AllocatePTY: true}
	p, err := NewProcess(opt, "sh", "-c", "echo 1 && sleep 1 && echo 2 1>&2")
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
		LogIO:       l,
		Env:         []string{"PS1=$ "},
		AllocatePTY: true,
	}
	p, err := NewProcess(opt, "sh")
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

func TestProcessLogReader(t *testing.T) {
	l := &tmpLog{}
	enc := json.NewEncoder(l)
	for _, line := range []logline{
		{Type: "stdout", Data: []byte("123"), EOF: false},
		{Type: "stdout", EOF: true},
		{Type: "stderr", EOF: true},
		{Type: "stdin", EOF: true},
	} {
		enc.Encode(line)
	}
	opt := &ProcessOption{LogIO: l}
	r, err := NewProcessLogReader(opt)
	if err != nil {
		t.Fatalf("failed to create process: %s", err)
	}
	r.Start()
}
