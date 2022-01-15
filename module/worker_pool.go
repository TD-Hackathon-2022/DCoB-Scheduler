package module

import "sync"

type workerStatus int

const (
	idle workerStatus = iota
	busy
	closing
)

type worker struct {
	id         string
	status     workerStatus
	occupiedBy string
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
		id:     id,
		status: idle,
	})
}

func (w *workerPool) remove() {

}

func (w *workerPool) occupy() *worker {
	return nil
}

func (w *workerPool) release(*worker) {

}
