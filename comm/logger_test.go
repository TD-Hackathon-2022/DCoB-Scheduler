package comm

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type fakeWriter struct {
	bBuf []byte
}

func (f *fakeWriter) Write(p []byte) (n int, err error) {
	f.bBuf = p
	return len(p), nil
}

func Test_initLogger(t *testing.T) {
	Convey("given info level and fake writer", t, func() {
		level := "info"
		writer := &fakeWriter{}

		Convey("when init logger", func() {
			initLogger(func(l *loggerOption) { l.loggerFile = writer }, WithLoggerLevel(level))

			Convey("then info msg will present", func() {

				sugarLogger.Infow("info")
				_ = sugarLogger.Sync()

				So(string(writer.bBuf), ShouldNotEqual, "")
			})

			Convey("then debug msg will not present", func() {

				sugarLogger.Debugw("debug")
				_ = sugarLogger.Sync()

				So(string(writer.bBuf), ShouldEqual, "")
			})
		})

	})
}
