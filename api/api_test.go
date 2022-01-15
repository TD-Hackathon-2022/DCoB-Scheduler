package api

import (
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestApi_ShouldMarshalThenUnmarshal(t *testing.T) {
	Convey("given msg", t, func() {
		msg := &Msg{
			Cmd: CMD_Status,
			Payload: &Msg_Status{
				Status: &StatusPayload{
					WorkStatus: WorkerStatus_Busy,
					TaskId:     "fake-task",
					TaskStatus: TaskStatus_Running,
					ExecResult: "fake-result",
				},
			},
		}

		Convey("when marshal", func() {
			data, err := proto.Marshal(msg)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			Convey("should get same msg when unmarshal", func() {
				resMsg := &Msg{}
				err := proto.Unmarshal(data, resMsg)
				if err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}

				So(proto.Equal(resMsg, msg), ShouldBeTrue)
			})
		})
	})
}
