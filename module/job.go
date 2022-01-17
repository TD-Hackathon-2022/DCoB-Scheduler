package module

type Spliterator interface {
	TryAdvance(func(task *Task)) (finished bool)
}

type Job interface {
	Spliterator
	Id() string
	// TODO: how to force stop ?
}

type JobRunner struct {
	jobQ    chan Job
	currJob Job
	taskQ   chan<- *Task
	stopCh  chan struct{}
}

func (j *JobRunner) Submit(job Job) {
	j.jobQ <- job
}

func (j *JobRunner) Start() {
Exit:
	for {
		select {
		case <-j.stopCh:
			break Exit
		case job := <-j.jobQ:
			j.currJob = job
			for {
				finished := j.currJob.TryAdvance(func(task *Task) { j.taskQ <- task })
				if finished {
					break
				}
			}
		}
	}
}

func (j *JobRunner) ShutDown() {
	// TODO: stop current job
	close(j.stopCh)
}

func NewJobRunner(taskQ chan<- *Task) *JobRunner {
	return &JobRunner{
		jobQ:   make(chan Job, 16),
		taskQ:  taskQ,
		stopCh: make(chan struct{}),
	}
}
