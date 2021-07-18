package chat

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
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
func _action_share(m *ice.Message, arg ...string) {
	switch msg := m.Cmd(web.SHARE, arg[1]); msg.Append(kit.MDB_TYPE) {
	case web.STORM:
		if len(arg) == 2 {
			_action_list(m, msg.Append(web.RIVER), msg.Append(web.STORM))
			return
		}

		if m.Warn(kit.Time() > kit.Time(msg.Append(kit.MDB_TIME)), ice.ErrExpire) {
			break // 分享超时
		}
		m.Log_AUTH(
			aaa.USERROLE, m.Option(ice.MSG_USERROLE, msg.Append(aaa.USERROLE)),
			aaa.USERNAME, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME)),
		)

		_action_show(m, msg.Append(web.RIVER), msg.Append(web.STORM), arg[2], arg[3:]...)

	case web.FIELD:
		if cmd := kit.Keys(msg.Append(web.RIVER), msg.Append(web.STORM)); len(arg) == 2 {
			m.Push("index", cmd)
			m.Push("title", msg.Append(kit.MDB_NAME))
			m.Push("args", msg.Append(kit.MDB_TEXT))
			break
		}

		if m.Warn(kit.Time() > kit.Time(msg.Append(kit.MDB_TIME)), ice.ErrExpire) {
			break // 分享超时
		}
		m.Log_AUTH(
			aaa.USERROLE, m.Option(ice.MSG_USERROLE, msg.Append(aaa.USERROLE)),
			aaa.USERNAME, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME)),
		)

		if m.Warn(!m.Right(arg[2:]), ice.ErrNotRight) {
			break // 没有授权
		}

		if m.Option(ice.MSG_UPLOAD) != "" {
			_action_upload(m) // 上传文件
		}
		m.Cmdy(arg[2:])
	}
}

func _action_list(m *ice.Message, river, storm string) {
	m.Option(ice.MSG_RIVER, river)
	m.Cmdy(TOOL, storm).Table(func(index int, value map[string]string, head []string) {
		m.Cmdy(m.Space(kit.Select(m.Option(POD), value[POD])), ctx.COMMAND, kit.Keys(value[CTX], value[CMD]))
	})
	m.SortInt(kit.MDB_ID)
}
func _action_show(m *ice.Message, river, storm, index string, arg ...string) {
	m.Option(ice.MSG_RIVER, river)
	m.Option(ice.MSG_STORM, storm)

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
						m.Cmdy(ctx.COMMAND, strings.TrimPrefix(k, "."))
					}
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, RIVER, kit.Keys(kit.MDB_HASH, m.Option(RIVER), TOOL, kit.MDB_HASH, m.Option(STORM)), mdb.LIST,
						kit.MDB_ID, m.Option(kit.MDB_ID), arg)
				}},
				SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
					_header_share(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "_share" {
					_action_share(m, arg...)
					return
				}

				if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
					return // 没有登录
				}
				if m.Warn(!_action_right(m, arg[0], arg[1]), ice.ErrNotRight) {
					return // 没有授权
				}

				if len(arg) == 2 {
					_action_list(m, arg[0], arg[1])
					return //命令列表
				}

				_action_show(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			"/cmd/": {Name: "/cmd/", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.RenderDownload(path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, kit.MDB_PATH)), "page/cmd.html"))
			}},
			"/cmd": {Name: "/cmd", Help: "命令", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(ctx.COMMAND, arg)
				}},
				cli.RUN: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Debug("waht %v %v", cmd, arg)
			}},
		}})
}
