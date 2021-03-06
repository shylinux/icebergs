package ice

import (
	"errors"
	"fmt"
	"io"
	"time"

	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/task"
)

func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			fileline := kit.FileLine(4, 5)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log("chain", msg.Format("chain"))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			m.Log("stack", msg.Format("stack"))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline)
			if len(hand) > 1 {
				// 捕获异常
				m.TryCatch(msg, safe, hand[1:]...)
			} else if !safe {
				// 抛出异常
				m.Assert(e)
			}
		}
	}()

	if len(hand) > 0 {
		// 运行函数
		hand[0](msg)
	}
	return m
}
func (m *Message) Assert(arg interface{}) bool {
	switch arg := arg.(type) {
	case nil:
		return true
	case bool:
		if arg == true {
			return true
		}
	case error:
		panic(arg)
	}
	panic(errors.New(fmt.Sprintf("error: %v", arg)))
}
func (m *Message) Sleep(arg string) *Message {
	m.Debug("sleep %s %s", arg, kit.FileLine(2, 3))
	time.Sleep(kit.Duration(arg))
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
	p := kit.Select("10s", m.Option("timeout"))
	t := time.AfterFunc(kit.Duration(p), func() {
		m.Warn(true, "%s timeout %v", p, m.Detailv())
		wait <- false
		m.Back(nil)
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
func (m *Message) Gos(msg *Message, cb interface{}, args ...interface{}) *Message {
	task.Put(kit.FileLine(3, 3), func(task *task.Task) error {
		msg.TryCatch(msg, true, func(msg *Message) {
			switch cb := cb.(type) {
			case func(*Message):
				cb(msg)
			case func():
				cb()
			}
		})
		return nil
	})
	return msg
}
func (m *Message) Go(cb interface{}, args ...interface{}) *Message {
	switch cb := cb.(type) {
	case func(*Message):
		return m.Gos(m.Spawn(), cb, args...)
	}
	return m.Gos(m, cb, args...)
}
