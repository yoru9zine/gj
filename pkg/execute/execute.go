package execute

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"

	"github.com/kr/pty"
)

var (
	// ErrProcessNotStarted is returned when Wait calls before Start
	ErrProcessNotStarted = errors.New("process not started")
)

// A ProcessOption is used to configure a Process
type ProcessOption struct {
	Dir  string
	Name string
	Env  []string

	LogWriteCloser io.WriteCloser
	AllocatePTY    bool
}

func (o *ProcessOption) logFile() string {
	return fmt.Sprintf("%s/%s/log.json", o.Dir, o.Name)
}

func (o *ProcessOption) writeCloser() (io.WriteCloser, error) {
	if o.LogWriteCloser != nil {
		return o.LogWriteCloser, nil
	}
	dir := path.Dir(o.logFile())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create `%s`: %s", dir, err)
	}
	f, err := os.Create(o.logFile())
	if err != nil {
		return nil, fmt.Errorf("failed to create ``")
	}
	return f, nil
}

// Process represents an external command
type Process struct {
	Stdin     io.WriteCloser
	cmd       *exec.Cmd
	logWriter *processLogWriter
	m         sync.Mutex
	tty       *os.File
	pty       *os.File

	started bool

	readerChannels map[string]*reader2chan
	handleFinish   chan struct{}
	ErrorAtLogging error
}

// Start starts process
func (p *Process) Start() error {
	p.handleFinish = make(chan struct{})
	if p.tty != nil {
		defer p.tty.Close()
	}
	p.started = true
	return p.cmd.Start()
}

// Wait waits process
func (p *Process) Wait() error {
	if !p.started {
		return ErrProcessNotStarted
	}
	cmdErr := p.cmd.Wait()
	for _, c := range p.readerChannels {
		c.Stop()
		<-p.handleFinish
	}
	if err := p.logWriter.Close(); err != nil {
		return fmt.Errorf("failed to close log: %s", err)
	}
	return cmdErr
}

func (p *Process) handleInput(c chan []byte, logType string) {
	for line := range c {
		if err := p.logWriter.Write(line, logType); err != nil {
			p.ErrorAtLogging = err
		}
	}
	p.handleFinish <- struct{}{}
}

func Execute(opt *ProcessOption, cmds ...string) (*Process, error) {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Env = opt.Env
	f, err := opt.writeCloser()
	if err != nil {
		return nil, fmt.Errorf("failed to open log", err)
	}
	logwriter, err := newProcessLogWriter(f)
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
	p.readerChannels = map[string]*reader2chan{}
	p.readerChannels["stdout"] = newReader2Chan(stdout)
	p.readerChannels["stderr"] = newReader2Chan(stderr)
	p.readerChannels["stdin"] = newReader2Chan(stdinbuf)
	for t, c := range p.readerChannels {
		go c.Start()
		go p.handleInput(c.Channel, t)
	}
	return p, nil
}

func ExecutePTY(opt *ProcessOption, cmds ...string) (*Process, error) {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Env = opt.Env
	f, err := opt.writeCloser()
	if err != nil {
		return nil, fmt.Errorf("failed to open log", err)
	}
	logwriter, err := newProcessLogWriter(f)
	if err != nil {
		return nil, fmt.Errorf(`failed to create logger: %s`, err)
	}
	pty, tty, err := pty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to create pty: %s", err)
	}

	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.Stdin = tty

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true

	p := &Process{
		Stdin:     pty,
		cmd:       cmd,
		logWriter: logwriter,
		tty:       tty,
		pty:       pty,
	}

	p.readerChannels = map[string]*reader2chan{}
	p.readerChannels["stdout"] = newReader2Chan(pty)
	for t, c := range p.readerChannels {
		go c.Start()
		go p.handleInput(c.Channel, t)
	}
	return p, nil
}
