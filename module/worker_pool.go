package module

import (
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
}

type workerPool struct {
	pool sync.Map
}

func (w *workerPool) add(id string) {
	_, exist := w.pool.Load(id)
	if exist {
		return
	}

	w.pool.Store(id, &worker{
		id:         id,
		status:     idle,
		occupiedBy: &notOccupied,
	})
}

func (w *workerPool) remove() {

}

func (w *workerPool) occupy(jobId string) (wkr *worker, found bool) {
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

func (w *workerPool) release(*worker) {

}
