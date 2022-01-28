package module

import (
	"container/list"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

var notOccupied = "not_occupied"
var notAvailable = "not_available"

type worker struct {
	id           string
	status       api.WorkerStatus
	occupiedBy   *string
	task         *Task
	ch           chan *api.Msg
	statusNotify func(*worker, *api.StatusPayload)
	exitNotify   func(*worker)
}

func (w *worker) atomicGetOccupiedBy() *string {
	return (*string)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy))))
}

func (w *worker) assign(t *Task, notify func(*worker, *api.StatusPayload), exitNotify func(*worker)) (success bool) {
	occupiedBy := w.atomicGetOccupiedBy()
	if occupiedBy == &notOccupied || *occupiedBy != t.JobId || w.task != nil {
		return false
	}

	w.statusNotify = notify
	w.exitNotify = exitNotify
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
	w.task = nil
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy)), unsafe.Pointer(&notOccupied))
}

func (w *worker) moribund() {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&w.occupiedBy)), unsafe.Pointer(&notAvailable))
}

type WorkerPool struct {
	pool     map[string]*worker
	freeList *list.List
	lock     sync.RWMutex
}

func (w *WorkerPool) Add(id string, ch chan *api.Msg) {
	w.lock.Lock()
	defer w.lock.Unlock()
	_, exist := w.pool[id]
	if exist {
		return
	}

	newWorker := &worker{
		id:         id,
		status:     api.WorkerStatus_Idle,
		occupiedBy: &notOccupied,
		ch:         ch,
	}
	w.pool[id] = newWorker
	w.freeList.PushFront(newWorker)
}

func (w *WorkerPool) Remove(id string) {
	w.lock.Lock()
	defer w.lock.Unlock()
	wkr, exist := w.pool[id]
	if !exist {
		return
	}

	// occupy with "not_available" to prevent from other goroutine try to apply this ready-to-close worker
	if !wkr.occupy(notAvailable) {
		// failed to occupy means this worker have been occupied by some job's task, notify to exit
		wkr.exitNotify(wkr)
	}

	// now the worker can be safe delete
	delete(w.pool, id)

	// no need to clear free list, we can eliminate it when the "not available" worker be applied
	wkr.moribund()
}

func (w *WorkerPool) apply(jobId string) (wkr *worker, found bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for {
		e := w.freeList.Back()
		if e == nil {
			return nil, false
		}

		wkr = w.freeList.Remove(e).(*worker)
		if *wkr.atomicGetOccupiedBy() == notAvailable {
			// ignore worker that already removed
			continue
		}

		if !wkr.occupy(jobId) {
			continue
		}

		break
	}

	return wkr, true
}

func (w *WorkerPool) returnBack(wkr *worker) {
	w.lock.Lock()
	defer w.lock.Unlock()

	wkr.status = api.WorkerStatus_Idle
	wkr.release()

	w.freeList.PushFront(wkr)
}

func (w *WorkerPool) UpdateStatus(id string, payload *api.StatusPayload) error {
	w.lock.RLock()
	wkr, exist := w.pool[id]
	w.lock.RUnlock()

	if !exist {
		return errors.Errorf("Worker id: %s not regsitered, no context found.", id)
	}

	if !wkr.occupied() {
		return errors.Errorf("Worker id: %s not occupied", id)
	}

	if wkr.task == nil || (wkr.task.Id != payload.TaskId) {
		return errors.Errorf("Task id: %s not assigned to worker %s", payload.TaskId, id)
	}

	wkr.status = payload.WorkStatus
	wkr.statusNotify(wkr, payload)
	return nil
}

func (w *WorkerPool) InterruptJobTasks(jobId string) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	for _, wkr := range w.pool {
		if wkr.occupied() && wkr.task.JobId == jobId {
			wkr.interrupt()
		}
	}
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		pool:     make(map[string]*worker),
		freeList: list.New(),
	}
}
