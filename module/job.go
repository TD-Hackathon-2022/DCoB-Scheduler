package module

import "sync/atomic"

type Spliterator interface {
	TryAdvance(func(task *Task)) (finished bool)
}

type Job interface {
	Spliterator
	Id() string
	GetResult() map[string]interface{}
}

var poisonTask = &Task{}

type JobRunner struct {
	jobQ           chan Job
	store          JobStore
	currJob        Job
	taskQ          chan<- *Task
	interruptJobCh chan struct{}
	stopCh         chan struct{}
	stopped        uint32
}

func (j *JobRunner) Submit(job Job) {
	select {
	case <-j.stopCh:
		// return error if closed
		return
	default:
		j.jobQ <- job
	}
}

func (j *JobRunner) Start() {
	atomic.StoreUint32(&j.stopped, 0)

Exit:
	for {
		select {
		case <-j.stopCh:
			break Exit
		case job := <-j.jobQ:
			j.currJob = job
			j.store.Store(job)
			for {
				select {
				case <-j.interruptJobCh:
					poisonTask.JobId = j.currJob.Id()
					j.taskQ <- poisonTask
					continue Exit
				default:
				}

				finished := j.currJob.TryAdvance(func(task *Task) { j.taskQ <- task })
				if finished {
					break
				}
			}
		}
	}

	atomic.StoreUint32(&j.stopped, 1)
}

func (j JobRunner) InterruptCurrentJob() {
	j.interruptJobCh <- struct{}{}
}

func (j *JobRunner) ShutDown() {
	close(j.stopCh)

	// if not stopped, always retry to interrupt current job
	for atomic.LoadUint32(&j.stopped) == 0 {
		select {
		case j.interruptJobCh <- struct{}{}:
		default:
		}
	}
}

func (j *JobRunner) GetJobById(jobId string) (job Job, exist bool) {
	v, exist := j.store.Load(jobId)
	if !exist {
		return nil, false
	}
	return v, true
}

func NewJobRunner(taskQ chan<- *Task, store JobStore) *JobRunner {
	return &JobRunner{
		jobQ:           make(chan Job, 16),
		store:          store,
		taskQ:          taskQ,
		interruptJobCh: make(chan struct{}),
		stopCh:         make(chan struct{}),
	}
}
