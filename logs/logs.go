package logs

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
)

const (
	AdapterConsole   = "console"
	AdapterFile      = "file"
	AdapterMultiFile = "multifile"
	AdapterMail      = "smtp"
	AdapterConn      = "conn"
	AdapterEs        = "es"
	AdapterJianLiao  = "jianliao"
	AdapterSlack     = "slack"
	AdapterAliLS     = "alils"
)

var logger = NewLoggerAdapter()

type LoggerAdapter struct {
	*logs.BeeLogger
}

func NewLoggerAdapter() *LoggerAdapter {
	return &LoggerAdapter{
		BeeLogger: logs.GetBeeLogger(),
	}
}

func Close() {
	logger.Close()
}

func Reset() {
	logger.Reset()
}

func Async(msgLen ...int64) *logs.BeeLogger {
	return logger.Async(msgLen...)
}

func SetLevel(l int) {
	logger.SetLevel(l)
}

func SetPrefix(s string) {
	logger.SetPrefix(s)
}

func EnableFuncCallDepth(b bool) {
	logger.BeeLogger.EnableFuncCallDepth(b)
}

func SetLogFuncCall(b bool) {
	logger.EnableFuncCallDepth(b)
	logger.SetLogFuncCallDepth(4)
}

func SetLogFuncCallDepth(d int) {
	logger.BeeLogger.SetLogFuncCallDepth(d)
}

func SetLogger(adapter string, config ...string) error {
	return logger.SetLogger(adapter, config...)
}

func Emergency(f interface{}, v ...interface{}) {
	logger.Emergency(formatLog(f, v...))
}

func Alert(f interface{}, v ...interface{}) {
	logger.Alert(formatLog(f, v...))
}

func Critical(f interface{}, v ...interface{}) {
	logger.Critical(formatLog(f, v...))
}

func Error(f interface{}, v ...interface{}) {
	logger.Error(formatLog(f, v...))
}

func Warning(f interface{}, v ...interface{}) {
	logger.Warn(formatLog(f, v...))
}

func Warn(f interface{}, v ...interface{}) {
	logger.Warn(formatLog(f, v...))
}

func Notice(f interface{}, v ...interface{}) {
	logger.Notice(formatLog(f, v...))
}

func Informational(f interface{}, v ...interface{}) {
	logger.Info(formatLog(f, v...))
}

func Info(f interface{}, v ...interface{}) {
	logger.Info(formatLog(f, v...))
}

func Debug(f interface{}, v ...interface{}) {
	logger.Debug(formatLog(f, v...))
}

func Trace(f interface{}, v ...interface{}) {
	logger.Trace(formatLog(f, v...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
