package main

import (
	. "DCoB-Scheduler/comm"
	"github.com/gorilla/websocket"
	"log"
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

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade: %v", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logger.Errorf("read: %v", err)
			break
		}

		GetLogger().Infof("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			logger.Errorf("write: %v", err)
			break
		}
	}
}

const addr = ":8080"

func main() {
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(addr, nil))
}
