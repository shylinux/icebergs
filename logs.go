package ice

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	log "shylinux.com/x/toolkits/logs"
)

var _log_disable = true
var Log func(m *Message, p, l, s string)

func (m *Message) log(level string, str string, arg ...interface{}) *Message {
	if _log_disable {
		return m // 禁用日志
	}
	if str = strings.TrimSpace(kit.Format(str, arg...)); Log != nil {
		Log(m, m.Format(kit.MDB_PREFIX), level, str) // 日志分流
	}

	// 日志颜色
	prefix, suffix := "", ""
	switch level {
	case LOG_IMPORT, LOG_CREATE, LOG_INSERT, LOG_MODIFY, LOG_EXPORT:
		prefix, suffix = "\033[36;44m", "\033[0m"
	case LOG_CMDS, LOG_START, LOG_SERVE:
		prefix, suffix = "\033[32m", "\033[0m"
	case LOG_WARN, LOG_CLOSE, LOG_ERROR:
		prefix, suffix = "\033[31m", "\033[0m"
	case LOG_AUTH, LOG_COST:
		prefix, suffix = "\033[33m", "\033[0m"
	}

	// 文件行号
	switch level {
	case LOG_CMDS, LOG_INFO, "refer", "form":
	case LOG_BEGIN:
	default:
		suffix += " " + kit.FileLine(3, 3)
	}

	// 长度截断
	switch level {
	case LOG_INFO, "send", "recv":
		if len(str) > 1024 {
			str = str[:1024]
		}
	}

	// 输出日志
	log.Info(fmt.Sprintf("%02d %9s %s%s %s%s", m.code,
		fmt.Sprintf("%4s->%-4s", m.source.Name, m.target.Name), prefix, level, str, suffix))
	return m
}
func (m *Message) join(arg ...interface{}) string {
	list := []string{}
	for i := 0; i < len(arg); i += 2 {
		if i == len(arg)-1 {
			list = append(list, kit.Format(arg[i]))
		} else {
			list = append(list, kit.Format(arg[i])+":", kit.Format(arg[i+1]))
		}
	}
	return strings.Join(list, " ")
}

func (m *Message) Log(level string, str string, arg ...interface{}) *Message {
	return m.log(level, str, arg...)
}
func (m *Message) Info(str string, arg ...interface{}) *Message {
	return m.log(LOG_INFO, str, arg...)
}
func (m *Message) Cost(arg ...interface{}) *Message {
	list := []string{m.FormatCost(), m.join(arg...)}
	return m.log(LOG_COST, strings.Join(list, " "))
}
func (m *Message) Warn(err bool, arg ...interface{}) bool {
	if !err || len(m.meta[MSG_RESULT]) > 0 && m.meta[MSG_RESULT][0] == ErrWarn {
		return err
	}

	m.meta[MSG_RESULT] = kit.Simple(ErrWarn, arg)
	m.log(LOG_WARN, m.join(kit.Simple(arg)))
	return err
}
func (m *Message) Error(err bool, str string, arg ...interface{}) bool {
	if err {
		m.Echo("error: ").Echo(str, arg...)
		m.log(LOG_ERROR, m.FormatStack())
		m.log(LOG_ERROR, str, arg...)
		m.log(LOG_ERROR, m.FormatChain())
		return true
	}
	return false
}
func (m *Message) Debug(str string, arg ...interface{}) {
	m.log(LOG_DEBUG, str, arg...)
}

func (m *Message) Logs(level string, arg ...interface{}) *Message {
	return m.log(level, m.join(arg...))
}
func (m *Message) Log_AUTH(arg ...interface{}) *Message {
	return m.log(LOG_AUTH, m.join(arg...))
}
func (m *Message) Log_SEND(arg ...interface{}) *Message {
	return m.log(LOG_AUTH, m.join(arg...))
}
func (m *Message) Log_CREATE(arg ...interface{}) *Message {
	return m.log(LOG_CREATE, m.join(arg...))
}
func (m *Message) Log_REMOVE(arg ...interface{}) *Message {
	return m.log(LOG_REMOVE, m.join(arg...))
}
func (m *Message) Log_INSERT(arg ...interface{}) *Message {
	return m.log(LOG_INSERT, m.join(arg...))
}
func (m *Message) Log_DELETE(arg ...interface{}) *Message {
	return m.log(LOG_DELETE, m.join(arg...))
}
func (m *Message) Log_MODIFY(arg ...interface{}) *Message {
	return m.log(LOG_MODIFY, m.join(arg...))
}
func (m *Message) Log_SELECT(arg ...interface{}) *Message {
	return m.log(LOG_SELECT, m.join(arg...))
}
func (m *Message) Log_EXPORT(arg ...interface{}) *Message {
	return m.log(LOG_EXPORT, m.join(arg...))
}
func (m *Message) Log_IMPORT(arg ...interface{}) *Message {
	return m.log(LOG_IMPORT, m.join(arg...))
}

