package ice

import (
	kit "github.com/shylinux/toolkits"
	log "github.com/shylinux/toolkits/logs"

	"fmt"
	"strings"
)

var ErrWarn = "warn: "
var ErrNotAuth = "not auth: "
var ErrNotJoin = "not join: "
var ErrNotFound = "not found: "
var ErrNotEnough = "not enough: "
var ErrStart = "err start: "

func (m *Message) log(level string, str string, arg ...interface{}) *Message {
	if str = strings.TrimSpace(kit.Format(str, arg...)); Log != nil {
		// 日志模块
		Log(m, level, str)
	}

	// 日志颜色
	prefix, suffix := "", ""
	switch level {
	case LOG_ENABLE, LOG_IMPORT, LOG_CREATE, LOG_INSERT, LOG_MODIFY, LOG_EXPORT:
		prefix, suffix = "\033[36;44m", "\033[0m"

	case LOG_LISTEN, LOG_SIGNAL, LOG_TIMERS, LOG_EVENTS:
		prefix, suffix = "\033[33m", "\033[0m"

	case LOG_CMDS, LOG_START, LOG_SERVE:
		prefix, suffix = "\033[32m", "\033[0m"
	case LOG_AUTH, LOG_CONF, LOG_COST:
		prefix, suffix = "\033[33m", "\033[0m"
	case LOG_WARN, LOG_ERROR, LOG_CLOSE:
		prefix, suffix = "\033[31m", "\033[0m"
	}

	// 文件行号
	switch level {
	// case LOG_CMDS, LOG_INFO, LOG_WARN, "refer", "form":
	case LOG_CMDS, LOG_INFO, "refer", "form":
	case "register", "begin":
	default:
		suffix += " " + kit.FileLine(3, 2)
	}
	if len(BinPack) > 0 {
		prefix, suffix = "", ""
	}

	switch level {
	case LOG_INFO, "send", "recv":
		if len(str) > 1000 {
			str = str[:1000]
		}
	}
	// 输出日志
	log.Info(fmt.Sprintf("%02d %9s %s%s %s%s", m.code, fmt.Sprintf("%4s->%-4s", m.source.Name, m.target.Name),
		prefix, level, str, suffix))
	return m
}
func (m *Message) Log(level string, str string, arg ...interface{}) *Message {
	return m.log(level, str, arg...)
}
func (m *Message) Logs(level string, arg ...interface{}) *Message {
	list := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		list = append(list, fmt.Sprintf("%v: %v", arg[i], arg[i+1]))
	}
	m.log(level, strings.Join(list, " "))
	return m
}
func (m *Message) Info(str string, arg ...interface{}) *Message {
	return m.log(LOG_INFO, str, arg...)
}
func (m *Message) Warn(err bool, arg ...interface{}) bool {
	if err {
		list := kit.Simple(arg...)
		if len(list) > 1 || len(m.meta[MSG_RESULT]) > 0 && m.meta[MSG_RESULT][0] != ErrWarn {
			m.meta[MSG_RESULT] = append([]string{ErrWarn}, list...)
		}
		return m.log(LOG_WARN, fmt.Sprint(arg...)) != nil
	}
	return false
}
func (m *Message) Error(err bool, str string, arg ...interface{}) bool {
	if err {
		m.Echo("error: ").Echo(str, arg...)
		m.log(LOG_ERROR, m.Format("stack"))
		m.log(LOG_ERROR, str, arg...)
		m.log(LOG_ERROR, m.Format("chain"))
		return true
	}
	return false
}
func (m *Message) Debug(str string, arg ...interface{}) {
	m.log(LOG_DEBUG, str, arg...)
}

func log_fields(arg ...interface{}) string {
	list := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		list = append(list, fmt.Sprintf("%v: %v", arg[i], arg[i+1]))
	}
	return strings.Join(list, " ")
}
func (m *Message) Log_IMPORT(arg ...interface{}) *Message {
	return m.log(LOG_IMPORT, log_fields(arg...))
}
func (m *Message) Log_EXPORT(arg ...interface{}) *Message {
	return m.log(LOG_EXPORT, log_fields(arg...))
}
func (m *Message) Log_CREATE(arg ...interface{}) *Message {
	return m.log(LOG_CREATE, log_fields(arg...))
}
func (m *Message) Log_REMOVE(arg ...interface{}) *Message {
	return m.log(LOG_REMOVE, log_fields(arg...))
}
func (m *Message) Log_INSERT(arg ...interface{}) *Message {
	return m.log(LOG_INSERT, log_fields(arg...))
}
func (m *Message) Log_DELETE(arg ...interface{}) *Message {
	return m.log(LOG_DELETE, log_fields(arg...))
}
func (m *Message) Log_SELECT(arg ...interface{}) *Message {
	return m.log(LOG_SELECT, log_fields(arg...))
}
func (m *Message) Log_MODIFY(arg ...interface{}) *Message {
	return m.log(LOG_MODIFY, log_fields(arg...))
}

func (m *Message) Log_CONF(arg ...interface{}) *Message {
	return m.log(LOG_CONF, log_fields(arg...))
}
func (m *Message) Log_AUTH(arg ...interface{}) *Message {
	return m.log(LOG_AUTH, log_fields(arg...))
}
func (m *Message) Cost(str string, arg ...interface{}) *Message {
	return m.log(LOG_COST, "%s: %s", m.Format("cost"), kit.Format(str, arg...))
}
