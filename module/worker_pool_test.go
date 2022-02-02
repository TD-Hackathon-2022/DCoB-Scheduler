package module

import (
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	. "github.com/smartystreets/goconvey/convey"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool_ShouldAddWorkerToPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()

		Convey("when add worker", func() {
			addr := "127.0.0.1:8081"
			wp.Add(addr, nil)

			Convey("then pool init new worker", func() {
				w, exist := wp.pool[addr]
				So(exist, ShouldBeTrue)
				So(w, ShouldHaveSameTypeAs, &worker{})
				So(w.id, ShouldEqual, addr)
				So(w.status, ShouldEqual, WorkerStatus_Idle)
				So(w.occupiedBy, ShouldEqual, &notOccupied)

				So(wp.freeList.Len(), ShouldEqual, 1)
				So(wp.freeList.Back().Value.(*worker), ShouldEqual, w)
			})
		})
	})
}

func TestWorkerPool_ShouldDoNothingWhenAddWorkerThatAlreadyInPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()
		addr := "127.0.0.1:8081"
		w := &worker{id: "fake-worker", status: WorkerStatus_Busy}
		wp.pool[addr] = w
		wp.freeList.PushFront(w)

		Convey("when add worker", func() {
			wp.Add(addr, nil)

			Convey("then do nothing", func() {
				w := wp.pool[addr]
				So(w.id, ShouldEqual, "fake-worker")
				So(w.status, ShouldEqual, WorkerStatus_Busy)
				So(wp.freeList.Len(), ShouldEqual, 1)
			})
		})
	})
}

func TestWorkerPool_ShouldBlockApplyWorker(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()
		addr := "127.0.0.1:8081"
		w1 := &worker{id: addr, status: WorkerStatus_Idle, occupiedBy: &notOccupied}
		wp.pool[addr] = w1
		wp.freeList.PushFront(w1)
		job0 := "job-id-0"
		job1 := "job-id-1"

		Convey("when try apply a worker", func() {
			wkr := wp.blockApply(job0)
			Convey("then worker should be returned without block", func() {
				So(wkr.id, ShouldEqual, addr)
				So(*wkr.occupiedBy, ShouldEqual, job0)
			})

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				wkr = wp.blockApply(job1)
				wg.Done()
			}()
			runtime.Gosched()
			time.Sleep(time.Second)

			Convey("should still blocked and not apply", func() {
				So(wkr.id, ShouldEqual, addr)
				So(*wkr.occupiedBy, ShouldEqual, job0)
			})

			wp.returnBack(wkr)
			runtime.Gosched()
			wg.Wait()
			Convey("then worker should be returned then apply again", func() {
				So(wkr.id, ShouldEqual, addr)
				So(*wkr.occupiedBy, ShouldEqual, job1)
			})
		})
	})
}

func TestWorkerPool_ShouldApplyWorker(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		w0 := &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &job0}
		wp.pool[addr0] = w0
		wp.freeList.PushFront(w0)
		addr1 := "127.0.0.1:8082"
		w1 := &worker{id: addr1, status: WorkerStatus_Idle, occupiedBy: &notOccupied}
		wp.pool[addr1] = w1
		wp.freeList.PushFront(w1)
		addr2 := "127.0.0.1:8083"
		job1 := "job-1"
		w2 := &worker{id: addr2, status: WorkerStatus_Idle, occupiedBy: &job1}
		wp.pool[addr2] = w2
		wp.freeList.PushFront(w2)

		Convey("when try apply a worker", func() {
			wFirst := wp.chooseFreeWorker("job-id-2")
			wSecond := wp.chooseFreeWorker("job-id-2")

			Convey("then worker1 should be returned", func() {
				So(wFirst, ShouldBeNil)
				So(wSecond.id, ShouldEqual, addr1)
				So(*wSecond.occupiedBy, ShouldEqual, "job-id-2")
			})
		})
	})
}

func TestWorkerPool_ShouldNotApplyWorkerIfNotAvailable(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		wp.pool[addr0] = &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &job0}
		addr1 := "127.0.0.1:8082"
		job1 := "job-1"
		wp.pool[addr1] = &worker{id: addr1, status: WorkerStatus_Idle, occupiedBy: &job1}
		addr2 := "127.0.0.1:8083"
		job2 := "job-2"
		wp.pool[addr2] = &worker{id: addr2, status: WorkerStatus_Busy, occupiedBy: &job2}

		Convey("when try apply a worker", func() {
			wkr := wp.chooseFreeWorker("job-id-3")

			Convey("then no worker should be returned", func() {
				So(wkr, ShouldBeNil)
			})
		})
	})
}

