package module

import "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"

type Decider struct {
	taskQ <-chan *Task
	pool  *WorkerPool
}

func (d *Decider) Start() {
	for task := range d.taskQ {
		if task == poisonTask {
			d.pool.InterruptJobTasks(task.JobId)
			continue
		}

		wkr, found := d.pool.occupy(task.JobId)
		if !found {
			// TODO: deal with retry
			continue
		}

		//lint:ignore SA4006 false positive
		success := wkr.assign(task, d.statusNotify)
		if !success {
			// TODO: deal with retry
			continue
		}
	}
}

func (d *Decider) statusNotify(w *worker) {
	task := w.task
	if task == nil {
		return
	}

	switch task.Ctx.Status {
	case api.TaskStatus_Error:
		fallthrough
	case api.TaskStatus_Interrupted:
		// TODO: maybe retry?
		fallthrough
	case api.TaskStatus_Finished:
		d.pool.release(w)
	default:

	}

	task.UpdateHandler(task)
}

func NewDecider(pool *WorkerPool, taskQ <-chan *Task) *Decider {
	return &Decider{
		pool:  pool,
		taskQ: taskQ,
	}
}
