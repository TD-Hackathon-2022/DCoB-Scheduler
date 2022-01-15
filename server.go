package main

import (
	"github.com/TD-Hackathon-2022/DCoB-Scheduler/api"
	. "github.com/TD-Hackathon-2022/DCoB-Scheduler/comm"
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

	for {
		mt, recvMsg, err := c.ReadMessage()
		if err != nil {
			logger.Errorf("read: %v", err)
			break
		}

		sendMsg, err := dispatch(c.RemoteAddr(), mt, recvMsg)
		if err != nil {
			logger.Errorf("dispatch: %v", err)
			break
		}

		err = c.WriteMessage(mt, sendMsg)
		if err != nil {
			logger.Errorf("write: %v", err)
			break
		}
	}
}

var echoHandler = func(msg *api.Msg) *api.Msg { return msg }

func dispatch(addr net.Addr, messageType int, data []byte) ([]byte, error) {
	if messageType != websocket.BinaryMessage {
		return nil, errors.New("wrong message type")
	}

	msg := &api.Msg{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Wrap(err, "wrong message structure")
	}

	resMsg := &api.Msg{}
	switch msg.Cmd {
	case api.CMD_Register:
		// TODO: handle
		resMsg = echoHandler(msg)
	case api.CMD_Close:
		// TODO: handle
		resMsg = echoHandler(msg)
	case api.CMD_Status:
		// TODO: handle
		resMsg = echoHandler(msg)
	case api.CMD_Assign:
		// TODO: handle
		resMsg = echoHandler(msg)
	case api.CMD_Interrupt:
		// TODO: handle
		resMsg = echoHandler(msg)
	default:
		resMsg = echoHandler(msg)
	}

	marshaledData, err := proto.Marshal(resMsg)

	return marshaledData, errors.Wrap(err, "marshal error")
}

const addr = ":8080"

func main() {
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(addr, nil))
}
