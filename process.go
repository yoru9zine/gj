package gj

import (
	"bytes"
	"io"
)

type Processes map[string]*Process

func (j Processes) ViewModels() map[string]*ProcessViewModel {
	models := make(map[string]*ProcessViewModel, len(j))
	for id, proc := range j {
		models[id] = proc.ViewModel()
	}
	return models
}

type Process struct {
	ID       string
	Name     string
	Dir      string
	Commands []*Command
	Running  bool
	Finished bool

	output *bytes.Buffer
}

func (j *Process) Start() error {
	j.output = &bytes.Buffer{}
	for _, cmd := range j.Commands {
		f, c, err := cmd.Start()
		if err != nil {
			return err
		}
		j.Running = true
		go io.Copy(j.output, f)

		err = c.Wait()
		j.Running = false
		j.Finished = true
		if err != nil {
			return err
		}
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
	}
}

type ProcessViewModel struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Dir      string     `json:"dir"`
	Commands [][]string `json:"commands"`

	Running  bool `json:"running"`
	Finished bool `json:"finished"`
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
	}
}
