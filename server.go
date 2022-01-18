package main

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/comm"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
)

var (
	log = comm.GetLogger()
)

const (
	addr             = ":8080"
	workerConnectUrl = "/connect"
)

func StartServer() {
	wh := NewWorkerHandler()

	router := gin.Default()
	router.GET(workerConnectUrl, func(c *gin.Context) { wh.handle(c.Writer, c.Request) })

	svr := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	log.Fatal(svr.ListenAndServe())
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

func NewWorkerHandler() *workerHandler {
	return &workerHandler{
		pool: module.NewWorkerPool(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func main() {
	StartServer()
}
