package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
)

func _action_domain(m *ice.Message, cmd string, arg ...string) (domain string) {
	m.Option(ice.MSG_LOCAL, "")
	m.Option(ice.MSG_DOMAIN, "")
	if m.Conf(ACTION, kit.Keys(kit.MDB_META, DOMAIN, cmd)) != "true" {
		return ""
	}

	storm := kit.Select(m.Option(ice.MSG_STORM), arg, 0)
	river := kit.Select(m.Option(ice.MSG_RIVER), arg, 1)
	m.Richs(RIVER, "", river, func(key string, value map[string]interface{}) {
		switch kit.Value(kit.GetMeta(value), kit.MDB_TYPE) {
		case PUBLIC:
			return
		case PROTECTED:
			m.Richs(RIVER, kit.Keys(kit.MDB_HASH, river, TOOL), storm, func(key string, value map[string]interface{}) {
				switch kit.Value(kit.GetMeta(value), kit.MDB_TYPE) {
				case PUBLIC:
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys("R"+river))
				case PROTECTED:
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys("R"+river, "S"+storm))
				case PRIVATE:
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys("R"+river, "U"+m.Option(ice.MSG_USERNAME)))
				}
			})
		case PRIVATE:
			domain = m.Option(ice.MSG_DOMAIN, kit.Keys("U"+m.Option(ice.MSG_USERNAME)))
		}
		m.Option(ice.MSG_LOCAL, path.Join(m.Conf(RIVER, kit.META_PATH), domain))
	})
	m.Log_AUTH(RIVER, river, STORM, storm, DOMAIN, domain)
	return
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
		m.Cmdy(m.Space(kit.Select(m.Option(POD), value[POD])), ctx.COMMAND, kit.Keys(value[CTX], value[CMD]))
	})
	m.SortInt(kit.MDB_ID)
}
func _action_show(m *ice.Message, river, storm, index string, arg ...string) {
	cmds := []string{index}
	prefix := kit.Keys(kit.MDB_HASH, river, TOOL, kit.MDB_HASH, storm)
	if m.Grows(RIVER, prefix, kit.MDB_ID, index, func(index int, value map[string]interface{}) {
		if cmds = kit.Simple(kit.Keys(value[CTX], value[CMD])); kit.Format(value[POD]) != "" {
			m.Option(POD, value[POD])
		}
	}) == nil && m.Warn(!m.Right(cmds), ice.ErrNotRight) {
		return
	}

	if _action_domain(m, cmds[0]); m.Option(ice.MSG_UPLOAD) != "" {
		_action_upload(m) // 上传文件
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
func _action_upload(m *ice.Message, arg ...string) {
	msg := m.Cmd(web.CACHE, web.UPLOAD)
	m.Option(ice.MSG_UPLOAD, msg.Append(kit.MDB_HASH), msg.Append(kit.MDB_NAME), msg.Append(kit.MDB_SIZE))
}

const (
	DOMAIN    = "domain"
	PUBLIC    = "public"
	PROTECTED = "protected"
	PRIVATE   = "private"
)

const P_ACTION = "/action"
const ACTION = "action"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ACTION: {Name: ACTION, Help: "应用", Value: kit.Data(DOMAIN, kit.Dict())},
		},
		Commands: map[string]*ice.Command{
			P_ACTION: {Name: "/action river storm action arg...", Help: "工作台", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					for _, k := range arg {
						m.Cmdy(ctx.COMMAND, k)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "_share" {
					switch msg := m.Cmd(web.SHARE, arg[1]); msg.Append(kit.MDB_TYPE) {
					case STORM:
						m.Option(kit.MDB_TITLE, msg.Append(kit.MDB_NAME))
						arg[0] = msg.Append(RIVER)
						arg[1] = msg.Append(STORM)
					default:
						return
					}
				} else {
					if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
						return // 没有登录
					}
					if m.Warn(!_action_right(m, arg[0], arg[1]), ice.ErrNotRight) {
						return // 没有授权
					}
				}

				m.Option(ice.MSG_RIVER, arg[0])
				m.Option(ice.MSG_STORM, arg[1])

				if len(arg) == 2 {
					_action_list(m, arg[0], arg[1])
					return //命令列表
				}
				_action_show(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
		}})
}
