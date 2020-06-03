package ice

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	kit "github.com/shylinux/toolkits"
)

func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case io.EOF:
		case nil:
		default:
			_, file, line, _ := runtime.Caller(3)
			if list := strings.Split(file, "/"); len(list) > 2 {
				file = strings.Join(list[len(list)-2:], "/")
			}
			m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line)
			m.Log(LOG_INFO, "chain: %s", msg.Format("chain"))
			m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line)
			m.Log(LOG_INFO, "stack: %s", msg.Format("stack"))
			if m.Log(LOG_WARN, "catch: %s %s:%d", e, file, line); len(hand) > 1 {
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
	}

	// 抛出异常
	panic(errors.New(fmt.Sprintf("error %v", arg)))
}
func (m *Message) Sleep(arg string) *Message {
	time.Sleep(kit.Duration(arg))
	return m
}
func (m *Message) Hold(n int) *Message {
	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	ctx.wg.Add(n)
	m.Log(LOG_TRACE, "%s wait %s %v", ctx.Name, m.target.Name, ctx.wg)
	return m
}
func (m *Message) Done() bool {
	defer func() { recover() }()

	ctx := m.target.root
	if c := m.target; c.context != nil && c.context.wg != nil {
		ctx = c.context
	}

	m.Log(LOG_TRACE, "%s done %s %v", ctx.Name, m.target.Name, ctx.wg)
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
func (m *Message) Gos(msg *Message, cb func(*Message)) *Message {
	go func() { msg.TryCatch(msg, true, func(msg *Message) { cb(msg) }) }()
	return m
}
