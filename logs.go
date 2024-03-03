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
	"shylinux.com/x/toolkits/task"
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
			if len(v) > 0 {
				list = append(list, kit.JoinKV(DF+SP, SP, v...))
			}
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
	args, traceid := []Any{}, ""
	for _, v := range arg {
		if v, ok := v.(logs.Meta); ok && v.Key == logs.TRACEID {
			traceid = kit.Select(strings.TrimSpace(v.Value), traceid)
			continue
		}
		args = append(args, v)
	}
	_source := logs.FileLineMeta(3)
	kit.If(Info.Log != nil, func() { Info.Log(m, m.FormatPrefix(traceid), level, logs.Format(str, append(args, _source)...)) })
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
	logs.Infof(str, append(args, logs.PrefixMeta(kit.Format("%s %s%s ", m.FormatShip(traceid), prefix, level)), logs.SuffixMeta(suffix), _source)...)
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
	kit.If(str == "" || len(arg) == 0, func() { str, meta = kit.Join(m.value(MSG_DETAIL), SP), []Any{logs.FileLineMeta(m._fileline())} })
	return m.log(LOG_COST, kit.Join([]string{m.FormatCost(), str}, SP), meta...)
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
	case string:
		if err != "" {
			return false
		}
	case bool:
		if !err {
			return false
		}
	case nil:
		return false
	}
	str, meta := m.join(arg...)
	if m.log(LOG_WARN, str, meta...); len(arg) > 0 {
		m.Option(MSG_TITLE, kit.JoinWord(kit.Keys(m.Option(MSG_USERPOD), m.CommandKey(), m.ActionKey()), logs.FileLines(-2)))
		m.error(arg...)
		kit.If(map[string]int{
			ErrNotLogin: http.StatusUnauthorized,
			ErrNotRight: http.StatusForbidden,
			ErrNotAllow: http.StatusMethodNotAllowed,
			ErrNotFound: http.StatusNotFound,
			ErrNotValid: http.StatusBadRequest,
		}[kit.Format(arg[0])], func(s int) { m.Render(RENDER_STATUS, s, str) })
	}
	return true
}
func (m *Message) WarnNotLogin(err Any, arg ...Any) bool {
	return m.Warn(err, ErrNotLogin, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) WarnNotRight(err Any, arg ...Any) bool {
	return m.Warn(err, ErrNotRight, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) WarnNotAllow(err Any, arg ...Any) bool {
	return m.Warn(err, ErrNotAllow, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) WarnNotFound(err Any, arg ...Any) bool {
	return m.Warn(err, ErrNotFound, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) WarnNotValid(err Any, arg ...Any) bool {
	return m.Warn(err, ErrNotValid, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) WarnNotValidTime(time Any, arg ...Any) bool {
	return m.Warn(kit.Format(time) < m.Time(), ErrNotValid, kit.Simple(arg...), time, m.Time(), logs.FileLineMeta(2))
}
func (m *Message) WarnAlreadyExists(err Any, arg ...Any) bool {
	return m.Warn(err, ErrAlreadyExists, kit.Simple(arg...), logs.FileLineMeta(2))
}
func (m *Message) ErrorNotImplement(arg ...Any) *Message {
	m.Error(true, append(kit.List(ErrNotImplement), append(arg, logs.FileLineMeta(2)))...)
	return m
}
func (m *Message) Error(err bool, arg ...Any) bool {
	if m.Warn(err, arg...) {
		str, _ := m.join(arg...)
		m.error(arg...)
		panic(str)
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
	return len(arg) == 0 && m.index(MSG_RESULT, 0) == ErrWarn || len(arg) > 0 && m.index(MSG_RESULT, 1) == arg[0]
}
func (m *Message) IsErrNotFound() bool { return m.IsErr(ErrNotFound) }
func (m *Message) Debug(str string, arg ...Any) {
	if m.Option(MSG_DEBUG) == TRUE {
		kit.Format(str == "", func() { str = m.FormatMeta() })
		m.log(LOG_DEBUG, str, arg...)
	}
}

func (m *Message) FormatTaskMeta() task.Meta {
	return task.Meta{Prefix: m.FormatShip() + " ", FileLine: kit.FileLine(2, 3)}
}
func (m *Message) FormatPrefix(traceid ...string) string {
	return kit.Format("%s %s", logs.FmtTime(logs.Now()), m.FormatShip(traceid...))
}
func (m *Message) FormatShip(traceid ...string) string {
	_traceid := ""
	kit.If(kit.Select(m.Option(LOG_TRACEID), traceid, 0), func(traceid string) {
		// _traceid = kit.Format("%s: %s ", logs.TRACEID, traceid+"-"+m.Option("task.id")+"-"+m.Option("work.id"))
		_traceid = kit.Format("%s: %s ", logs.TRACEID, traceid)
	})
	return kit.Format("%s%02d %4s->%-4s", _traceid, m.code, m.source.Name, m.target.Name)
}
func (m *Message) FormatSize() string {
	n := len(m.value(MSG_APPEND))
	kit.If(m.FieldsIsDetail(), func() { n = len(m.value(KEY)) })
	return kit.Format("%dx%d %v %v", m.Length(), n, kit.Simple(m.value(MSG_APPEND)), kit.FmtSize(len(m.Result())))
}
func (m *Message) FormatCost() string { return kit.FmtDuration(time.Since(m.time)) }
func (m *Message) FormatMeta() string {
	defer m.lock.RLock()()
	return kit.Format(m._meta)
}
func (m *Message) FormatsMeta(w io.Writer, arg ...string) (res string) {
	if w == nil {
		buf := bytes.NewBuffer(make([]byte, 0, MOD_BUFS))
		defer func() { res = buf.String() }()
		w = buf
	}
	kit.For(m.value(MSG_OPTION), func(i int, k string) {
		ls := m.value(k)
		kit.If(len(ls) == 0 || len(ls) == 1 && ls[0] == "", func() { m.index(MSG_OPTION, i, "") })
	})
	kit.For(m.Optionv(""), func(key string) { kit.If(strings.HasPrefix(key, "sessid_"), func() { arg = append(arg, key) }) })
	m.value(MSG_OPTION, kit.Filters(m.value(MSG_OPTION), kit.Simple(MSG_CMDS, MSG_SESSID, MSG_OPTS, "", MSG_COUNT, MSG_CHECKER, arg)...)...)
	bio, count := bufio.NewWriter(w), 0
	defer bio.Flush()
	SP, NL := SP, NL
	kit.If(m.Option(MSG_DEBUG) != TRUE, func() { SP, NL = "", "" })
	echo := func(arg ...Any) { fmt.Fprint(bio, arg...) }
	push := func(k string) {
		if len(m.value(k)) == 0 {
			return
		}
		kit.If(count > 0, func() { echo(FS, NL) })
		echo(kit.Format("%s%q:%s", SP, k, SP))
		b, _ := json.Marshal(m.value(k))
		bio.Write(b)
		count++
	}
	echo("{", NL)
	defer echo(NL, "}", NL)
	kit.For(kit.Simple(MSG_DETAIL, MSG_OPTION, m.value(MSG_OPTION), m.value(MSG_APPEND), MSG_APPEND, MSG_RESULT), push)
	return
}
func (m *Message) FormatChain() string {
	ms := []*Message{}
	for msg := m; msg != nil; msg = msg.message {
		ms = append(ms, msg)
	}
	show := func(msg *Message, key string, arg ...string) string {
		ls := msg.value(key)
		if len(ls) == 0 || len(ls) == 1 && ls[0] == "" {
			return ""
		}
		return kit.Format("%s%s:%s%d %v", kit.Select("", arg, 0), key, kit.Select("", arg, 1), len(ls), ls)
	}
	meta := []string{}
	for i := len(ms) - 1; i >= 0; i-- {
		msg := ms[i]
		meta = append(meta, kit.Join([]string{msg.FormatPrefix(), show(msg, MSG_DETAIL), show(msg, MSG_OPTION), show(msg, MSG_APPEND), show(msg, MSG_RESULT), msg._cmd.FileLines()}, SP))
		kit.For(msg.value(MSG_OPTION), func(k string) { kit.If(show(msg, k, TB, SP), func(s string) { meta = append(meta, s) }) })
		kit.For(msg.value(MSG_APPEND), func(k string) { kit.If(show(msg, k, TB, SP), func(s string) { meta = append(meta, s) }) })
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
			if kit.HasPrefix(name,
				"icebergs.(*Context)._action",
				"icebergs.(*Context)._command",
				"icebergs.(*Message)._command",
				"icebergs.(*Message).Cmd",
				"icebergs.(*Message).CmdHand",
				"icebergs.(*Message).Search",
				"icebergs.(*Message).TryCatch",
				"icebergs.(*Message).Go",
			) {
				break
			}
			list = append(list, kit.Format("%s:%d\t%s", file, frame.Line, name))
		}
		if len(list) >= n || !more {
			break
		}
	}
	return kit.Join(list, NL)
}
