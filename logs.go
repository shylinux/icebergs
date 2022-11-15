package ice

import (
	"io"
	"runtime"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func (m *Message) join(arg ...Any) (string, []Any) {
	list, meta := []string{}, []Any{}
	for i := 0; i < len(arg); i += 2 {
		switch v := arg[i].(type) {
		case logs.Meta:
			meta = append(meta, v)
			i--
			continue
		case []string:
			list = append(list, v...)
			i--
			continue
		}
		key := strings.TrimSpace(kit.Format(arg[i]))
		if i == len(arg)-1 {
			list = append(list, key)
			continue
		}
		switch v := arg[i+1].(type) {
		case logs.Meta:
			list = append(list, key)
			meta = append(meta, v)
			continue
		case time.Time:
			arg[i+1] = v.Format(MOD_TIME)
		}
		list = append(list, key+kit.Select("", DF, !strings.HasSuffix(key, DF)), kit.Format(kit.Select("", kit.Simple(arg[i+1]), 0)))
	}
	return kit.Join(list, SP), meta
}
func (m *Message) log(level string, str string, arg ...Any) *Message {
	_source := logs.FileLineMeta(logs.FileLine(3, 3))
	if Info.Log != nil {
		Info.Log(m, m.FormatPrefix(), level, logs.Format(str, append(arg, _source)...)) // 日志回调
	}
	if m.Option("log.disable") == TRUE {
		return m
	}

	// 日志颜色
	prefix, suffix := "", ""
	if Info.Colors {
		switch level {
		case LOG_CMDS:
			prefix, suffix = "\033[32m", "\033[0m"
		case LOG_AUTH, LOG_COST:
			prefix, suffix = "\033[33m", "\033[0m"
		case LOG_WARN:
			prefix, suffix = "\033[31m", "\033[0m"
		}
	}

	// 长度截断
	switch level {
	case LOG_INFO:
		if len(str) > 4096 {
			str = str[:4096]
		}
	}

	// 输出日志
	logs.Infof(str, append(arg, logs.PrefixMeta(kit.Format("%02d %4s->%-4s %s%s ", m.code, m.source.Name, m.target.Name, prefix, level)), logs.SuffixMeta(suffix), _source)...)
	return m
}
func (m *Message) Log(level string, str string, arg ...Any) *Message {
	return m.log(level, str, arg...)
}
func (m *Message) Logs(level string, arg ...Any) *Message {
	str, meta := m.join(arg...)
	return m.log(level, str, meta...)
}

func (m *Message) Auth(arg ...Any) *Message {
	str, meta := m.join(arg...)
	return m.log(LOG_AUTH, str, meta...)
}
func (m *Message) Cost(arg ...Any) *Message {
	str, meta := m.join(arg...)
	if len(arg) == 0 {
		str = kit.Join(m.meta[MSG_DETAIL], SP)
	}
	list := []string{m.FormatCost(), str}
	return m.log(LOG_COST, kit.Join(list, SP), meta...)
}
func (m *Message) Info(str string, arg ...Any) *Message {
	return m.log(LOG_INFO, str, arg...)
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
	str, meta := m.join(arg...)
	m.log(LOG_WARN, str, meta...)
	if !m.IsErr() {
		if m.error(arg...); len(arg) > 0 {
			switch kit.Format(arg[0]) {
			case ErrNotLogin:
				m.RenderStatusUnauthorized(str)
			case ErrNotRight:
				m.RenderStatusForbidden(str)
			}
		}
	}
	return true
}
func (m *Message) Debug(str string, arg ...Any) {
	if str == "" {
		str = m.FormatMeta()
	}
	m.log(LOG_DEBUG, str, arg...)
}
func (m *Message) Error(err bool, arg ...Any) bool {
	if err {
		m.error(arg...)
		m.log(LOG_ERROR, m.FormatStack(1, 100))
		str, meta := m.join(arg...)
		m.log(LOG_ERROR, str, meta)
		m.log(LOG_ERROR, m.FormatChain())
		return true
	}
	return false
}
func (m *Message) ErrorNotImplement(arg ...Any) {
	m.Error(true, append(kit.List(ErrNotImplement), arg...)...)
}
func (m *Message) error(arg ...Any) {
	if len(arg) == 0 {
		arg = append(arg, "", "")
	} else if len(arg) == 1 {
		arg = append(arg, "")
	}
	str, meta := m.join(arg[2:]...)
	m.meta[MSG_RESULT] = kit.Simple(ErrWarn, arg[0], arg[1], str, meta)
}

func (m *Message) IsErrNotFound() bool { return m.IsErr(ErrNotFound) }
func (m *Message) IsErr(arg ...string) bool {
	return len(arg) > 0 && m.Result(1) == arg[0] || len(arg) == 0 && m.Result(0) == ErrWarn
}

func (m *Message) FormatPrefix() string {
	return kit.Format("%s %d %s->%s", logs.FmtTime(logs.Now()), m.code, m.source.Name, m.target.Name)
}
func (m *Message) FormatShip() string {
	return kit.Format("%s->%s", m.source.Name, m.target.Name)
}
func (m *Message) FormatCost() string {
	return kit.FmtDuration(time.Since(m.time))
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
func (m *Message) FormatStack(s, n int) string {
	pc := make([]uintptr, n+10)
	frames := runtime.CallersFrames(pc[:runtime.Callers(s+1, pc)])

	list := []string{}
	for {
		frame, more := frames.Next()
		file := kit.Slice(kit.Split(frame.File, PS, PS), -1)[0]
		name := kit.Slice(kit.Split(frame.Function, PS, PS), -1)[0]

		switch ls := kit.Split(name, PT, PT); kit.Select("", ls, 0) {
		// case "reflect", "runtime", "http", "task", "icebergs":
		case "reflect", "runtime", "http":
		default:
			list = append(list, kit.Format("%s:%d\t%s", file, frame.Line, name))
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
