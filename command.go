package gj

import (
	"io"
	"os/exec"

	"github.com/kr/pty"
)

type Command struct {
	Name string
	Args []string
}

func (c *Command) Start(opt *CommandOption) (io.Writer, io.Reader, io.Reader, *exec.Cmd, error) {
	cmd := exec.Command(c.Name, c.Args...)
	f, err := pty.Start(cmd)
	return f, f, nil, cmd, err
}

type CommandOption struct {
	PTY bool
}