func TestWorkerPool_ShouldNotApplyWorkerIfNoWorker(t *testing.T) {
	Convey("given empty worker pool", t, func() {
		wp := NewWorkerPool()

		Convey("when try apply a worker", func() {
			wkr := wp.chooseFreeWorker("job-id-3")

			Convey("then no worker should be returned", func() {
				So(wkr, ShouldBeNil)
			})
		})
	})
}

func TestWorkerPool_ShouldApplyWorkerThenRelease(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := NewWorkerPool()
		addr0 := "127.0.0.1:8081"
		w := &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &notOccupied}
		wp.pool[addr0] = w
		wp.freeList.PushFront(w)

		Convey("when try apply a worker", func() {
			w, found := wp.apply("job-id-0")

			Convey("then worker1 should be returned", func() {
				So(found, ShouldBeTrue)
				So(w.id, ShouldEqual, addr0)
				So(*w.occupiedBy, ShouldEqual, "job-id-0")
				So(wp.freeList.Len(), ShouldEqual, 0)
			})

			wp.returnBack(w)

			Convey("can returnBack", func() {
				So(w.status, ShouldEqual, WorkerStatus_Idle)
				So(w.occupiedBy, ShouldEqual, &notOccupied)
				So(wp.freeList.Len(), ShouldEqual, 1)
			})
		})
	})
}

func TestWorker_ShouldNotAssignTaskWhenNotOccupied(t *testing.T) {
	Convey("given worker", t, func() {
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Idle, occupiedBy: &notOccupied}

		Convey("when try assign a task", func() {
			success := w.assign(&Task{}, func(*worker, *StatusPayload) {}, func(*worker) {})

			Convey("then assign failed", func() {
				So(success, ShouldBeFalse)
			})
		})
	})
}

func TestWorker_ShouldNotAssignTaskWhenOccupiedByAnotherJob(t *testing.T) {
	Convey("given worker", t, func() {
		job0 := "job0"
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Idle, occupiedBy: &job0}

		Convey("when try assign a task", func() {
			job1 := "job1"
			success := w.assign(&Task{JobId: job1}, func(*worker, *StatusPayload) {}, func(*worker) {})

			Convey("then assign failed", func() {
				So(success, ShouldBeFalse)
			})
		})
	})
}

func TestWorker_ShouldNotAssignTaskWhenThereIsARunningTask(t *testing.T) {
	Convey("given worker", t, func() {
		job0 := "job0"
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Idle, occupiedBy: &job0, task: &Task{
			Ctx: &Context{Status: TaskStatus_Running},
		}}

		Convey("when try assign a task", func() {
			success := w.assign(&Task{JobId: job0}, func(*worker, *StatusPayload) {}, func(*worker) {})

			Convey("then assign failed", func() {
				So(success, ShouldBeFalse)
			})
		})
	})
}

func TestWorker_ShouldAssignTaskToOutputCh(t *testing.T) {
	Convey("given worker", t, func() {
		job0 := "job0"
		ch := make(chan *Msg, 1)
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Idle, occupiedBy: &job0, ch: ch}

		Convey("when try assign a task", func() {
			task0 := "task0"
			funcId := "hash-func"
			success := w.assign(&Task{
				Id:    task0,
				JobId: job0,
				Ctx: &Context{
					InitData: "fake-data",
				},
				FuncId: funcId,
			}, func(*worker, *StatusPayload) {}, func(*worker) {})

			Convey("then assign success", func() {
				So(success, ShouldBeTrue)

				msg := <-ch
				So(msg.Cmd, ShouldEqual, CMD_Assign)
				So(msg.GetAssign().TaskId, ShouldEqual, task0)
				So(msg.GetAssign().FuncId, ShouldEqual, funcId)
				So(msg.GetAssign().Data, ShouldEqual, "fake-data")
			})
		})
	})
}

func TestWorker_ShouldInterruptTaskToOutputCh(t *testing.T) {
	Convey("given worker", t, func() {
		job0 := "job0"
		ch := make(chan *Msg, 1)

		task0 := "task0"
		funcId := "hash-func"
		task := &Task{
			Id:    task0,
			JobId: job0,
			Ctx: &Context{
				InitData: "fake-data",
			},
			FuncId: funcId,
		}
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Busy, occupiedBy: &job0, ch: ch, task: task}

		Convey("when try interrupt a task", func() {

			w.interrupt()

			Convey("then assign success", func() {
				msg := <-ch
				So(msg.Cmd, ShouldEqual, CMD_Interrupt)
				So(msg.GetInterrupt().TaskId, ShouldEqual, task0)
			})
		})
	})
}

