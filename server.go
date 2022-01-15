package main

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/comm"
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/module"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	logger = GetLogger()
)

func handle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade: %v", err)
		return
	}
	defer c.Close()

	stopCh := make(chan struct{})
	writeCh := make(chan *api.Msg)
	defer close(writeCh)
	go func() {
		for msg := range writeCh {
			marshaledData, err := proto.Marshal(msg)
			if err != nil {
				logger.Error(errors.Wrap(err, "marshal error"))
				break
			}

			err = c.WriteMessage(websocket.BinaryMessage, marshaledData)
			if err != nil {
				logger.Errorf("write: %v", err)
				break
			}
		}

		close(stopCh)
	}()

	for {
		select {
		case <-stopCh:
			break
		default:
		}

		mt, inputData, err := c.ReadMessage()
		if err != nil {
			logger.Errorf("read: %v", err)
			break
		}

		if mt != websocket.BinaryMessage {
			logger.Error(errors.New("wrong message type"))
			break
		}

		recvMsg := &api.Msg{}
		err = proto.Unmarshal(inputData, recvMsg)
		if err != nil {
			logger.Error(errors.Wrap(err, "unmarshal error"))
			break
		}

		err = dispatch(c.RemoteAddr(), recvMsg, writeCh)
		if err != nil {
			logger.Errorf("dispatch: %v", err)
			break
		}
	}
}

var echoHandler = func(msg *api.Msg) *api.Msg { return msg }

func dispatch(addr net.Addr, inputMsg *api.Msg, outputCh chan *api.Msg) error {

	resMsg := &api.Msg{}
	switch inputMsg.Cmd {
	case api.CMD_Register:
		register(addr, outputCh)
	case api.CMD_Close:
		// TODO: handle
		resMsg = echoHandler(inputMsg)
	case api.CMD_Status:
		// TODO: handle
		resMsg = echoHandler(inputMsg)
	case api.CMD_Assign:
		// TODO: handle
		resMsg = echoHandler(inputMsg)
	case api.CMD_Interrupt:
		// TODO: handle
		resMsg = echoHandler(inputMsg)
	default:
		resMsg = echoHandler(inputMsg)
	}

	outputCh <- resMsg
	return nil
}

func register(addr net.Addr, outputCh chan *api.Msg) {
	workerPool.Add(addr.String(), outputCh)
}

const addr = ":8080"

var workerPool = module.NewWorkerPool()

func main() {
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(addr, nil))
}
