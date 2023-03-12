package output

import (
	"fmt"
	"github.com/xtclalala/ylog"
)

var useColor = true

func SetUseColor(b bool) {
	useColor = b
}

// log stdin color
const (
	cBlack = iota + 30
	cRed
	cGreen
	cYellow
	cBlue
	cPurple
	cCyan
	cWhite
)

func Black(str string) string {
	return color(cBlack, str)
}
func Red(str string) string {
	return color(cRed, str)
}
func Yellow(str string) string {
	return color(cYellow, str)
}
func Green(str string) string {
	return color(cGreen, str)
}
func Cyan(str string) string {
	return color(cCyan, str)
}
func Blue(str string) string {
	return color(cBlue, str)
}
func Purple(str string) string {
	return color(cPurple, str)
}
func White(str string) string {
	return color(cWhite, str)
}

// string with color
func color(color int, str string) string {
	if useColor {
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", color, str)
	}
	return str
}

func SetColor() {
	ylog.AddHookLevel(ylog.ErrorLevel, []ylog.HookFn{func(entry *ylog.Entry) {
		entry.Data["type"] = Red(entry.Data["type"])
	}})
	ylog.AddHookLevel(ylog.FatalLevel, []ylog.HookFn{func(entry *ylog.Entry) {
		entry.Data["type"] = Red(entry.Data["type"])
	}})
	ylog.AddHookLevel(ylog.PanicLevel, []ylog.HookFn{func(entry *ylog.Entry) {
		entry.Data["type"] = Purple(entry.Data["type"])
	}})
	ylog.AddHookLevel(ylog.WarnLevel, []ylog.HookFn{func(entry *ylog.Entry) {
		entry.Data["type"] = Yellow(entry.Data["type"])
	}})
	ylog.AddHookLevel(ylog.InfoLevel, []ylog.HookFn{func(entry *ylog.Entry) {
		entry.Data["type"] = Green(entry.Data["type"])
	}})
}
