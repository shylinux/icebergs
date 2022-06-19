package ice

import (
	"io"
	"runtime"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	log "shylinux.com/x/toolkits/logs"
)

func (m *Message) log(level string, str string, arg ...Any) *Message {
	if log.LogDisable {
		return m // 禁用日志
	}
	if str = strings.TrimSpace(kit.Format(str, arg...)); Info.Log != nil {
		Info.Log(m, m.FormatPrefix(), level, str) // 日志分流
	}

	// 日志颜色
	prefix, suffix := "", ""
	if Info.Colors {
		switch level {
		case LOG_CREATE, LOG_INSERT, LOG_MODIFY, LOG_EXPORT, LOG_IMPORT:
			prefix, suffix = "\033[36;44m", "\033[0m"
		case LOG_CMDS, LOG_START, LOG_SERVE:
			prefix, suffix = "\033[32m", "\033[0m"
		case LOG_WARN, LOG_ERROR, LOG_CLOSE:
			prefix, suffix = "\033[31m", "\033[0m"
		case LOG_AUTH, LOG_COST:
			prefix, suffix = "\033[33m", "\033[0m"
		}
	}

	// 文件行号
	switch level {
	case LOG_INFO, LOG_CMDS, "refer", "form":
	case LOG_BEGIN:
	default:
		suffix += SP + kit.FileLine(3, 3)
	}

	// 长度截断
	switch level {
	case LOG_INFO, LOG_SEND, LOG_RECV:
		if len(str) > 4096 {
			str = str[:4096]
		}
	}

	// 输出日志
	log.Info(kit.Format("%02d %9s %s%s %s%s", m.code,
		kit.Format("%4s->%-4s", m.source.Name, m.target.Name), prefix, level, str, suffix))
	return m
}
func (m *Message) join(arg ...Any) string {
	args := kit.Simple(arg...)
	list := []string{}
	for i := 0; i < len(args); i += 2 {
		if i == len(args)-1 {
			list = append(list, args[i])
		} else {
			list = append(list, strings.TrimSpace(args[i])+kit.Select("", DF, !strings.HasSuffix(strings.TrimSpace(args[i]), DF)), kit.Format(args[i+1]))
		}
	}
	return kit.Join(list, SP)
}

func (m *Message) Log(level string, str string, arg ...Any) *Message {
	return m.log(level, str, arg...)
}
func (m *Message) Info(str string, arg ...Any) *Message {
	return m.log(LOG_INFO, str, arg...)
}
func (m *Message) Cost(arg ...Any) *Message {
	list := []string{m.FormatCost(), m.join(arg...)}
	return m.log(LOG_COST, kit.Join(list, SP))
}
func (m *Message) Warn(err Any, arg ...Any) bool {
	switch err := err.(type) {
	case error:
		if err == io.EOF {
			return false
		}
		arg = append(arg, ERR, err)
	case bool:
		if !err {
			return false
		}
	case nil:
		return false
	}
	m.log(LOG_WARN, m.join(kit.Simple(arg...)))

	if len(arg) == 0 {
		arg = append(arg, "", "")
	} else if len(arg) == 1 {
		arg = append(arg, "")
	}
	m.meta[MSG_RESULT] = kit.Simple(ErrWarn, arg[0], arg[1], SP, m.join(kit.Simple(arg[2:]...)))
	return true
}
func (m *Message) Error(err bool, str string, arg ...Any) bool {
	if err {
		m.Echo(ErrWarn).Echo(str, arg...)
		m.log(LOG_ERROR, m.FormatStack(1, 100))
		m.log(LOG_ERROR, str, arg...)
		m.log(LOG_ERROR, m.FormatChain())
		return true
	}
	return false
}
func (m *Message) Debug(str string, arg ...Any) {
	if str == "" {
		str = m.FormatMeta()
	}
	m.log(LOG_DEBUG, str, arg...)
}

