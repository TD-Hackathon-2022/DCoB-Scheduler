package module

import (
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

var notOccupied = "not_occupied"

type worker struct {
	id         string
	status     WorkerStatus
	occupiedBy *string
	task       *Task
	ch         chan *Msg
}

func (w *worker) atomicGetOccupiedBy() *string {
	return (*string)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy))))
}

func (w *worker) assign(t *Task) (success bool) {
	occupiedBy := w.atomicGetOccupiedBy()
	if occupiedBy == &notOccupied {
		return false
	}

	if *occupiedBy != t.JobId {
		return false
	}

	if w.task != nil && w.task.Ctx.status == TaskStatus_Running {
		return false
	}

	w.task = t
	w.ch <- &Msg{
		Cmd: CMD_Assign,
		Payload: &Msg_Assign{
			Assign: &AssignPayload{
				TaskId: t.Id,
				Data:   t.Ctx.initData.(string),
				FuncId: t.FuncId,
			},
		},
	}
	return true
}

func (w *worker) interrupt(t *Task) {
	occupiedBy := w.atomicGetOccupiedBy()
	if occupiedBy == &notOccupied || *occupiedBy != t.JobId {
		return
	}

	w.task = t
	w.ch <- &Msg{
		Cmd: CMD_Interrupt,
		Payload: &Msg_Interrupt{
			Interrupt: &InterruptPayload{
				TaskId: t.Id,
			},
		},
	}
}

func (w *worker) occupied() bool {
	return w.atomicGetOccupiedBy() != &notOccupied
}

func (w *worker) occupy(jobId string) (success bool) {
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

func (w *WorkerPool) Add(id string, ch chan *Msg) {
	_, exist := w.pool.Load(id)
	if exist {
		return
	}

	w.pool.Store(id, &worker{
		id:         id,
		status:     WorkerStatus_Idle,
		occupiedBy: &notOccupied,
		ch:         ch,
	})
}

func (w *WorkerPool) Remove(id string) {
	wkr, exist := w.pool.Load(id)
	if !exist {
		return
	}

	task := wkr.(*worker).task
	if task != nil {
		if task.Ctx.status == TaskStatus_Running {
			task.Ctx.status = TaskStatus_Interrupted
		}

		task.UpdateNotify()
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
	wkr.status = WorkerStatus_Idle
	wkr.release()
}

func (w *WorkerPool) UpdateStatus(id string, payload *StatusPayload) error {
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
	wkr.task.Ctx.status = payload.TaskStatus
	if payload.TaskStatus == TaskStatus_Finished {
		wkr.task.Ctx.finalData = payload.ExecResult
	} else {
		wkr.task.Ctx.intermediateData = payload.ExecResult
	}

	wkr.task.UpdateNotify()
	return nil
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{}
}
