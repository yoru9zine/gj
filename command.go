package gj

import (
	"os"
	"os/exec"

	"github.com/kr/pty"
)

type Command struct {
	Name string
	Args []string
}

func (c *Command) Start() (*os.File, *exec.Cmd, error) {
	cmd := exec.Command(c.Name, c.Args...)
	f, err := pty.Start(cmd)
	return f, cmd, err
}
