package ice

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"time"

	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/task"
)

func (m *Message) TryCatch(catch bool, cb ...func(*Message)) {
	defer func() {
		switch e := recover(); e {
		case io.EOF, nil:
		default:
			fileline := m.FormatStack(2, 1)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Log("chain", "\n"+m.FormatChain())
			m.Log(LOG_WARN, "catch: %s %s", e, kit.FileLine(4, 10)).Log("stack", "\n"+m.FormatStack(2, 1000))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Result(ErrWarn, e, SP, m.FormatStack(2, 5))
			if len(cb) > 1 {
				m.TryCatch(catch, cb[1:]...)
			} else if !catch {
				m.Assert(e)
			}
		}
	}()
	kit.If(len(cb) > 0, func() { cb[0](m) })
}
func (m *Message) Assert(expr Any) bool {
	switch e := expr.(type) {
	case nil:
		return true
	case string:
		if expr != "" {
			return true
		}
	case bool:
		if e == true {
			return true
		}
	case error:
	default:
		expr = errors.New(kit.Format("error: %v", e))
	}
	m.Result(ErrWarn, expr, logs.FileLine(2))
	panic(expr)
}
func (m *Message) Sleep(d Any, arg ...Any) *Message {
	defer kit.If(len(arg) > 0, func() { m.Cmdy(arg...) })
	time.Sleep(kit.Duration(d))
	return m
}
func (m *Message) Sleep300ms(arg ...Any) *Message { return m.Sleep("300ms", arg...) }
func (m *Message) Sleep30ms(arg ...Any) *Message  { return m.Sleep("30ms", arg...) }
func (m *Message) Sleep3s(arg ...Any) *Message    { return m.Sleep("3s", arg...) }
func (m *Message) GoSleep(t string, cb func(), arg ...Any) *Message {
	return m.Go(func() { m.Spawn(kit.Dict(MSG_COUNT, "0")).Sleep(t); cb() })
}
func (m *Message) GoSleep300ms(cb func(), arg ...Any) { m.GoSleep("300ms", cb, arg...) }
func (m *Message) GoSleep30ms(cb func(), arg ...Any)  { m.GoSleep("30ms", cb, arg...) }
func (m *Message) GoSleep3s(cb func(), arg ...Any)    { m.GoSleep("3s", cb, arg...) }
func (m *Message) Go(cb func(), arg ...Any) *Message {
	meta := m.FormatTaskMeta()
	meta.FileLine = kit.FileLine(2, 3)
	kit.If(len(arg) > 0, func() { meta.FileLine = kit.Format(arg[0]) })
	task.Put(meta, nil, func(task *task.Task) {
		m.TryCatch(true, func(m *Message) {
			m.Option("work.id", kit.Format(task.WorkId()))
			m.Option("task.id", kit.Format(task.TaskId()))
			cb()
		})
	})
	return m
}
func (m *Message) GoWait(cb func(func()), arg ...Any) *Message {
	res := make(chan bool, 2)
	defer func() { <-res }()
	return m.Go(func() { cb(func() { res <- true }) }, arg...)
}
func (m *Message) Wait(d string, cb ...Handler) (wait func() bool, done Handler) {
	if d == "" {
		return nil, nil
	}
	sync := make(chan bool, 2)
	t := time.AfterFunc(kit.Duration(d), func() { sync <- false })
	return func() bool { return <-sync }, func(msg *Message, arg ...string) {
		defer func() { t.Stop(); sync <- true }()
		kit.If(len(cb) > 0 && cb[0] != nil, func() { cb[0](msg, arg...) }, func() { m.Copy(msg) })
	}
}

