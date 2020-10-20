package ice

import (
	"path"

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
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
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
	m.Cmd("gdb.event", "action", "listen", "event", key, "cmd", strings.Join(arg, " "))
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd("gdb.event", "action", "action", "event", key, arg)
	return m
}
func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == "root" || !m.Warn(m.Cmdx("aaa.role", "right", m.Option(MSG_USERROLE), strings.ReplaceAll(kit.Keys(arg...), "/", ".")) != "ok", ErrNotAuth, m.Option(MSG_USERROLE), " of ", strings.Join(kit.Simple(arg), "."))
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || arg == "" || kit.Format(arg) == m.Conf("cli.runtime", "node.name") {
		return nil
	}
	return []string{"web.space", kit.Format(arg)}
}

func (m *Message) PushPlugin(key string, arg ...string) *Message {
	m.Option("_process", "_field")
	m.Option("_prefix", arg)
	m.Cmdy("command", key)
	return m
}
func (m *Message) PushRender(key, view, name string, arg ...string) *Message {
	if m.Option(MSG_USERUA) == "" {
		return m
	}

	if strings.Contains(m.Option(MSG_USERUA), "curl") {
		return m
	}

	switch view {
	case "button":
		list := []string{}
		for _, k := range kit.Split(name) {
			list = append(list, fmt.Sprintf(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m.cmd.Meta, kit.Keys("trans", k)))))
		}
		m.Push(key, strings.Join(list, ""))
	case "video":
		m.Push(key, fmt.Sprintf(`<video src="%s" height=%s controls>`, name, kit.Select("120", arg, 0)))
	case "img":
		m.Push(key, fmt.Sprintf(`<img src="%s" height=%s>`, name, kit.Select("120", arg, 0)))
	case "a":
		m.Push(key, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, kit.Select(name, arg, 0), name))
	case "download":
		m.Push(key, fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, kit.Select(name, arg, 0), path.Base(name), name))
	default:
		m.Push(key, name)
	}
	return m
}
func (m *Message) PushButton(arg ...string) {
	m.PushRender("action", "button", strings.Join(arg, ","))
}
func (m *Message) PushAnchor(name string, arg ...string) {
	m.PushRender("link", "a", name, arg...)
}
func (m *Message) PushDownload(name string, arg ...string) {
	if len(arg) == 0 {
		name = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/local", name), "pod", m.Option(MSG_USERPOD))
	} else {
		arg[0] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/local", arg[0]), "pod", m.Option(MSG_USERPOD))
	}
	m.PushRender("link", "download", name, arg...)
}
func (m *Message) PushAction(list ...interface{}) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.PushRender(kit.MDB_ACTION, kit.MDB_BUTTON, strings.Join(kit.Simple(list...), ","))
	})
}

func (m *Message) SortStr(key string)   { m.Sort(key, "str") }
func (m *Message) SortStrR(key string)  { m.Sort(key, "str_r") }
func (m *Message) SortInt(key string)   { m.Sort(key, "int") }
func (m *Message) SortIntR(key string)  { m.Sort(key, "int_r") }
func (m *Message) SortTime(key string)  { m.Sort(key, "time") }
func (m *Message) SortTimeR(key string) { m.Sort(key, "time_r") }

var count = int32(0)

func (m *Message) AddCmd(cmd *Command) string {
	name := fmt.Sprintf("_cb_%d", atomic.AddInt32(&count, 1))
	m.target.Commands[name] = cmd
	return kit.Keys(m.target.Cap(CTX_FOLLOW), name)
}
