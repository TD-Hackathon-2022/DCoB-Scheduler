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
	Status           api.TaskStatus
	InitData         interface{}
	IntermediateData interface{}
	FinalData        interface{}
}
