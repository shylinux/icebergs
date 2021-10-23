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
			fileline := kit.FileLine(4, 5)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log(kit.MDB_CHAIN, msg.FormatChain())
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log(kit.MDB_STACK, msg.FormatStack())
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			if len(hand) > 1 { // 捕获异常
				m.TryCatch(msg, silent, hand[1:]...)
			} else if !silent { // 抛出异常
				m.Assert(e)
			}
		}
	}()

	if len(hand) > 0 { // 运行函数
		hand[0](msg)
	}
	return m
}
func (m *Message) Assert(expr interface{}) bool {
	switch expr := expr.(type) {
	case nil:
		return true
	case error:
		panic(expr)
	case bool:
		if expr == true {
			return true
		}
	}
	panic(errors.New(kit.Format("error: %v", expr)))
}
func (m *Message) Sleep(d string) *Message {
	m.Debug("sleep %s %s", d, kit.FileLine(2, 3))
	time.Sleep(kit.Duration(d))
	return m
}
func (m *Message) Hold(n int) *Message {
	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	ctx.wg.Add(n)
	return m
}
func (m *Message) Done(b bool) bool {
	if !b {
		return false
	}
	defer func() { recover() }()

	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	ctx.wg.Done()
	return true
}
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	wait := make(chan bool, 2)

	p := kit.Select("10s", m.Option(kit.MDB_TIMEOUT))
	t := time.AfterFunc(kit.Duration(p), func() {
		m.Warn(true, "%s timeout %v", p, m.Detailv())
		m.Back(nil)
		wait <- false
	})

	m.cb = func(sub *Message) *Message {
		if sync {
			wait <- true
			t.Stop()
		}
		return cb(sub)
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
	m.Cmd("event", "action", "listen", "event", key, CMD, strings.Join(arg, SP))
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd("event", "action", "action", "event", key, arg)
	return m
}
func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == "root" || !m.Warn(m.Cmdx("role", "right",
		m.Option(MSG_USERROLE), strings.ReplaceAll(kit.Keys(arg...), "/", PT)) != OK,
		ErrNotRight, m.Option(MSG_USERROLE), OF, strings.Join(kit.Simple(arg), PT), " at ", kit.FileLine(2, 3))
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || arg == "" || kit.Format(arg) == m.Conf("runtime", "node.name") {
		return nil
	}
	return []string{"space", kit.Format(arg)}
}
func (m *Message) PodCmd(arg ...interface{}) bool {
	if pod := m.Option(POD); pod != "" {
		if m.Option(POD, ""); m.Option(MSG_UPLOAD) != "" {
			msg := m.Cmd("cache", "upload")
			m.Option(MSG_UPLOAD, msg.Append(kit.MDB_HASH), msg.Append(kit.MDB_NAME), msg.Append(kit.MDB_SIZE))
		}
		m.Cmdy(append([]interface{}{"space", pod}, arg...))
		return true
	}
	return false
}
