package ice

import (
	"errors"
	"io"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

func (m *Message) TryCatch(msg *Message, silent bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			fileline := m.FormatStack(2, 1)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log("chain", msg.FormatChain())
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log("stack", msg.FormatStack(2, 100))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Echo("%v", e)
			if len(hand) > 1 {
				m.TryCatch(msg, silent, hand[1:]...)
			} else if !silent {
				m.Assert(e) // 抛出异常
			}
		}
	}()

	if len(hand) > 0 {
		hand[0](msg)
	}
	return m
}
func (m *Message) Assert(expr interface{}) bool {
	switch expr := expr.(type) {
	case nil:
		return true
	case error:
	case bool:
		if expr == true {
			return true
		}
	default:
		expr = errors.New(kit.Format("error: %v", expr))
	}
	m.Result(ErrPanic, expr)
	panic(expr)
}
func (m *Message) Sleep(d string, arg ...interface{}) *Message {
	m.Debug("sleep %s %s", d, kit.FileLine(2, 3))
	if time.Sleep(kit.Duration(d)); len(arg) > 0 {
		m.Cmdy(arg...)
	}
	return m
}
func (m *Message) Sleep300ms(arg ...interface{}) *Message { return m.Sleep("300ms", arg...) }
func (m *Message) Sleep30ms(arg ...interface{}) *Message  { return m.Sleep("30ms", arg...) }
func (m *Message) Sleep3s(arg ...interface{}) *Message    { return m.Sleep("3s", arg...) }
func (m *Message) Sleep30s(arg ...interface{}) *Message   { return m.Sleep("30s", arg...) }
func (m *Message) Hold(n int) *Message {
	for ctx := m.target; ctx != nil; ctx = ctx.context {
		if ctx.wg != nil {
			ctx.wg.Add(n)
			break
		}
	}
	return m
}
func (m *Message) Done(ok bool) bool {
	if !ok {
		return false
	}
	defer func() { recover() }()

	for ctx := m.target; ctx != nil; ctx = ctx.context {
		if ctx.wg != nil {
			ctx.wg.Done()
			break
		}
	}
	return ok
}
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	wait := make(chan bool, 2)

	p := kit.Select("10s", m.Option("timeout"))
	t := time.AfterFunc(kit.Duration(p), func() {
		m.Warn(true, ErrTimeout, m.Detailv())
		m.Back(nil)
		wait <- false
	})

	m.cb = func(res *Message) *Message {
		if res = cb(res); sync {
			wait <- true
			t.Stop()
		}
		return res
	}

	if sync {
		<-wait
	} else {
		t.Stop()
	}
	return m
}
func (m *Message) Back(res *Message) *Message {
	if m.cb != nil {
		if sub := m.cb(res); m.message != nil {
			m.message.Back(sub)
		}
	}
	return m
}
func (m *Message) Go(cb interface{}) *Message {
	task.Put(kit.FileLine(cb, 3), func(task *task.Task) error {
		m.TryCatch(m, true, func(m *Message) {
			switch cb := cb.(type) {
			case func(*Message):
				cb(m.Spawn())
			case func():
				cb()
			}
		})
		return nil
	})
	return m
}

func (m *Message) Watch(key string, arg ...string) *Message {
	if len(arg) == 0 {
		arg = append(arg, m.Prefix(AUTO))
	}
	m.Cmd("event", ACTION, "listen", "event", key, CMD, kit.Join(arg, SP))
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd("event", ACTION, "action", "event", key, arg)
	return m
}
func (m *Message) Right(arg ...interface{}) bool {
	key := strings.ReplaceAll(kit.Keys(arg...), PS, PT)
	return m.Option(MSG_USERROLE) == "root" || !m.Warn(m.Cmdx("role", "right", m.Option(MSG_USERROLE), key) != OK,
		ErrNotRight, kit.Join(kit.Simple(arg), PT), "userrole", m.Option(MSG_USERROLE), "fileline", kit.FileLine(2, 3))
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || arg == "" || kit.Format(arg) == m.Conf("runtime", "node.name") {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}
func (m *Message) PodCmd(arg ...interface{}) bool {
	if pod := m.Option(POD); pod != "" {
		if m.Option(POD, ""); m.Option(MSG_UPLOAD) != "" {
			msg := m.Cmd(CACHE, "upload")
			m.Option(MSG_UPLOAD, msg.Append(HASH), msg.Append(NAME), msg.Append("size"))
		}
		m.Cmdy(append(kit.List(SPACE, pod), arg...))
		return true
	}
	return false
}
