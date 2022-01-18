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

		wkr, found := d.pool.apply(task.JobId)
		if !found {
			// TODO: deal with retry and backoff policy, this will BURN CPU when no worker available!!!
			continue
		}

		//lint:ignore SA4006 false positive
		success := wkr.assign(task, d.statusNotify, d.exitNotify)
		if !success {
			// TODO: deal with retry
			continue
		}
	}
}

func (d *Decider) statusNotify(w *worker, payload *api.StatusPayload) {
	task := w.task
	task.Ctx.Status = payload.TaskStatus
	if payload.TaskStatus == api.TaskStatus_Finished {
		task.Ctx.FinalData = payload.ExecResult
	} else {
		task.Ctx.IntermediateData = payload.ExecResult
	}

	task.UpdateHandler(task)

	switch task.Ctx.Status {
	case api.TaskStatus_Error:
		fallthrough
	case api.TaskStatus_Interrupted:
		// TODO: maybe retry?
		fallthrough
	case api.TaskStatus_Finished:
		d.pool.returnBack(w)
	default:
	}
}

func (d *Decider) exitNotify(w *worker) {
	task := w.task
	if task != nil && task.Ctx.Status == api.TaskStatus_Running {
		task.Ctx.Status = api.TaskStatus_Interrupted
	}
}

func NewDecider(pool *WorkerPool, taskQ <-chan *Task) *Decider {
	return &Decider{
		pool:  pool,
		taskQ: taskQ,
	}
}
