package ice

import (
	kit "github.com/shylinux/toolkits"

	"fmt"
	"os"
	"runtime"
	"strings"
)

func (m *Message) log(level string, str string, arg ...interface{}) *Message {
	if str = strings.TrimSpace(fmt.Sprintf(str, arg...)); Log != nil {
		// 日志模块
		Log(m, level, str)
	}

	// 日志颜色
	prefix, suffix := "", ""
	switch level {
	case LOG_ENABLE, LOG_IMPORT, LOG_CREATE, LOG_INSERT, LOG_EXPORT:
		prefix, suffix = "\033[36;44m", "\033[0m"

	case LOG_LISTEN, LOG_SIGNAL, LOG_TIMERS, LOG_EVENTS:
		prefix, suffix = "\033[33m", "\033[0m"

	case LOG_CMDS, LOG_START, LOG_SERVE:
		prefix, suffix = "\033[32m", "\033[0m"
	case LOG_AUTH, LOG_COST:
		prefix, suffix = "\033[33m", "\033[0m"
	case LOG_WARN, LOG_ERROR, LOG_CLOSE:
		prefix, suffix = "\033[31m", "\033[0m"
	}

	switch level {
	case LOG_CMDS, LOG_INFO, LOG_WARN, LOG_AUTH, LOG_COST:
	default:
		_, file, line, _ := runtime.Caller(2)
		ls := strings.Split(file, "/")
		if len(ls) > 2 {
			ls = ls[len(ls)-2:]
		}
		suffix += fmt.Sprintf(" %s:%d", strings.Join(ls, "/"), line)
	}

	if os.Getenv("ctx_mod") != "" && m != nil {
		// 输出日志
		fmt.Fprintf(os.Stderr, "%s %02d %9s %s%s %s%s\n",
			m.time.Format(ICE_TIME), m.code, fmt.Sprintf("%4s->%-4s", m.source.Name, m.target.Name),
			prefix, level, str, suffix,
		)
	}
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
func (m *Message) Warn(err bool, str string, arg ...interface{}) bool {
	if err {
		_, file, line, _ := runtime.Caller(1)
		m.Echo("warn: ").Echo(str, arg...)
		return m.log(LOG_WARN, "%s:%d %s", file, line, fmt.Sprintf(str, arg...)) != nil
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
func (m *Message) Trace(key string, str string, arg ...interface{}) *Message {
	if m.Options(key) {
		m.Echo("trace: ").Echo(str, arg...)
		return m.log(LOG_TRACE, str, arg...)
	}
	return m
}
func (m *Message) Cost(str string, arg ...interface{}) *Message {
	return m.log(LOG_COST, "%s: %s", m.Format("cost"), kit.Format(str, arg...))
}

func log_fields(arg ...interface{}) string {
	list := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		list = append(list, fmt.Sprintf("%v: %v", arg[i], arg[i+1]))
	}
	return strings.Join(list, " ")
}
func (m *Message) Log_INSERT(arg ...interface{}) *Message {
	return m.log(LOG_INSERT, log_fields(arg...))
}
func (m *Message) Log_MODIFY(arg ...interface{}) *Message {
	return m.log(LOG_MODIFY, log_fields(arg...))
}
func (m *Message) Log_REMOVE(arg ...interface{}) *Message {
	return m.log(LOG_REMOVE, log_fields(arg...))
}
func (m *Message) Log_CREATE(arg ...interface{}) *Message {
	return m.log(LOG_CREATE, log_fields(arg...))
}

func (m *Message) Log_AUTH(arg ...interface{}) *Message {
	return m.log(LOG_AUTH, log_fields(arg...))
}