func (m *Message) Logs(level string, arg ...Any) *Message {
	return m.log(level, m.join(arg...))
}
func (m *Message) Log_AUTH(arg ...Any) *Message {
	return m.log(LOG_AUTH, m.join(arg...))
}
func (m *Message) Log_SEND(arg ...Any) *Message {
	return m.log(LOG_AUTH, m.join(arg...))
}
func (m *Message) Log_CREATE(arg ...Any) *Message {
	return m.log(LOG_CREATE, m.join(arg...))
}
func (m *Message) Log_REMOVE(arg ...Any) *Message {
	return m.log(LOG_REMOVE, m.join(arg...))
}
func (m *Message) Log_INSERT(arg ...Any) *Message {
	return m.log(LOG_INSERT, m.join(arg...))
}
func (m *Message) Log_DELETE(arg ...Any) *Message {
	return m.log(LOG_DELETE, m.join(arg...))
}
func (m *Message) Log_MODIFY(arg ...Any) *Message {
	return m.log(LOG_MODIFY, m.join(arg...))
}
func (m *Message) Log_SELECT(arg ...Any) *Message {
	return m.log(LOG_SELECT, m.join(arg...))
}
func (m *Message) Log_EXPORT(arg ...Any) *Message {
	return m.log(LOG_EXPORT, m.join(arg...))
}
func (m *Message) Log_IMPORT(arg ...Any) *Message {
	return m.log(LOG_IMPORT, m.join(arg...))
}

func (m *Message) FormatPrefix() string {
	return kit.Format("%s %d %s->%s", m.Time(), m.code, m.source.Name, m.target.Name)
}
func (m *Message) FormatTime() string {
	return m.Time()
}
func (m *Message) FormatShip() string {
	return kit.Format("%s->%s", m.source.Name, m.target.Name)
}
func (m *Message) FormatCost() string {
	return kit.FmtTime(kit.Int64(time.Since(m.time)))
}
func (m *Message) FormatSize() string {
	return kit.Format("%dx%d %v", m.Length(), len(m.meta[MSG_APPEND]), kit.Simple(m.meta[MSG_APPEND]))
}
func (m *Message) FormatMeta() string {
	return kit.Format(m.meta)
}
func (m *Message) FormatsMeta() string {
	return kit.Formats(m.meta)
}
func (m *Message) FormatStack(s, n int) string {
	pc := make([]uintptr, n+10)
	frames := runtime.CallersFrames(pc[:runtime.Callers(s+1, pc)])

	list := []string{}
	for {
		frame, more := frames.Next()
		file := kit.Slice(kit.Split(frame.File, PS, PS), -1)[0]
		name := kit.Slice(kit.Split(frame.Function, PS, PS), -1)[0]

		switch ls := kit.Split(name, PT, PT); kit.Select("", ls, 0) {
		case "reflect", "runtime", "http", "task", "icebergs":
		default:
			switch kit.Select("", ls, 1) {
			case "(*Frame)":
			default:
				list = append(list, kit.Format("%s:%d\t%s", file, frame.Line, name))
			}
		}

		if len(list) >= n {
			break
		}
		if !more {
			break
		}
	}
	return kit.Join(list, NL)
}
func (m *Message) FormatChain() string {
	ms := []*Message{}
	for msg := m; msg != nil; msg = msg.message {
		ms = append(ms, msg)
	}

	meta := append([]string{}, NL)
	for i := len(ms) - 1; i >= 0; i-- {
		msg := ms[i]
		meta = append(meta, kit.Format("%s %s:%d %v %s:%d %v %s:%d %v %s:%d %v", msg.FormatPrefix(),
			MSG_DETAIL, len(msg.meta[MSG_DETAIL]), msg.meta[MSG_DETAIL],
			MSG_OPTION, len(msg.meta[MSG_OPTION]), msg.meta[MSG_OPTION],
			MSG_APPEND, len(msg.meta[MSG_APPEND]), msg.meta[MSG_APPEND],
			MSG_RESULT, len(msg.meta[MSG_RESULT]), msg.meta[MSG_RESULT],
		))
		for _, k := range msg.meta[MSG_OPTION] {
			if v, ok := msg.meta[k]; ok {
				meta = append(meta, kit.Format("\t%s: %d %v", k, len(v), v))
			}
		}
		for _, k := range msg.meta[MSG_APPEND] {
			if v, ok := msg.meta[k]; ok {
				meta = append(meta, kit.Format("\t%s: %d %v", k, len(v), v))
			}
		}
	}
	return kit.Join(meta, NL)
}

func (m *Message) IsErr(arg ...string) bool {
	return len(arg) > 0 && m.Result(1) == arg[0] || m.Result(0) == ErrWarn
}
func (m *Message) IsErrNotFound() bool { return m.Result(1) == ErrNotFound }