func TestWorkerPool_ShouldUpdateWorkerStatusAndNotify(t *testing.T) {
	Convey("given worker pool", t, func() {
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		wp := NewWorkerPool()
		addr := "127.0.0.1:8081"
		notified := false
		w := &worker{id: addr, status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task, statusNotify: func(*worker, *StatusPayload) {
			notified = true
		}}
		wp.pool[addr] = w
		wp.freeList.PushFront(w)

		Convey("when update status", func() {
			status := &StatusPayload{
				WorkStatus: WorkerStatus_Idle,
				TaskId:     task.Id,
				TaskStatus: TaskStatus_Finished,
				ExecResult: "success",
			}

			_ = wp.UpdateStatus(addr, status)

			Convey("then worker updated", func() {
				w := wp.pool[addr]
				So(w.id, ShouldEqual, addr)
				So(w.status, ShouldEqual, WorkerStatus_Idle)
				So(notified, ShouldBeTrue)
			})
		})
	})
}

func TestWorkerPool_ShouldRemoveOccupiedWorkerAndNotifyExit(t *testing.T) {
	Convey("given worker pool", t, func() {
		notified := false
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		wp := NewWorkerPool()
		addr := "127.0.0.1:8081"
		w := &worker{id: addr, status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task, exitNotify: func(*worker) {
			notified = true
		}}
		wp.pool[addr] = w
		wp.freeList.PushFront(w)

		Convey("when remove worker", func() {
			wp.Remove(addr)

			Convey("then worker updated", func() {
				_, exist := wp.pool[addr]
				So(exist, ShouldBeFalse)
				So(notified, ShouldBeTrue)
				So(*w.occupiedBy, ShouldEqual, notAvailable)
			})
		})
	})
}

func TestWorkerPool_ShouldRemoveIdleWorker(t *testing.T) {
	Convey("given worker pool", t, func() {
		addr := "127.0.0.1:8081"
		wkr := &worker{id: addr, status: WorkerStatus_Idle, occupiedBy: &notOccupied}

		wp := NewWorkerPool()
		wp.pool[addr] = wkr
		wp.freeList.PushFront(wkr)

		Convey("when remove worker", func() {
			wp.Remove(addr)

			Convey("then worker updated", func() {
				_, exist := wp.pool[addr]
				So(exist, ShouldBeFalse)
				So(*wkr.occupiedBy, ShouldEqual, notAvailable)
			})
		})
	})
}

func TestWorkerPool_ShouldInterruptAllWorkerWithGivenJob(t *testing.T) {
	Convey("given worker pool", t, func() {
		jobId0 := "fake-job-id-0"
		task0 := &Task{
			Id:    "fake-task-0",
			JobId: jobId0,
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		task1 := &Task{
			Id:    "fake-task-1",
			JobId: jobId0,
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		jobId1 := "fake-job-id-1"
		task2 := &Task{
			Id:    "fake-task-2",
			JobId: jobId1,
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		outputCh := make(chan *Msg, 5)
		wp := NewWorkerPool()
		addr0 := "127.0.0.1:8081"
		w0 := &worker{id: addr0, status: WorkerStatus_Busy, occupiedBy: &jobId0, task: task0, ch: outputCh}
		wp.pool[addr0] = w0
		wp.freeList.PushFront(w0)
		addr1 := "127.0.0.1:8082"
		w1 := &worker{id: addr1, status: WorkerStatus_Busy, occupiedBy: &jobId0, task: task1, ch: outputCh}
		wp.pool[addr1] = w1
		wp.freeList.PushFront(w1)
		addr2 := "127.0.0.1:8083"
		w2 := &worker{id: addr2, status: WorkerStatus_Busy, occupiedBy: &jobId1, task: task2, ch: outputCh}
		wp.pool[addr2] = w2
		wp.freeList.PushFront(w2)
		addr3 := "127.0.0.1:8084"
		w3 := &worker{id: addr3, status: WorkerStatus_Idle, occupiedBy: &notOccupied, ch: outputCh}
		wp.pool[addr2] = w3
		wp.freeList.PushFront(w3)

		Convey("when remove worker", func() {
			wp.InterruptJobTasks(jobId0)

			Convey("then worker updated", func() {
				So(len(outputCh), ShouldEqual, 2)
				So((<-outputCh).Cmd, ShouldEqual, CMD_Interrupt)
				So((<-outputCh).Cmd, ShouldEqual, CMD_Interrupt)
			})
		})
	})
}
