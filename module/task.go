package module

import "github.com/TD-Hackathon-2022/DCoB-Scheduler/api"

type Task struct {
	Id            string
	JobId         string
	Ctx           *Context
	FuncId        string
	UpdateHandler func(*Task)
}

type Context struct {
	status           api.TaskStatus
	initData         interface{}
	intermediateData interface{}
	finalData        interface{}
}
