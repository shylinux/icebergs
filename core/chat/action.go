package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _action_upload(m *ice.Message, arg ...string) {
	msg := m.Cmd(web.CACHE, web.UPLOAD)
	m.Debug(msg.Formats("meta"))
	m.Option("_upload", msg.Append("data"), msg.Append("name"))
	m.Debug(m.Formats("meta"))

}
func _action_right(m *ice.Message, river string, storm string) (ok bool) {
	if ok = true; m.Option(ice.MSG_USERROLE) == aaa.VOID {
		m.Richs(RIVER, "", river, func(key string, value map[string]interface{}) {
			if ok = m.Richs(RIVER, kit.Keys(kit.MDB_HASH, key, USER), m.Option(ice.MSG_USERNAME), nil) != nil; ok {
				m.Log_AUTH(RIVER, river, STORM, storm)
			}
		})
	}
	return ok
}
func _action_list(m *ice.Message, river, storm string) {
	m.Option(ice.MSG_RIVER, river)
	m.Cmdy(TOOL, storm).Table(func(index int, value map[string]string, head []string) {
		m.Cmdy(m.Space(value[POD]), ctx.COMMAND, kit.Keys(value[CTX], value[CMD]))
	})
}
func _action_show(m *ice.Message, river, storm, index string, arg ...string) {
	cmds := []string{index}
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	if m.Grows(RIVER, prefix, kit.MDB_ID, index, func(index int, value map[string]interface{}) {
		if cmds = kit.Simple(kit.Keys(value[CTX], value[CMD])); kit.Format(value[POD]) != "" {
			m.Option(kit.GDB_POD, value[POD])
		}
	}) == nil && m.Warn(!m.Right(cmds), ice.ErrNotAuth) {
		return
	}
	m.Cmdy(_action_proxy(m), cmds, arg)
}
func _action_proxy(m *ice.Message) (proxy []string) {
	if p := m.Option(POD); p != "" {
		proxy = append(proxy, web.SPACE, p)
		m.Option(POD, "")
	}
	return proxy
}

const ACTION = "action"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/action": {Name: "/action", Help: "工作台", Action: map[string]*ice.Action{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				for _, k := range arg {
					m.Cmdy(ctx.COMMAND, k)
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
				return // 没有登录
			}
			if m.Warn(!_action_right(m, m.Option(ice.MSG_RIVER, arg[0]), m.Option(ice.MSG_STORM, arg[1])), ice.ErrNotAuth) {
				return // 没有授权
			}

			if len(arg) == 2 {
				_action_list(m, arg[0], arg[1])
				return //命令列表
			}

			if m.Option("_upload") != "" {
				_action_upload(m)
			}
			_action_show(m, arg[0], arg[1], arg[2], arg[3:]...)
		}},
	}}, nil)
}
