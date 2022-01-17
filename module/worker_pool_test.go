package module

import (
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWorkerPool_ShouldAddWorkerToPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := WorkerPool{}

		Convey("when add worker", func() {
			addr := "127.0.0.1:8081"
			wp.Add(addr, nil)

			Convey("then pool init new worker", func() {
				w, exist := wp.pool.Load(addr)
				So(exist, ShouldBeTrue)
				So(w, ShouldHaveSameTypeAs, &worker{})
				So(w.(*worker).id, ShouldEqual, addr)
				So(w.(*worker).status, ShouldEqual, WorkerStatus_Idle)
				So(w.(*worker).occupiedBy, ShouldEqual, &notOccupied)
			})
		})
	})
}

func TestWorkerPool_ShouldDoNothingWhenAddWorkerThatAlreadyInPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := WorkerPool{}
		addr := "127.0.0.1:8081"
		wp.pool.Store(addr, &worker{id: "fake-worker", status: WorkerStatus_Busy})

		Convey("when add worker", func() {
			wp.Add(addr, nil)

			Convey("then do nothing", func() {
				w, _ := wp.pool.Load(addr)
				So(w.(*worker).id, ShouldEqual, "fake-worker")
				So(w.(*worker).status, ShouldEqual, WorkerStatus_Busy)
			})
		})
	})
}

func TestWorkerPool_ShouldOccupyWorker(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := WorkerPool{}
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		wp.pool.Store(addr0, &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &job0})
		addr1 := "127.0.0.1:8082"
		wp.pool.Store(addr1, &worker{id: addr1, status: WorkerStatus_Idle, occupiedBy: &notOccupied})
		addr2 := "127.0.0.1:8083"
		job1 := "job-1"
		wp.pool.Store(addr2, &worker{id: addr2, status: WorkerStatus_Idle, occupiedBy: &job1})

		Convey("when try occupy a worker", func() {
			w, found := wp.occupy("job-id-2")

			Convey("then worker1 should be returned", func() {
				So(found, ShouldBeTrue)
				So(w.id, ShouldEqual, addr1)
				So(*w.occupiedBy, ShouldEqual, "job-id-2")
			})
		})
	})
}

func TestWorkerPool_ShouldNotOccupyWorkerIfNotAvailable(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := WorkerPool{}
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		wp.pool.Store(addr0, &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &job0})
		addr1 := "127.0.0.1:8082"
		job1 := "job-1"
		wp.pool.Store(addr1, &worker{id: addr1, status: WorkerStatus_Idle, occupiedBy: &job1})
		addr2 := "127.0.0.1:8083"
		job2 := "job-2"
		wp.pool.Store(addr2, &worker{id: addr2, status: WorkerStatus_Busy, occupiedBy: &job2})

		Convey("when try occupy a worker", func() {
			_, found := wp.occupy("job-id-3")

			Convey("then no worker should be returned", func() {
				So(found, ShouldBeFalse)
			})
		})
	})
}

func TestWorkerPool_ShouldNotOccupyWorkerIfNoWorker(t *testing.T) {
	Convey("given empty worker pool", t, func() {
		wp := WorkerPool{}

		Convey("when try occupy a worker", func() {
			_, found := wp.occupy("job-id-3")

			Convey("then no worker should be returned", func() {
				So(found, ShouldBeFalse)
			})
		})
	})
}

func TestWorkerPool_ShouldOccupyWorkerThenRelease(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := WorkerPool{}
		addr0 := "127.0.0.1:8081"
		wp.pool.Store(addr0, &worker{id: addr0, status: WorkerStatus_Idle, occupiedBy: &notOccupied})

		Convey("when try occupy a worker", func() {
			w, found := wp.occupy("job-id-0")

			Convey("then worker1 should be returned", func() {
				So(found, ShouldBeTrue)
				So(w.id, ShouldEqual, addr0)
				So(*w.occupiedBy, ShouldEqual, "job-id-0")
			})

			wp.release(w)

			Convey("can release", func() {
				So(w.status, ShouldEqual, WorkerStatus_Idle)
				So(w.occupiedBy, ShouldEqual, &notOccupied)
			})
		})
	})
}

func TestWorker_ShouldNotAssignTaskWhenNotOccupied(t *testing.T) {
	Convey("given worker", t, func() {
		w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Idle, occupiedBy: &notOccupied}

		Convey("when try assign a task", func() {
			success := w.assign(&Task{}, func(*worker) {})

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
			success := w.assign(&Task{JobId: job1}, func(*worker) {})

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
			Ctx: &Context{status: TaskStatus_Running},
		}}

		Convey("when try assign a task", func() {
			success := w.assign(&Task{JobId: job0}, func(*worker) {})

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
					initData: "fake-data",
				},
				FuncId: funcId,
			}, func(*worker) {})

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
				initData: "fake-data",
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

func TestWorkerPool_ShouldUpdateWorkerAndTaskStatus(t *testing.T) {
	Convey("given worker pool", t, func() {
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				status:   TaskStatus_Running,
				initData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		wp := WorkerPool{}
		addr := "127.0.0.1:8081"
		notified := false
		wp.pool.Store(addr, &worker{id: addr, status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task, notify: func(*worker) {
			notified = true
		}})

		Convey("when update status", func() {
			status := &StatusPayload{
				WorkStatus: WorkerStatus_Idle,
				TaskId:     task.Id,
				TaskStatus: TaskStatus_Finished,
				ExecResult: "success",
			}

			_ = wp.UpdateStatus(addr, status)

			Convey("then worker updated", func() {
				w, _ := wp.pool.Load(addr)
				So(w.(*worker).id, ShouldEqual, addr)
				So(w.(*worker).status, ShouldEqual, WorkerStatus_Idle)
				So(w.(*worker).task.Ctx.status, ShouldEqual, TaskStatus_Finished)
				So(w.(*worker).task.Ctx.finalData, ShouldEqual, status.ExecResult)
				So(notified, ShouldBeTrue)
			})
		})
	})
}

func TestWorkerPool_ShouldRemoveWorkerAndNotifyTask(t *testing.T) {
	Convey("given worker pool", t, func() {
		notified := false
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				status:   TaskStatus_Running,
				initData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		wp := WorkerPool{}
		addr := "127.0.0.1:8081"
		wp.pool.Store(addr, &worker{id: addr, status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task, notify: func(*worker) {
			notified = true
		}})

		Convey("when remove worker", func() {
			wp.Remove(addr)

			Convey("then worker updated", func() {
				_, exist := wp.pool.Load(addr)
				So(exist, ShouldBeFalse)
				So(task.Ctx.status, ShouldEqual, TaskStatus_Interrupted)
				So(notified, ShouldBeTrue)
			})
		})
	})
}
