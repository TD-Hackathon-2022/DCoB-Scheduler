package job

import (
	"encoding/base64"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

const total = 1000000

type CalPi struct {
	id         string
	taskCnt    uint64
	funcId     string
	resultLock sync.Mutex
	sumCnt     uint64
}

func (h *CalPi) Id() string {
	return h.id
}

func (h *CalPi) GetResult() map[string]interface{} {
	return map[string]interface{}{"pi": h.getPi()}
}

func (h *CalPi) TryAdvance(fn func(task *module.Task)) (finished bool) {
	file, err := os.Open("./custom_func/monte_carlo_pi_bg.wasm")
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	task := &module.Task{
		Id:    h.id + "-task-" + strconv.FormatUint(atomic.AddUint64(&h.taskCnt, 1), 10),
		JobId: h.id,
		Ctx: &module.Context{
			InitData: base64.StdEncoding.EncodeToString(b),
		},
		FuncId:        h.funcId,
		UpdateHandler: h.handleUpdate,
	}

	fn(task)
	return false
}

func (h *CalPi) handleUpdate(task *module.Task) {
	if task.Ctx.Status == api.TaskStatus_Finished {
		finalData := task.Ctx.FinalData.(string)
		cnt, _ := strconv.ParseFloat(finalData, 6)
		atomic.AddUint64(&h.sumCnt, uint64(cnt))
		log.Infof("CalPi received result: [%s] : %d, new pi calculated as: %f", task.Id, cnt, h.getPi())
	}
}

func NewCalPi() module.Job {
	h := &CalPi{
		funcId: "custom-func-monte_carlo_pi",
	}

	h.id = "CalPi-" + strconv.Itoa(rand.Int())
	return h
}

func (h *CalPi) getPi() float64 {
	totalTried := atomic.LoadUint64(&h.taskCnt) * total
	return 4 * (float64(atomic.LoadUint64(&h.sumCnt)) / float64(totalTried))
}
