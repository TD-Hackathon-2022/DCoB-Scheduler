package module

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"sync"
	"sync/atomic"
	"unsafe"
)

type workerStatus int

const (
	idle workerStatus = iota
	busy
	closing
)

var notOccupied = "not_occupied"

type worker struct {
	id         string
	status     workerStatus
	occupiedBy *string
	ch         chan *api.Msg
}

func (w *worker) assign(t *Task) (success bool) {
	occupiedBy := (*string)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy))))

	if occupiedBy == &notOccupied {
		return false
	}

	if *occupiedBy != t.JobId {
		return false
	}

	w.ch <- &api.Msg{
		Cmd: api.CMD_Assign,
		Payload: &api.Msg_Assign{
			Assign: &api.AssignPayload{
				TaskId: t.Id,
				Data:   t.Ctx.initData.(string),
				FuncId: t.FuncId,
			},
		},
	}
	return true
}

func (w *worker) occupied(jobId string) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy)),
		unsafe.Pointer(&notOccupied),
		unsafe.Pointer(&jobId))
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

func (w *WorkerPool) Add(id string, ch chan *api.Msg) {
	_, exist := w.pool.Load(id)
	if exist {
		return
	}

	w.pool.Store(id, &worker{
		id:         id,
		status:     idle,
		occupiedBy: &notOccupied,
		ch:         ch,
	})
}

func (w *WorkerPool) remove() {
	panic("not supported yet")
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

	if wkr != nil {
		return wkr, true
	}

	return nil, false
}

func (w *WorkerPool) release(wkr *worker) {
	wkr.status = idle
	wkr.release()
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{}
}
