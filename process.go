package gj

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/yoru9zine/gj/pkg/id"
)

var (
	ErrProcessNotFound = errors.New("process not found")
	ErrNotUniq         = errors.New("multiple process matched")
)

type Processes map[string]*Process

func (j Processes) ViewModels() map[string]*ProcessViewModel {
	models := make(map[string]*ProcessViewModel, len(j))
	for id, proc := range j {
		models[id] = proc.ViewModel()
	}
	return models
}

func (j Processes) Find(prefix string) (*Process, error) {
	keys := []string{}
	for k, _ := range j {
		keys = append(keys, k)
	}
	match, err := id.Search(keys, prefix)
	if err != nil {
		switch err {
		case id.ErrDuplicated:
			return nil, ErrNotUniq
		case id.ErrNotFound:
			return nil, ErrProcessNotFound
		}
		return nil, err
	}
	return j[match], nil
}

type Process struct {
	ID       string
	Name     string
	Dir      string
	Commands []*Command
	Running  bool
	Finished bool
	PTY      bool
	output   *bytes.Buffer
}

func (j *Process) Start(logDir string) error {
	j.output = &bytes.Buffer{}
	for i, cmd := range j.Commands {
		logPath := fmt.Sprintf("%s/%d", logDir, i+1)
		if err := j.start(cmd, logPath); err != nil {
			return err
		}
	}
	return nil
}

func (j *Process) start(cmd *Command, logPath string) error {
	w, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log `%s`: %s", logPath, err)
	}
	defer w.Close()

	logTap := &bytes.Buffer{}
	mw := io.MultiWriter(w, logTap)

	defer func() {
		clr := newCommandLogReader(logTap)
		io.Copy(j.output, clr.Stdout)
		io.Copy(j.output, clr.Stderr)
	}()

	_, c, err := cmd.Start(&CommandOption{
		PTY:       j.PTY,
		LogWriter: mw,
	})
	if err != nil {
		return err
	}

	j.Running = true
	err = c.Wait()
	j.Running = false
	j.Finished = true
	if err != nil {
		return err
	}
	return nil
}

func (j *Process) ViewModel() *ProcessViewModel {
	cmds := [][]string{}
	for _, c := range j.Commands {
		cmd := []string{c.Name}
		cmd = append(cmd, c.Args...)
		cmds = append(cmds, cmd)
	}
	return &ProcessViewModel{
		ID:       j.ID,
		Name:     j.Name,
		Dir:      j.Dir,
		Commands: cmds,
		Running:  j.Running,
		Finished: j.Finished,
		PTY:      j.PTY,
	}
}

type ProcessViewModel struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Dir      string     `json:"dir"`
	Commands [][]string `json:"commands"`

	Running  bool `json:"running"`
	Finished bool `json:"finished"`
	PTY      bool `json:"pty"`
}

func (j *ProcessViewModel) Process() *Process {
	cmds := []*Command{}
	for _, c := range j.Commands {
		cmds = append(cmds, &Command{
			Name: c[0],
			Args: c[1:],
		})
	}
	return &Process{
		ID:       j.ID,
		Name:     j.Name,
		Dir:      j.Dir,
		Commands: cmds,
		PTY:      j.PTY,
	}
}
