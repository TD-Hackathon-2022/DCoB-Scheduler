package job

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestHashMiner_ShouldCreateTaskAndOutputResult(t *testing.T) {
	Convey("given hash miner", t, func() {
		miner := NewHashMiner(2)

		Convey("when try advance", func() {
			tasks := make([]*module.Task, 0, 2)
			fn := func(task *module.Task) {
				tasks = append(tasks, task)
			}

			miner.TryAdvance(fn)
			miner.TryAdvance(fn)

			// finish task
			for _, task := range tasks {
				task.Ctx.Status = api.TaskStatus_Finished
				task.Ctx.FinalData = "006c6f636b436861696ee3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
				task.UpdateHandler(task)
			}

			Convey("then get output", func() {
				So(strings.Contains(tasks[0].Id, "-task-1"), ShouldBeTrue)
				So(strings.Contains(tasks[1].Id, "-task-2"), ShouldBeTrue)

				res := miner.GetResult()
				So(res[tasks[0].Id], ShouldEqual, "006c6f636b436861696ee3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
				So(res[tasks[1].Id], ShouldEqual, "006c6f636b436861696ee3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
				So(len(res), ShouldEqual, 2)
			})
		})
	})
}
