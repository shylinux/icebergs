package ice

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
	"unicode"

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
			list = append(list, kit.JoinKV(DF+SP, SP, v...))
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
		case []string:
			arg[i+1] = kit.Join(v, SP)
		}
		list = append(list, key+kit.Select("", DF, !strings.Contains(key, DF)), kit.Format(arg[i+1]))
	}
	return kit.Join(list, SP), meta
}
func (m *Message) log(level string, str string, arg ...Any) *Message {
	if m.Option(LOG_DISABLE) == TRUE {
		return m
	}
	_source := logs.FileLineMeta(3)
	kit.If(Info.Log != nil, func() { Info.Log(m, m.FormatPrefix(), level, logs.Format(str, append(arg, _source)...)) })
	prefix, suffix := "", ""
	if Info.Colors {
		switch level {
		case LOG_CMDS:
			prefix, suffix = "\033[32m", "\033[0m"
		case LOG_AUTH, LOG_COST:
			prefix, suffix = "\033[33m", "\033[0m"
		case LOG_WARN, LOG_ERROR:
			prefix, suffix = "\033[31m", "\033[0m"
		}
	}
	kit.If(level == LOG_INFO && len(str) > 4096, func() { str = str[:4096] })
	logs.Infof(str, append(arg, logs.PrefixMeta(kit.Format("%02d %4s->%-4s %s%s ", m.code, m.source.Name, m.target.Name, prefix, level)), logs.SuffixMeta(suffix), _source)...)
	return m
}
func (m *Message) Log(level string, str string, arg ...Any) *Message {
	return m.log(level, str, arg...)
}
func (m *Message) Logs(level string, arg ...Any) *Message {
	str, meta := m.join(arg...)
	kit.If(len(level) > 0 && unicode.IsUpper([]rune(level)[0]), func() { meta = []Any{logs.FileLineMeta("")} })
	return m.log(level, str, meta...)
}

