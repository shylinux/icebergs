package ice

import (
	"errors"
	"io"
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
			m.Result(ErrWarn, e)
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
func (m *Message) Assert(expr Any) bool {
	switch e := expr.(type) {
	case nil:
		return true
	case error:
	case bool:
		if e == true {
			return true
		}
	default:
		expr = errors.New(kit.Format("error: %v", e))
	}
	m.Result(ErrWarn, expr)
	panic(expr)
}
func (m *Message) Sleep(d string, arg ...Any) *Message {
	if time.Sleep(kit.Duration(d)); len(arg) > 0 {
		m.Cmdy(arg...)
	}
	return m
}
func (m *Message) Sleep300ms(arg ...Any) *Message { return m.Sleep("300ms", arg...) }
func (m *Message) Sleep30ms(arg ...Any) *Message  { return m.Sleep("30ms", arg...) }
func (m *Message) Sleep3s(arg ...Any) *Message    { return m.Sleep("3s", arg...) }
func (m *Message) Go(cb Any) *Message {
	task.Put(kit.FileLine(cb, 3), func(task *task.Task) error {
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
