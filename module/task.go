package module

type Task struct {
	Id     string
	JobId  string
	Ctx    *Context
	FuncId string
}

type taskStatus int

const (
	running taskStatus = iota
	finished
	err
	interrupt
)

type Context struct {
	status           taskStatus
	initData         interface{}
	intermediateData interface{}
	finalData        interface{}
}
