package module

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

var notOccupied = "not_occupied"

type worker struct {
	id         string
	status     api.WorkerStatus
	occupiedBy *string
	task       *Task
	ch         chan *api.Msg
	notify     func(*worker)
}

func (w *worker) atomicGetOccupiedBy() *string {
	return (*string)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy))))
}

func (w *worker) assign(t *Task, notify func(*worker)) (success bool) {
	occupiedBy := w.atomicGetOccupiedBy()
	if occupiedBy == &notOccupied {
		return false
	}

	if *occupiedBy != t.JobId {
		return false
	}

	if w.task != nil && w.task.Ctx.Status == api.TaskStatus_Running {
		return false
	}

	w.notify = notify
	w.task = t
	w.ch <- &api.Msg{
		Cmd: api.CMD_Assign,
		Payload: &api.Msg_Assign{
			Assign: &api.AssignPayload{
				TaskId: t.Id,
				Data:   t.Ctx.InitData.(string),
				FuncId: t.FuncId,
			},
		},
	}
	return true
}

func (w *worker) interrupt() {
	w.ch <- &api.Msg{
		Cmd: api.CMD_Interrupt,
		Payload: &api.Msg_Interrupt{
			Interrupt: &api.InterruptPayload{
				TaskId: w.task.Id,
			},
		},
	}
}

func (w *worker) occupied() bool {
	return w.atomicGetOccupiedBy() != &notOccupied
}

func (w *worker) occupy(jobId string) (success bool) {
	// TODO: consider closing status
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy)),
		unsafe.Pointer(&notOccupied),
		unsafe.Pointer(&jobId))
}

func (w *worker) release() {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy)), unsafe.Pointer(&notOccupied))
}

type WorkerPool struct {
	pool sync.Map
}

func (w *WorkerPool) Add(id string, ch chan *api.Msg) {
	_, exist := w.pool.Load(id)
	if exist {
		return
	}

	w.pool.Store(id, &worker{
		id:         id,
		status:     api.WorkerStatus_Idle,
		occupiedBy: &notOccupied,
		ch:         ch,
	})
}

func (w *WorkerPool) Remove(id string) {
	wkrOri, exist := w.pool.Load(id)
	if !exist {
		return
	}

	wkr := wkrOri.(*worker)
	task := wkr.task
	if task != nil {
		if task.Ctx.Status == api.TaskStatus_Running {
			task.Ctx.Status = api.TaskStatus_Interrupted
		}

		wkr.notify(wkr)
	}

	w.pool.Delete(id)
}

func (w *WorkerPool) occupy(jobId string) (wkr *worker, found bool) {
	w.pool.Range(func(_, v interface{}) bool {
		wkr = v.(*worker)
		if wkr.occupy(jobId) {
			return false
		}

		// clear wkr
		wkr = nil
		return true
	})

	return wkr, wkr != nil
}

func (w *WorkerPool) release(wkr *worker) {
	wkr.status = api.WorkerStatus_Idle
	wkr.release()
}

func (w *WorkerPool) UpdateStatus(id string, payload *api.StatusPayload) error {
	wkrOri, exist := w.pool.Load(id)
	if !exist {
		return errors.Errorf("Worker id: %s not regsitered, no context found.", id)
	}

	wkr := wkrOri.(*worker)
	if !wkr.occupied() {
		return errors.Errorf("Worker id: %s not occupied", id)
	}

	if wkr.task.Id != payload.TaskId {
		return errors.Errorf("Task id: %s not assigned to worker %s", payload.TaskId, id)
	}

	wkr.status = payload.WorkStatus
	wkr.task.Ctx.Status = payload.TaskStatus
	if payload.TaskStatus == api.TaskStatus_Finished {
		wkr.task.Ctx.FinalData = payload.ExecResult
	} else {
		wkr.task.Ctx.IntermediateData = payload.ExecResult
	}

	wkr.notify(wkr)
	return nil
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{}
}
