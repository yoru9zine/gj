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

func (j *Job) ViewModel() *JobViewModel {
	return &JobViewModel{Name: j.Name}
}
