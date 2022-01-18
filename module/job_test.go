package module

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"runtime"
	"testing"
	"time"
)

type MockJob struct {
	id string
	mock.Mock
}

func (f *MockJob) Id() string {
	return f.id
}

func (f *MockJob) TryAdvance(fn func(task *Task)) bool {
	args := f.Mock.Called(fn)
	return args.Get(0).(bool)
}

func (f *MockJob) GetResult() map[string]interface{} {
	return make(map[string]interface{})
}

func TestJobRunner_Submit(t *testing.T) {
	Convey("given job runner", t, func() {
		runner := NewJobRunner(make(chan *Task), &simpleStore{})

		Convey("try submit a job", func() {
			job := &MockJob{}
			runner.Submit(job)

			Convey("then job q should contains that job", func() {
				So(len(runner.jobQ), ShouldEqual, 1)
				So(<-runner.jobQ, ShouldEqual, job)
			})
		})
	})
}

func TestJobRunner_SendTask(t *testing.T) {
	Convey("given job runner", t, func() {
		runner := NewJobRunner(make(chan *Task), &simpleStore{})
		go runner.Start()
		defer runner.ShutDown()

		Convey("try submit a job", func() {
			job0 := &MockJob{id: "job0"}
			job0.Mock.On("TryAdvance", mock.Anything).Times(1).Return(true)
			job1 := &MockJob{id: "job1"}
			job1.Mock.On("TryAdvance", mock.Anything).Times(1).Return(true)
			job2 := &MockJob{id: "job2"}
			job2.Mock.On("TryAdvance", mock.Anything).Times(1).Return(true)
			runner.Submit(job0)
			runner.Submit(job1)
			runner.Submit(job2)
			time.Sleep(time.Second)

			Convey("then job q should contains that job", func() {
				So(runner.currJob, ShouldEqual, job2)
			})
		})
	})
}

func TestJobRunner_InterruptCurrentJob(t *testing.T) {
	Convey("given job runner", t, func() {
		taskQ := make(chan *Task, 1)
		runner := NewJobRunner(taskQ, &simpleStore{})
		go runner.Start()
		defer runner.ShutDown()

		Convey("try submit a job", func() {
			job := &MockJob{id: "job0"}
			job.Mock.On("TryAdvance", mock.Anything).Maybe().Return(false)
			runner.Submit(job)
			runtime.Gosched()
			runner.InterruptCurrentJob()
			time.Sleep(time.Second)

			Convey("then job q should contains that job", func() {
				task := <-taskQ
				So(task, ShouldEqual, poisonTask)
				So(task.JobId, ShouldEqual, job.id)
			})
		})
	})
}
