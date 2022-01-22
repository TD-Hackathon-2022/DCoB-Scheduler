package main

import (
	"context"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/comm"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module/job"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	log = comm.GetLogger()
)

const (
	addr                     = ":8080"
	workerConnectUrl         = "/connect"
	taskQueueCapacity        = 128
	adminStartUrl            = "/admin/start"
	adminShutdownUrl         = "/admin/shutdown"
	adminRunMineJobUrl       = "/admin/job/run-mine"
	adminRunCalPiJobUrl      = "/admin/job/run-pi"
	adminInterruptCurrJobUrl = "/admin/job/interrupt-curr"
	adminGetJobResultUrl     = "/admin/job/:id"
)

func BuildServer(wh *workerHandler, ah *adminHandler) *http.Server {
	router := gin.Default()
	router.GET(workerConnectUrl, func(c *gin.Context) { wh.handle(c.Writer, c.Request) })
	router.POST(adminStartUrl, ah.start)
	router.POST(adminShutdownUrl, ah.shutdown)
	router.POST(adminRunMineJobUrl, ah.runMinerJob)
	router.POST(adminRunCalPiJobUrl, ah.runCalPiJob)
	router.POST(adminInterruptCurrJobUrl, ah.interruptCurrentJob)
	router.GET(adminGetJobResultUrl, ah.getJobInfo)
	router.Static("/ui", "./ui")

	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

type workerHandler struct {
	pool     *module.WorkerPool
	upgrader websocket.Upgrader
}

func (h *workerHandler) handle(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("upgrade: %v", err)
		return
	}
	defer func() {
		// clean worker pool when connection exit
		h.pool.Remove(c.RemoteAddr().String())
		_ = c.Close()
		log.Debugf("Connection closed: %s", c.RemoteAddr())
	}()
	log.Debugf("Connection established: %s", c.RemoteAddr())

	writeCh := make(chan *api.Msg)
	defer close(writeCh)

	go h.handleSend(c, writeCh)
	h.handleRecv(c, writeCh)
}

func (h *workerHandler) handleSend(c *websocket.Conn, writeCh chan *api.Msg) {
	for msg := range writeCh {
		marshaledData, err := proto.Marshal(msg)
		if err != nil {
			log.Error(errors.Wrap(err, "marshal error"))
			break
		}

		err = c.WriteMessage(websocket.BinaryMessage, marshaledData)
		if err != nil {
			log.Errorf("write: %v", err)
			break
		}
		log.Debugf("Msg sent: %v", msg)
	}
}

func (h *workerHandler) handleRecv(c *websocket.Conn, writeCh chan *api.Msg) {
	for {
		mt, inputData, err := c.ReadMessage()
		if err != nil {
			log.Errorf("read: %v", err)
			return
		}

		if mt != websocket.BinaryMessage {
			log.Error(errors.New("wrong message type"))
			return
		}

		recvMsg := &api.Msg{}
		err = proto.Unmarshal(inputData, recvMsg)
		if err != nil {
			log.Error(errors.Wrap(err, "unmarshal error"))
			return
		}

		log.Debugf("Msg recieved: %v", recvMsg)
		err = h.dispatch(c.RemoteAddr(), recvMsg, writeCh)
		if err != nil {
			log.Errorf("dispatch: %v", err)
			return
		}
	}
}

func (h *workerHandler) dispatch(addr net.Addr, inputMsg *api.Msg, outputCh chan *api.Msg) (err error) {
	switch inputMsg.Cmd {
	case api.CMD_Register:
		h.pool.Add(addr.String(), outputCh)
	case api.CMD_Close:
		// TODO: handle
		outputCh <- inputMsg
	case api.CMD_Status:
		err = h.pool.UpdateStatus(addr.String(), inputMsg.GetStatus())
	default:
		outputCh <- inputMsg
	}

	return err
}

func NewWorkerHandler(pool *module.WorkerPool) *workerHandler {
	return &workerHandler{
		pool: pool,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type adminHandler struct {
	jobRunner *module.JobRunner
}

func (h *adminHandler) start(_ *gin.Context) {
	go h.jobRunner.Start()
}

func (h *adminHandler) shutdown(_ *gin.Context) {
	h.jobRunner.ShutDown()
}

func (h *adminHandler) runMinerJob(c *gin.Context) {
	var difficulty = 5
	if d, err := strconv.Atoi(c.Request.URL.Query().Get("difficulty")); err == nil {
		difficulty = d
	}

	minerJob := job.NewHashMiner(difficulty)
	h.jobRunner.Submit(minerJob)

	c.JSON(http.StatusCreated, minerJob.Id())
}

func (h *adminHandler) runCalPiJob(c *gin.Context) {
	calPi := job.NewCalPi()
	h.jobRunner.Submit(calPi)

	c.JSON(http.StatusCreated, calPi.Id())
}

func (h *adminHandler) interruptCurrentJob(_ *gin.Context) {
	h.jobRunner.InterruptCurrentJob()
}

type uiData struct {
	Coins  int   `json:"coins,omitempty"`
	Hashes int   `json:"hashes,omitempty"`
	Now    int64 `json:"now,omitempty"`
}

func (h *adminHandler) getJobInfo(c *gin.Context) {
	jobId := c.Param("id")
	j, exist := h.jobRunner.GetJobById(jobId)
	if !exist {
		c.Status(http.StatusNotFound)
		return
	}

	result := j.GetResult()
	switch j.(type) {
	case *job.HashMiner:
		d := &uiData{Coins: len(result), Hashes: len(result), Now: time.Now().Unix()}
		/*d := &uiData{Hashes: make([]string, 0, len(result)), Now: time.Now().Unix()}
		for _, v := range result {
			hashBytes, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				continue
			}
			d.Hashes = append(d.Hashes, fmt.Sprintf("%x", hashBytes))
		}*/
		c.JSON(http.StatusOK, d)
	case *job.CalPi:
		c.JSON(http.StatusOK, result)
	}
}

func NewAdminHandler(taskQ chan<- *module.Task, store module.JobStore) *adminHandler {
	return &adminHandler{
		jobRunner: module.NewJobRunner(taskQ, store),
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	taskQ := make(chan *module.Task, taskQueueCapacity)
	pool := module.NewWorkerPool()
	decider := module.NewDecider(pool, taskQ)
	go decider.Start()

	store := module.NewSimpleStore()
	svr := BuildServer(
		NewWorkerHandler(pool),
		NewAdminHandler(taskQ, store))
	go log.Fatal(svr.ListenAndServe())

	waitToShutdown(svr)
}

func waitToShutdown(server *http.Server) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
