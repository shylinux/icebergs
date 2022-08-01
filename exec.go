package ice

import (
	"errors"
	"io"
	"path"
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
func (m *Message) Assert(expr Any) bool {
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
func (m *Message) Sleep30s(arg ...Any) *Message   { return m.Sleep("30s", arg...) }
func (m *Message) Call(sync bool, cb func(*Message) *Message) *Message {
	wait := make(chan bool, 2)

	p := kit.Select("10s", m.Option("timeout"))
	t := time.AfterFunc(kit.Duration(p), func() {
		m.Warn(true, ErrNotValid, m.Detailv())
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

func (m *Message) Go(cb Any, arg ...string) *Message {
	task.Put(kit.Select(kit.FileLine(cb, 3), arg, 0), func(task *task.Task) error {
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

func (m *Message) Watch(key string, arg ...string) *Message {
	if len(arg) == 0 {
		arg = append(arg, m.Prefix(AUTO))
	}
	m.Cmd(EVENT, ACTION, LISTEN, EVENT, key, CMD, kit.Join(arg, SP))
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd(EVENT, ACTION, HAPPEN, EVENT, key, arg)
	return m
}
func (m *Message) Right(arg ...Any) bool {
	key := path.Join(strings.ReplaceAll(kit.Keys(arg...), PT, PS))
	key = strings.TrimPrefix(strings.TrimSuffix(strings.ReplaceAll(key, PS, PT), PT), PT)
	return m.Option(MSG_USERROLE) == ROOT || !m.Warn(m.Cmdx(ROLE, RIGHT, m.Option(MSG_USERROLE), key) != OK,
		ErrNotRight, kit.Join(kit.Simple(arg), PT), USERROLE, m.Option(MSG_USERROLE), FILELINE, kit.FileLine(2, 3))
}
func (m *Message) Space(arg Any) []string {
	if arg == nil || arg == "" || kit.Format(arg) == Info.NodeName {
		return nil
	}
	return []string{SPACE, kit.Format(arg)}
}
func (m *Message) PodCmd(arg ...Any) bool {
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
