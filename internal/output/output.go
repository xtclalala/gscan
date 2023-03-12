package output

import (
	"errors"
	"github.com/xtclalala/ylog"
	"io"
	"time"
)

var (
	ErrLogLevelTooBig   = errors.New("log level is too big")
	ErrLogLevelTooSmall = errors.New("log level is too small")
)

// SetLogLevel 设置log输出等级
func SetLogLevel(level ylog.LogLevel) (err error) {

	if level >= ylog.DebugLevel {
		return ErrLogLevelTooBig
	}
	ylog.SetLogLevel(level)
	return
}

// AddOutputter  添加控制输出者
func AddOutputter(o io.Writer) {
	ylog.AddOuts(o)
}

// SetFormatter 设置输出方式
func SetFormatter(formatter ylog.Formatter) {
	ylog.SetFormatter(formatter)
}

func AddTime() {
	ylog.AddHook(func(level ylog.LogLevel, entry *ylog.Entry) error {
		entry.WithField("time", time.Now().Local().Format("2006-01-02 15:04:05"))
		return nil
	})

}
