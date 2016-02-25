package gj

type Jobs map[string]*Job

func (j Jobs) ViewModels() map[string]*JobViewModel {
	models := make(map[string]*JobViewModel, len(j))
	for id, job := range j {
		models[id] = job.ViewModel()
	}
	return models
}

type Job struct {
	ID       string
	Name     string
	Dir      string
	Commands []*Command
}

func (j *Job) Start() error {
	for _, cmd := range j.Commands {
		_, c, err := cmd.Start()
		if err != nil {
			return err
		}
		if err := c.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (j *Job) ViewModel() *JobViewModel {
	cmds := [][]string{}
	for _, c := range j.Commands {
		cmd := []string{c.Name}
		cmd = append(cmd, c.Args...)
		cmds = append(cmds, cmd)
	}
	return &JobViewModel{
		ID:       j.ID,
		Name:     j.Name,
		Dir:      j.Dir,
		Commands: cmds,
	}
}

type JobViewModel struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Dir      string     `json:"dir"`
	Commands [][]string `json:"commands"`
}

func (j *JobViewModel) Job() *Job {
	cmds := []*Command{}
	for _, c := range j.Commands {
		cmds = append(cmds, &Command{
			Name: c[0],
			Args: c[1:],
		})
	}
	return &Job{
		ID:       j.ID,
		Name:     j.Name,
		Dir:      j.Dir,
		Commands: cmds,
	}
}
