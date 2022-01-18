package job

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
)

type HashMiner struct {
	id         string
	taskCnt    uint64
	funcId     string
	difficulty int
	resultLock sync.Mutex
	resultMap  map[string]string
}

func (h *HashMiner) Id() string {
	return h.id
}

func (h *HashMiner) GetResult() map[string]interface{} {
	res := make(map[string]interface{})
	h.resultLock.Lock()
	defer h.resultLock.Unlock()
	for k, v := range h.resultMap {
		res[k] = v
	}

	return res
}

func (h *HashMiner) TryAdvance(fn func(task *module.Task)) (finished bool) {
	task := &module.Task{
		Id:    h.id + "-task-" + strconv.FormatUint(atomic.AddUint64(&h.taskCnt, 1), 10),
		JobId: h.id,
		Ctx: &module.Context{
			InitData: strconv.Itoa(h.difficulty),
		},
		FuncId:        h.funcId,
		UpdateHandler: h.handleUpdate,
	}

	fn(task)
	return false
}

func (h *HashMiner) handleUpdate(task *module.Task) {
	if task.Ctx.Status == api.TaskStatus_Finished {
		h.resultLock.Lock()
		defer h.resultLock.Unlock()
		h.resultMap[task.Id] = task.Ctx.FinalData.(string)
	}
}

func NewHashMiner(difficulty int) module.Job {
	h := &HashMiner{
		funcId:     "hash-miner",
		difficulty: difficulty,
		resultMap:  make(map[string]string),
	}

	h.id = "HashMiner-" + strconv.Itoa(rand.Int())
	return h
}
