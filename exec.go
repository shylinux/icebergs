package ice

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/task"
)

func (m *Message) TryCatch(msg *Message, catch bool, cb ...func(msg *Message)) {
	defer func() {
		switch e := recover(); e {
		case io.EOF, nil:
		default:
			fileline := m.FormatStack(2, 1)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Log("chain", msg.FormatChain())
			m.Log(LOG_WARN, "catch: %s %s", e, kit.FileLine(4, 10)).Log("stack", m.FormatStack(2, 100))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Result(ErrWarn, e, SP, m.FormatStack(2, 5))
			if len(cb) > 1 {
				m.TryCatch(msg, catch, cb[1:]...)
			} else if !catch {
				m.Assert(e)
			}
		}
	}()
	kit.If(len(cb) > 0, func() { cb[0](msg) })
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
func (m *Message) GoSleep(t string, arg ...Any)   { m.Go(func() { m.Sleep(t).Cmd(arg...) }) }
func (m *Message) Go(cb func(), arg ...Any) {
	kit.If(len(arg) == 0, func() { arg = append(arg, logs.FileLine(cb)) })
	task.Put(arg[0], func(task *task.Task) { m.TryCatch(m, true, func(m *Message) { cb() }) })
}
func (m *Message) Wait(cb ...Handler) (wait func(), done Handler) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	t := time.AfterFunc(kit.Duration("30s"), func() { wg.Done() })
	return func() { wg.Wait() }, func(msg *Message, arg ...string) {
		defer wg.Done()
		defer t.Stop()
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
	res := strings.TrimSpace(kit.Select("", m._command(arg...).meta[MSG_RESULT], 0))
	return kit.Select("", res, res != ErrWarn)
}
func (m *Message) Cmdy(arg ...Any) *Message { return m.Copy(m._command(arg...)) }
func (m *Message) CmdHand(cmd *Command, key string, arg ...string) *Message {
	if m._cmd, m._key = cmd, key; cmd == nil {
		return m
	}
	if m._target = cmd.FileLines(); key == SELECT {
		m.Log(LOG_CMDS, "%s.%s %d %v %v", m.Target().Name, key, len(arg), arg, m.Optionv(MSG_FIELDS), logs.FileLineMeta(m._fileline()))
	} else {
		m.Log(LOG_CMDS, "%s.%s %d %v", m.Target().Name, key, len(arg), arg, logs.FileLineMeta(m._fileline()))
	}
	if cmd.Hand != nil {
		cmd.Hand(m, arg...)
	} else if cmd.Actions != nil && cmd.Actions[SELECT] != nil {
		cmd.Actions[SELECT].Hand(m, arg...)
	}
	return m
}
func (m *Message) ActionHand(cmd *Command, key, sub string, arg ...string) *Message {
	if action, ok := cmd.Actions[sub]; !m.Warn(!ok, ErrNotFound, sub, cmd.FileLines()) {
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
	list := kit.Simple(args...)
	kit.If(len(list) == 0, func() { list = m.meta[MSG_DETAIL] })
	if len(list) == 0 {
		return m
	}
	ok := false
	run := func(msg *Message, ctx *Context, cmd *Command, key string, arg ...string) {
		ok = true
		msg._source = _source
		key = kit.Slice(strings.Split(key, PT), -1)[0]
		kit.If(cbs, func() { msg.OptionCB(key, cbs) })
		kit.For(opts, func(k string, v Any) { msg.Option(k, v) })
		m = ctx._command(msg, cmd, key, arg...)
	}
	if list[0] == "" {
		run(m.Spawn(), m.target, m._cmd, m._key, list[1:]...)
	} else if cmd, ok := m.target.Commands[strings.TrimPrefix(list[0], m.target.Prefix()+PT)]; ok {
		run(m.Spawn(), m.target, cmd, list[0], list[1:]...)
	} else if cmd, ok := m.source.Commands[strings.TrimPrefix(list[0], m.source.Prefix()+PT)]; ok {
		run(m.Spawn(m.source), m.source, cmd, list[0], list[1:]...)
	} else {
		m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) { run(m.Spawn(s), s, cmd, key, list[1:]...) })
	}
	m.Warn(!ok, ErrNotFound, kit.Format(list))
	return m
}
func (c *Context) _command(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._cmd, m._key, m._sub = cmd, key, SELECT; cmd == nil {
		return m
	}
	if m.meta[MSG_DETAIL] = kit.Simple(m.PrefixKey(), arg); cmd.Actions != nil {
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
			if m.Warn(m.OptionDefault(name, kit.Format(kit.Value(v, VALUE))) == "" && kit.Value(v, "need") == "must", ErrNotValid, name) {
				return m
			}
		}
	}
	m._target = kit.Select(logs.FileLine(h.Hand), cmd.FileLines(), cmd.RawHand != nil)
	m.Log(LOG_CMDS, "%s.%s %s %d %v", c.Name, key, sub, len(arg), arg, logs.FileLineMeta(m._fileline()))
	h.Hand(m, arg...)
	return m
}
func (c *Command) FileLines() string {
	return kit.Join(kit.Slice(kit.Split(c.FileLine(), PS), -3), PS)
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