func (m *Message) FormatStack() string {
	// 调用栈
	pc := make([]uintptr, 100)
	pc = pc[:runtime.Callers(5, pc)]
	frames := runtime.CallersFrames(pc)

	meta := []string{}
	for {
		frame, more := frames.Next()
		file := strings.Split(frame.File, "/")
		name := strings.Split(frame.Function, "/")
		meta = append(meta, fmt.Sprintf("\n%s:%d\t%s", file[len(file)-1], frame.Line, name[len(name)-1]))
		if !more {
			break
		}
	}
	return strings.Join(meta, "")
}
func (m *Message) FormatChain() string {
	ms := []*Message{}
	for msg := m; msg != nil; msg = msg.message {
		ms = append(ms, msg)
	}

	meta := append([]string{}, "\n\n")
	for i := len(ms) - 1; i >= 0; i-- {
		msg := ms[i]

		meta = append(meta, fmt.Sprintf("%s ", msg.Format("prefix")))
		if len(msg.meta[MSG_DETAIL]) > 0 {
			meta = append(meta, fmt.Sprintf("detail:%d %v", len(msg.meta[MSG_DETAIL]), msg.meta[MSG_DETAIL]))
		}

		if len(msg.meta[MSG_OPTION]) > 0 {
			meta = append(meta, fmt.Sprintf("option:%d %v\n", len(msg.meta[MSG_OPTION]), msg.meta[MSG_OPTION]))
			for _, k := range msg.meta[MSG_OPTION] {
				if v, ok := msg.meta[k]; ok {
					meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
				}
			}
		} else {
			meta = append(meta, "\n")
		}

		if len(msg.meta[MSG_APPEND]) > 0 {
			meta = append(meta, fmt.Sprintf("  append:%d %v\n", len(msg.meta[MSG_APPEND]), msg.meta[MSG_APPEND]))
			for _, k := range msg.meta[MSG_APPEND] {
				if v, ok := msg.meta[k]; ok {
					meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
				}
			}
		}
		if len(msg.meta[MSG_RESULT]) > 0 {
			meta = append(meta, fmt.Sprintf("  result:%d %v\n", len(msg.meta[MSG_RESULT]), msg.meta[MSG_RESULT]))
		}
	}
	return strings.Join(meta, "")
}
func (m *Message) FormatTime() string {
	return m.Format("time")
}
func (m *Message) FormatMeta() string {
	return m.Format("meta")
}
func (m *Message) FormatSize() string {
	return m.Format("size")
}
func (m *Message) FormatCost() string {
	return m.Format("cost")
}
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case []byte:
		json.Unmarshal(key, &m.meta)
	case string:
		switch key {
		case "cost":
			return kit.FmtTime(kit.Int64(time.Since(m.time)))
		case "meta":
			return kit.Format(m.meta)
		case "size":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d", 0, 0)
			} else {
				return fmt.Sprintf("%dx%d", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]))
			}
		case "append":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d %s", 0, 0, "[]")
			} else {
				return fmt.Sprintf("%dx%d %s", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]), kit.Format(m.meta["append"]))
			}

		case "time":
			return m.Time()
		case "ship":
			return fmt.Sprintf("%s->%s", m.source.Name, m.target.Name)
		case "prefix":
			return fmt.Sprintf("%s %d %s->%s", m.Time(), m.code, m.source.Name, m.target.Name)

		case "chain":
			return m.FormatChain()
		case "stack":
			return m.FormatStack()
		}
	}
	return m.time.Format(MOD_TIME)
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	}
	return m.Format(key)
}
