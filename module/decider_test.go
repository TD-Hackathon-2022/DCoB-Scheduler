package module

import (
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	. "github.com/smartystreets/goconvey/convey"
	"runtime"
	"testing"
)

func TestDecider_ShouldOccupyAndAssignTaskToWorker(t *testing.T) {
	Convey("given decider", t, func() {
		jobId := "fake-job-id"
		task := &Task{
			Id:    "fake-task",
			JobId: jobId,
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
		}

		addr := "127.0.0.1:8081"
		w := &worker{id: addr, status: WorkerStatus_Idle, occupiedBy: &notOccupied}
		wp := NewWorkerPool()
		wp.pool[addr] = w
		wp.freeList.PushFront(w)

		taskQ := make(chan *Task, 1)
		taskQ <- task

		decider := NewDecider(wp, taskQ)

		Convey("when decider start", func() {
			go decider.Start()
			runtime.Gosched()

			Convey("then worker should be occupied and assigned", func() {
				So(*w.occupiedBy, ShouldEqual, jobId)
				So(w.task, ShouldEqual, task)
				So(w.statusNotify, ShouldEqual, decider.statusNotify)
			})
		})

		close(taskQ)
	})
}

func TestDecider_ShouldUpdateTaskStatusThenCallTaskHandlerWhenNotify(t *testing.T) {
	Convey("given decider", t, func() {
		notified := false
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				Status:   TaskStatus_Running,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
			UpdateHandler: func(*Task) {
				notified = true
			},
		}

		wp := NewWorkerPool()
		decider := NewDecider(wp, nil)

		Convey("when notify with finished status", func() {
			w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task}

			payload := &StatusPayload{
				TaskStatus: TaskStatus_Finished,
				ExecResult: "fake-result",
			}
			decider.statusNotify(w, payload)

			Convey("then worker updated", func() {
				So(task.Ctx.Status, ShouldEqual, TaskStatus_Finished)
				So(task.Ctx.FinalData, ShouldEqual, payload.ExecResult)
				So(notified, ShouldBeTrue)
			})
		})

		Convey("when notify with running status", func() {
			w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task}

			payload := &StatusPayload{
				TaskStatus: TaskStatus_Running,
				ExecResult: "fake-result",
			}
			decider.statusNotify(w, payload)

			Convey("then worker updated", func() {
				So(task.Ctx.Status, ShouldEqual, TaskStatus_Running)
				So(task.Ctx.IntermediateData, ShouldEqual, payload.ExecResult)
				So(notified, ShouldBeTrue)
			})
		})
	})
}

func TestDecider_ShouldReleaseWorkerWhenNotifyWithTaskNotRunningStatus(t *testing.T) {
	Convey("given decider", t, func() {
		notified := false
		task := &Task{
			Id:    "fake-task",
			JobId: "fake-job-id",
			Ctx: &Context{
				Status:   TaskStatus_Finished,
				InitData: "fake-data",
			},
			FuncId: "fake-func-id",
			UpdateHandler: func(*Task) {
				notified = true
			},
		}

		wp := NewWorkerPool()
		decider := NewDecider(wp, nil)

		Convey("when notify", func() {
			w := &worker{id: "127.0.0.1:8081", status: WorkerStatus_Busy, occupiedBy: &task.JobId, task: task}
			decider.statusNotify(w, &StatusPayload{
				TaskStatus: TaskStatus_Error,
			})

			Convey("then worker updated", func() {
				So(notified, ShouldBeTrue)
				So(w.status, ShouldEqual, WorkerStatus_Idle)
				So(w.occupiedBy, ShouldEqual, &notOccupied)
			})
		})
	})
}