func (m *Message) Cmd(arg ...Any) *Message  { return m._command(arg...) }
func (m *Message) Cmds(arg ...Any) *Message { return m.Cmd(append(arg, OptionFields(""))...) }
func (m *Message) Cmdv(arg ...Any) string {
	args := kit.Simple(arg...)
	field := kit.Slice(args, -1)[0]
	return m._command(kit.Slice(args, 0, -1), OptionFields(field)).Append(field)
}
func (m *Message) Cmdx(arg ...Any) string {
	res := strings.TrimSpace(m._command(arg...).index(MSG_RESULT, 0))
	return kit.Select("", res, res != strings.TrimSpace(ErrWarn))
}
func (m *Message) Cmdy(arg ...Any) *Message { return m.Copy(m._command(arg...)) }
func (m *Message) CmdList(arg ...string) []string {
	msg, list := m._command(arg), []string{}
	kit.For(msg._cmd.List, func(value Map) {
		kit.If(!kit.IsIn(kit.Format(kit.Value(value, TYPE)), html.BUTTON), func() { list = append(list, kit.Format(kit.Value(value, NAME))) })
	})
	return msg.Appendv(kit.Select(kit.Select("", list, 0), list, len(arg)-1))
}
func (m *Message) CmdHand(cmd *Command, key string, arg ...string) *Message {
	if m._cmd, m._key, m._sub = cmd, key, LIST; cmd == nil {
		return m
	}
	level := LOG_CMDS
	if m._target = cmd.FileLines(); key == SELECT {
		m.Log(level, "%s.%s %v %v", m.Target().Name, key, formatArg(arg...), m.Optionv(MSG_FIELDS), logs.FileLineMeta(m._fileline()))
	} else {
		m.Log(level, "%s.%s %v", m.Target().Name, key, formatArg(arg...), logs.FileLineMeta(m._fileline()))
	}
	if cmd.Hand != nil {
		cmd.Hand(m, arg...)
	} else if cmd.Actions != nil && cmd.Actions[SELECT] != nil {
		m._sub = SELECT
		cmd.Actions[SELECT].Hand(m, arg...)
	}
	return m
}
func (m *Message) ActionHand(cmd *Command, key, sub string, arg ...string) *Message {
	if action, ok := cmd.Actions[sub]; !m.WarnNotFoundIndex(!ok, sub, cmd.FileLines()) {
		return m.Target()._action(m, cmd, key, sub, action, arg...)
	}
	return m
}
func (m *Message) _command(arg ...Any) *Message {
	args, opts, cbs, _source := []Any{}, Map{}, kit.Value(nil), logs.FileLine(3)
	for _, v := range arg {
		switch val := v.(type) {
		case nil:
		case string:
			args = append(args, v)
		case Maps:
			kit.For(val, func(k, v string) { opts[k] = v })
		case Map:
			kit.For(kit.KeyValue(nil, "", val), func(k string, v Any) { opts[k] = v })
		case Option:
			opts[val.Name] = val.Value
		case logs.Meta:
			kit.If(val.Key == "fileline", func() { _source = val.Value })
		case func(int, Maps, []string):
			defer func() { m.Table(val) }()
		case func(Maps):
			defer func() { m.Table(val) }()
		default:
			if reflect.TypeOf(val).Kind() == reflect.Func {
				cbs = val
			} else {
				args = append(args, v)
			}
		}
	}
	if count := kit.Int(m.Option(MSG_COUNT, kit.Format(kit.Int(m.Option(MSG_COUNT))+1))); m.Warn(count > 3000, ErrTooDeepCount) {
		panic(count)
	}
	list := kit.Simple(args...)
	kit.If(len(list) == 0, func() { list = m.value(MSG_DETAIL) })
	if len(list) == 0 {
		return m
	}
	ok := false
	run := func(msg *Message, ctx *Context, cmd *Command, key string, arg ...string) {
		ok, msg._source = true, _source
		key = kit.Slice(strings.Split(key, PT), -1)[0]
		kit.If(cbs, func() { msg.OptionCB(key, cbs) })
		kit.For(opts, func(k string, v Any) { msg.Option(k, v) })
		m = ctx._command(msg, cmd, key, arg...)
	}
	if list[0] == "" {
		run(m.Spawn(), m.target, m._cmd, m._key, list[1:]...)
	} else {
		_target, _key := m.target, m._key
		m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
			m.target, m._key = _target, _key
			run(m.Spawn(s), s, cmd, key, list[1:]...)
		})
	}
	m.WarnNotFoundIndex(!ok, kit.Format(list))
	return m
}
func (c *Context) _command(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._cmd, m._key = cmd, key; cmd == nil {
		return m
	}
	if m.value(MSG_DETAIL, kit.Simple(m.ShortKey(), arg)...); cmd.Actions != nil {
		if len(arg) > 1 && arg[0] == ACTION {
			if h, ok := cmd.Actions[arg[1]]; ok {
				return c._action(m, cmd, key, arg[1], h, arg[2:]...)
			}
		} else if len(arg) > 0 {
			if h, ok := cmd.Actions[arg[0]]; ok {
				return c._action(m, cmd, key, arg[0], h, arg[1:]...)
			}
		}
	}
	if len(arg) > 0 && arg[0] == ACTION && arg[1] == INPUTS {
		return m
	}
	return m.CmdHand(cmd, key, arg...)
}
func (c *Context) _action(m *Message, cmd *Command, key string, sub string, h *Action, arg ...string) *Message {
	if h.Hand == nil {
		return m.Cmdy(kit.Split(kit.Select(sub, h.Name)), arg)
	}
	if m._cmd, m._key, m._sub = cmd, key, sub; len(h.List) > 0 && sub != SEARCH {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, NAME))
			if i == 0 {
				if len(arg) > 0 && arg[0] == name {
					kit.For(arg, func(k, v string) { m.Option(k, v) })
				} else {
					order = true
				}
			}
			kit.If(order && i < len(arg), func() { m.Option(name, arg[i]) })
			if m.WarnNotValid(m.OptionDefault(name, kit.Format(kit.Value(v, VALUE))) == "" && kit.Value(v, "need") == "must", name, key, sub) {
				return m
			}
		}
	}
	m._target = kit.Select(logs.FileLine(h.Hand), cmd.FileLines(), cmd.RawHand != nil)
	level := LOG_CMDS
	kit.If(key == "role" && sub == "right", func() { level = LOG_AUTH })
	m.Log(level, "%s.%s %s %v", c.Name, key, sub, formatArg(arg...), logs.FileLineMeta(m._fileline()))
	h.Hand(m, arg...)
	return m
}
func formatArg(arg ...string) string {
	return kit.Format("%d %v", len(arg), kit.ReplaceAll(kit.Format("%v", arg), "\r\n", "\\r\\n", "\t", "\\t", "\n", "\\n"))
}
func (c *Command) FileLines() string {
	return kit.Join(kit.Slice(kit.Split(c.FileLine(), PS), -2), PS)
}
func (c *Command) FileLine() string {
	if c == nil {
		return ""
	} else if c.RawHand != nil {
		switch h := c.RawHand.(type) {
		case string:
			return h
		default:
			return logs.FileLines(c.RawHand)
		}
	} else if c.Hand != nil {
		return logs.FileLines(c.Hand)
	} else {
		return ""
	}
}
