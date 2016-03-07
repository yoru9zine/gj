package execute

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/kr/pty"
)

var m sync.Mutex

type Process struct {
	Stdin     io.WriteCloser
	cmd       *exec.Cmd
	logWriter *ProcessLogWriter
	m         sync.Mutex
	tty       *os.File
	pty       *os.File

	readerChannels map[string]*reader2chan
}

func (p *Process) Start() error {
	if p.tty != nil {
		defer p.tty.Close()
	}
	return p.cmd.Start()
}

func (p *Process) Wait() error {
	p.m.Lock()
	defer p.m.Unlock()
	cmdErr := p.cmd.Wait()
	for _, c := range p.readerChannels {
		c.Stop()
	}
	if err := p.logWriter.Close(); err != nil {
		return fmt.Errorf("failed to close log: %s", err)
	}
	return cmdErr
}

/*
func (p *Process) handleInput(r io.Reader, logType string) {
	buf := make([]byte, 1024)
	for {
		p.m.Lock()
		p.logWriter.WriteOutput(buf, r, logType)
		time.Sleep(100 * time.Millisecond)
		p.m.Unlock()
	}
}
*/
func (p *Process) handleInput2(c chan []byte, logType string) {
	for line := range c {
		p.logWriter.WriteOutput2(line, logType)
	}
}

type ProcessOption struct {
	Dir  string
	Name string

	AllocatePTY bool
}

func (o *ProcessOption) logFile() string {
	return fmt.Sprintf("%s/%s/log.json", o.Dir, o.Name)
}

/*
func Execute(opt *ProcessOption, cmds ...string) (*Process, error) {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	logwriter, err := newProcessLogWriter(opt.logFile())
	if err != nil {
		return nil, fmt.Errorf(`failed to create logger: %s`, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe for STDOUT: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe for STDERR: %s", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe for STDIN: %s", err)
	}
	stdinbuf := &bytes.Buffer{}
	p := &Process{
		Stdin:     &multiIO{ioObjects: []interface{}{stdin, stdinbuf}},
		cmd:       cmd,
		logWriter: logwriter,
	}
	go p.handleInput(stdout, "stdout")
	go p.handleInput(stderr, "stderr")
	go p.handleInput(stdinbuf, "stdin")
	return p, nil
}
*/

func ExecutePTY(opt *ProcessOption, cmds ...string) (*Process, error) {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	logwriter, err := newProcessLogWriter(opt.logFile())
	if err != nil {
		return nil, fmt.Errorf(`failed to create logger: %s`, err)
	}
	pty, tty, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to create pty: %s", err)
	}
	stdinbuf := &bytes.Buffer{}

	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.Stdin = tty

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	p := &Process{
		Stdin:     &multiIO{ioObjects: []interface{}{pty, stdinbuf}},
		cmd:       cmd,
		logWriter: logwriter,
		tty:       tty,
		pty:       pty,
	}

	p.readerChannels = map[string]*reader2chan{}
	p.readerChannels["stdout"] = &reader2chan{reader: pty, Channel: make(chan []byte), fin: make(chan struct{})}
	for t, c := range p.readerChannels {
		go c.Start()
		go p.handleInput2(c.Channel, t)
	}
	return p, nil
}

type reader2chan struct {
	reader  io.Reader
	Channel chan []byte
	fin     chan struct{}
}

func (r *reader2chan) Start() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-r.fin:
			break
		case <-time.After(100 * time.Millisecond):
			m.Lock()
			n, err := r.reader.Read(buf)
			m.Unlock()
			if err != nil {
				if err == io.EOF {
					continue
				}
				close(r.Channel)
			}
			r.Channel <- buf[:n]
		}
	}
}

func (r *reader2chan) Stop() {
	r.fin <- struct{}{}
}
