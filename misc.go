package ice

import (
	kit "github.com/shylinux/toolkits"

	"fmt"
	"strings"
	"sync/atomic"
)

func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) Save(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd("ctx.config", "save", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}
func (m *Message) Load(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, kit.Keys(m.Cap(CTX_FOLLOW), k))
	}
	m.Cmd("ctx.config", "load", kit.Keys(m.Cap(CTX_FOLLOW), "json"), list)
	return m
}

func (m *Message) Watch(key string, arg ...string) *Message {
	if len(arg) == 0 {
		arg = append(arg, m.Prefix("auto"))
	}
	m.Cmd("gdb.event", "listen", key, arg)
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd("gdb.event", "action", key, arg)
	return m
}
func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == "root" || !m.Warn(m.Cmdx("aaa.role", "right", m.Option(MSG_USERROLE), kit.Keys(arg...)) != "ok", ErrNotAuth, strings.Join(kit.Simple(arg), "."))
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || arg == "" || kit.Format(arg) == m.Conf("cli.runtime", "node.name") {
		return nil
	}
	return []string{"web.space", kit.Format(arg)}
}

var count = int32(0)

func (m *Message) AddCmd(cmd *Command) string {
	name := fmt.Sprintf("_cb_%d", atomic.AddInt32(&count, 1))
	m.target.Commands[name] = cmd
	return kit.Keys(m.target.Cap(CTX_FOLLOW), name)
}

func (m *Message) PushAction(list ...interface{}) {
	m.Table(func(index int, value map[string]string, head []string) {
		action := []string{}
		for _, k := range kit.Simple(list...) {
			action = append(action, m.Cmdx("render", "button", k))
		}
		m.Push("action", strings.Join(action, ""))
	})
}
func (m *Message) PushDetail(value interface{}, arg ...interface{}) *Message {
	return m.Push("detail", value, arg...)
}
