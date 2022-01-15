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
		swapped := atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&wkr.occupiedBy)),
			unsafe.Pointer(&notOccupied),
			unsafe.Pointer(&jobId))
		if swapped {
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
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&wkr.occupiedBy)), unsafe.Pointer(&notOccupied))
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{}
}
