package module

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWorkerPool_ShouldAddWorkerToPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := workerPool{}

		Convey("when add worker", func() {
			addr := "127.0.0.1:8081"
			wp.add(addr)

			Convey("then pool init new worker", func() {
				w, exist := wp.pool.Load(addr)
				So(exist, ShouldBeTrue)
				So(w, ShouldHaveSameTypeAs, &worker{})
				So(w.(*worker).id, ShouldEqual, addr)
				So(w.(*worker).status, ShouldEqual, idle)
				So(w.(*worker).occupiedBy, ShouldEqual, &notOccupied)
			})
		})
	})
}

func TestWorkerPool_ShouldDoNothingWhenAddWorkerThatAlreadyInPool(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := workerPool{}
		addr := "127.0.0.1:8081"
		wp.pool.Store(addr, &worker{id: "fake-worker", status: busy})

		Convey("when add worker", func() {
			wp.add(addr)

			Convey("then do nothing", func() {
				w, _ := wp.pool.Load(addr)
				So(w.(*worker).id, ShouldEqual, "fake-worker")
				So(w.(*worker).status, ShouldEqual, busy)
			})
		})
	})
}

func TestWorkerPool_ShouldOccupyWorker(t *testing.T) {
	Convey("given worker pool", t, func() {
		wp := workerPool{}
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		wp.pool.Store(addr0, &worker{id: addr0, status: idle, occupiedBy: &job0})
		addr1 := "127.0.0.1:8082"
		wp.pool.Store(addr1, &worker{id: addr1, status: idle, occupiedBy: &notOccupied})
		addr2 := "127.0.0.1:8083"
		job1 := "job-1"
		wp.pool.Store(addr2, &worker{id: addr2, status: busy, occupiedBy: &job1})

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
		wp := workerPool{}
		addr0 := "127.0.0.1:8081"
		job0 := "job-0"
		wp.pool.Store(addr0, &worker{id: addr0, status: idle, occupiedBy: &job0})
		addr1 := "127.0.0.1:8082"
		job1 := "job-1"
		wp.pool.Store(addr1, &worker{id: addr1, status: idle, occupiedBy: &job1})
		addr2 := "127.0.0.1:8083"
		job2 := "job-2"
		wp.pool.Store(addr2, &worker{id: addr2, status: busy, occupiedBy: &job2})

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
		wp := workerPool{}

		Convey("when try occupy a worker", func() {
			_, found := wp.occupy("job-id-3")

			Convey("then no worker should be returned", func() {
				So(found, ShouldBeFalse)
			})
		})
	})
}