func (m *Message) Auth(arg ...Any) *Message {
	str, meta := m.join(arg...)
	return m.log(LOG_AUTH, str, meta...)
}
func (m *Message) Cost(arg ...Any) *Message {
	str, meta := m.join(arg...)
	kit.If(str == "" || len(arg) == 0, func() { str, meta = kit.Join(m.meta[MSG_DETAIL], SP), []Any{logs.FileLineMeta(m._fileline())} })
	return m.log(LOG_COST, kit.Join([]string{m.FormatCost(), str}, SP), meta...)
}
func (m *Message) Info(str string, arg ...Any) *Message {
	return m.log(LOG_INFO, str, arg...)
}
func (m *Message) Warn(err Any, arg ...Any) bool {
	switch err := err.(type) {
	case error:
		if err == io.EOF {
			return true
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
	if m.log(LOG_WARN, str, meta...); !m.IsErr() && len(arg) > 0 {
		m.error(arg...)
		kit.If(map[string]int{
			ErrNotLogin: http.StatusUnauthorized,
			ErrNotRight: http.StatusForbidden,
			ErrNotFound: http.StatusNotFound,
			ErrNotValid: http.StatusBadRequest,
		}[kit.Format(arg[0])], func(s int) { m.Render(RENDER_STATUS, s, str) })
	}
	return true
}
func (m *Message) WarnTimeNotValid(time Any, arg ...Any) bool {
	return m.Warn(kit.Format(time) < m.Time(), ErrNotValid, kit.Simple(arg), time, m.Time(), logs.FileLineMeta(2))
}
func (m *Message) ErrorNotImplement(arg ...Any) *Message {
	m.Error(true, append(kit.List(ErrNotImplement), arg...)...)
	return m
}
func (m *Message) Error(err bool, arg ...Any) bool {
	if err {
		str, meta := m.join(arg...)
		m.log(LOG_ERROR, m.FormatChain()).log(LOG_ERROR, str, meta).log(LOG_ERROR, m.FormatStack(2, 100)).error(arg...)
		return true
	}
	return false
}
func (m *Message) error(arg ...Any) {
	if len(arg) > 2 {
		str, _ := m.join(arg[2:]...)
		m.Resultv(ErrWarn, kit.Simple(arg[0], arg[1], SP+str))
	} else {
		m.Resultv(ErrWarn, kit.Simple(arg))
	}
}
func (m *Message) IsOk() bool { return m.Result() == OK }
func (m *Message) IsErr(arg ...string) bool {
	return len(arg) == 0 && kit.Select("", m.meta[MSG_RESULT], 0) == ErrWarn || len(arg) > 0 && kit.Select("", m.meta[MSG_RESULT], 1) == arg[0]
}
func (m *Message) IsErrNotFound() bool {
	return m.IsErr(ErrNotFound)
}
func (m *Message) Debug(str string, arg ...Any) {
	if m.Option("debug") == TRUE {
		kit.Format(str == "", func() { str = m.FormatMeta() })
		m.log(LOG_DEBUG, str, arg...)
	}
}

func (m *Message) FormatPrefix() string {
	return kit.Format("%s %d %s->%s", logs.FmtTime(logs.Now()), m.code, m.source.Name, m.target.Name)
}
func (m *Message) FormatSize() string {
	return kit.Format("%dx%d %v", m.Length(), len(m.meta[MSG_APPEND]), kit.Simple(m.meta[MSG_APPEND]))
}
func (m *Message) FormatCost() string { return kit.FmtDuration(time.Since(m.time)) }
func (m *Message) FormatMeta() string { return kit.Format(m.meta) }
func (m *Message) FormatsMeta(w io.Writer, arg ...string) (res string) {
	if w == nil {
		buf := bytes.NewBuffer(make([]byte, 0, MOD_BUFS))
		defer func() { res = buf.String() }()
		w = buf
	}
	kit.For(m.meta[MSG_OPTION], func(i int, k string) {
		kit.If(len(m.meta[k]) == 0 || len(m.meta[k]) == 1 && m.meta[k][0] == "", func() { m.meta[MSG_OPTION][i] = "" })
	})
	m.meta[MSG_OPTION] = kit.Filters(m.meta[MSG_OPTION], MSG_CMDS, MSG_FIELDS, MSG_SESSID, MSG_OPTS, MSG_OUTPUT, MSG_INDEX, "", "aaa.checker")
	kit.If(len(arg) == 0 && m.Option(DEBUG) == TRUE, func() { arg = []string{SP, SP, NL} })
	bio, count, NL := bufio.NewWriter(w), 0, kit.Select("", arg, 2)
	defer bio.Flush()
	echo := func(arg ...Any) { fmt.Fprint(bio, arg...) }
	push := func(k string) {
		if len(m.meta[k]) == 0 {
			return
		}
		kit.If(count > 0, func() { echo(FS, NL) })
		echo(kit.Format("%s%q:%s", kit.Select("", arg, 0), k, kit.Select("", arg, 1)))
		b, _ := json.Marshal(m.meta[k])
		bio.Write(b)
		count++
	}
	echo("{", NL)
	defer echo("}", NL)
	kit.For(kit.Simple(MSG_DETAIL, MSG_OPTION, m.meta[MSG_OPTION], m.meta[MSG_APPEND], MSG_APPEND, MSG_RESULT), push)
	return
}
func (m *Message) FormatChain() string {
	ms := []*Message{}
	for msg := m; msg != nil; msg = msg.message {
		ms = append(ms, msg)
	}
	show := func(msg *Message, key string, arg ...string) string {
		if len(msg.meta[key]) == 0 || len(msg.meta[key]) == 1 && msg.meta[key][0] == "" {
			return ""
		}
		return kit.Format("%s%s:%s%d %v", kit.Select("", arg, 0), key, kit.Select("", arg, 1), len(msg.meta[key]), msg.meta[key])
	}
	meta := []string{}
	for i := len(ms) - 1; i >= 0; i-- {
		msg := ms[i]
		meta = append(meta, kit.Join([]string{msg.FormatPrefix(), show(msg, MSG_DETAIL), show(msg, MSG_OPTION), show(msg, MSG_APPEND), show(msg, MSG_RESULT), msg._cmd.FileLines()}, SP))
		kit.For(msg.meta[MSG_OPTION], func(k string) { kit.If(show(msg, k, TB, SP), func(s string) { meta = append(meta, s) }) })
		kit.For(msg.meta[MSG_APPEND], func(k string) { kit.If(show(msg, k, TB, SP), func(s string) { meta = append(meta, s) }) })
	}
	return kit.Join(meta, NL)
}
func (m *Message) FormatStack(s, n int) string {
	list, pc := []string{}, make([]uintptr, n+10)
	frames := runtime.CallersFrames(pc[:runtime.Callers(s+1, pc)])
	for {
		frame, more := frames.Next()
		name := kit.Slice(kit.Split(frame.Function, PS, PS), -1)[0]
		file := kit.Join(kit.Slice(kit.Split(frame.File, PS, PS), -2), PS)
		switch ls := kit.Split(name, PT, PT); kit.Select("", ls, 0) {
		case "reflect", "runtime", "http":
		default:
			list = append(list, kit.Format("%s:%d\t%s", file, frame.Line, name))
		}
		if len(list) >= n || !more {
			break
		}
	}
	return kit.Join(list, NL)
}
