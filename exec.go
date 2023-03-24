package ice

import (
	"errors"
	"io"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/task"
)

func (m *Message) TryCatch(msg *Message, catch bool, cb ...func(msg *Message)) *Message {
	defer func() {
		switch e := recover(); e {
		case nil, io.EOF:
		default:
			fileline := m.FormatStack(2, 1)
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Log("chain", msg.FormatChain())
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Log("stack", m.FormatStack(2, 100))
			m.Log(LOG_WARN, "catch: %s %s", e, fileline).Result(ErrWarn, e, SP, m.FormatStack(2, 5))
			if len(cb) > 1 {
				m.TryCatch(msg, catch, cb[1:]...)
			} else if !catch {
				m.Assert(e)
			}
		}
	}()
	kit.If(len(cb) > 0, func() { cb[0](msg) })
	return m
}
func (m *Message) Assert(expr Any) bool {
	switch e := expr.(type) {
	case nil:
		return true
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
	if time.Sleep(kit.Duration(d)); len(arg) > 0 {
		m.Cmdy(arg...)
	}
	return m
}
func (m *Message) Sleep300ms(arg ...Any) *Message { return m.Sleep("300ms", arg...) }
func (m *Message) Sleep30ms(arg ...Any) *Message  { return m.Sleep("30ms", arg...) }
func (m *Message) Sleep3s(arg ...Any) *Message    { return m.Sleep("3s", arg...) }
func (m *Message) Go(cb Any, arg ...Any) *Message {
	kit.If(len(arg) == 0, func() { arg = append(arg, logs.FileLine(cb)) })
	task.Put(arg[0], func(task *task.Task) error {
		m.TryCatch(m, true, func(m *Message) {
			switch cb := cb.(type) {
			case func(*Message):
				cb(m.Spawn())
			case func():
				cb()
			default:
				m.ErrorNotImplement(cb)
			}
		})
		return nil
	})
	return m
}
