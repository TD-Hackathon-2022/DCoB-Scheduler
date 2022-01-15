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
				So(w.(*worker).occupiedBy, ShouldEqual, "")
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
