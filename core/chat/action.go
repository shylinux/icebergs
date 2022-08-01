package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _action_right(m *ice.Message, river string, storm string) (ok bool) {
	if ok = true; m.Option(ice.MSG_USERROLE) == aaa.VOID {
		m.Richs(RIVER, "", river, func(key string, value ice.Map) {
			if ok = m.Richs(RIVER, kit.Keys(mdb.HASH, key, OCEAN), m.Option(ice.MSG_USERNAME), nil) != nil; ok {
				m.Log_AUTH(RIVER, river, STORM, storm)
			}
		})
	}
	return ok
}
func _action_key(m *ice.Message, arg ...string) string {
	return kit.Keys(mdb.HASH, kit.Select(m.Option(RIVER), arg, 0), STORM, mdb.HASH, kit.Select(m.Option(STORM), arg, 1))
}
func _action_list(m *ice.Message, river, storm string) {
	m.Cmdy(STORM, storm, ice.Option{ice.MSG_RIVER, river}).Tables(func(value ice.Maps) {
		m.Cmdy(m.Space(kit.Select(m.Option(ice.POD), value[ice.POD])), ctx.COMMAND, kit.Keys(value[ice.CTX], value[ice.CMD]))
	})
}
func _action_exec(m *ice.Message, river, storm, index string, arg ...string) {
	m.Option(ice.MSG_RIVER, river)
	m.Option(ice.MSG_STORM, storm)

	cmds := []string{index}
	if m.Grows(RIVER, _action_key(m, river, storm), mdb.ID, index, func(index int, value ice.Map) {
		if cmds = kit.Simple(kit.Keys(value[ice.CTX], value[ice.CMD])); kit.Format(value[ice.POD]) != "" {
			m.Option(ice.POD, value[ice.POD]) // 远程节点
		}
	}) == nil && m.Option(ice.MSG_USERPOD) == "" && !m.Right(cmds) {
		return // 没有授权
	}

	if _action_domain(m, cmds[0]); m.Option(ice.MSG_UPLOAD) != "" {
		_action_upload(m) // 上传文件
	}

	if cmds[0] == "web.chat.node" || !m.PodCmd(cmds, arg) {
		m.Cmdy(cmds, arg) // 执行命令
	}
}
func _action_auth(m, msg *ice.Message) bool {
	if m.Warn(kit.Time() > kit.Time(msg.Append(mdb.TIME)), ice.ErrNotValid) {
		return false
	}
	m.Log_AUTH(
		aaa.USERROLE, m.Option(ice.MSG_USERROLE, msg.Append(aaa.USERROLE)),
		aaa.USERNAME, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME)),
		aaa.USERNICK, m.Option(ice.MSG_USERNICK, msg.Append(aaa.USERNICK)),
		RIVER, m.Option(ice.MSG_RIVER, msg.Append(RIVER)),
		STORM, m.Option(ice.MSG_STORM, msg.Append(STORM)),
	)
	return _action_right(m, msg.Append(web.RIVER), msg.Append(web.STORM))
}
func _action_share(m *ice.Message, arg ...string) {
	switch msg := m.Cmd(web.SHARE, arg[0]); msg.Append(mdb.TYPE) {
	case web.STORM:
		if !_action_auth(m, msg) {
			break // 没有授权
		}

		if len(arg) == 1 {
			_action_list(m, msg.Append(web.RIVER), msg.Append(web.STORM))
			break // 命令列表
		}

		// 执行命令
		_action_exec(m, msg.Append(web.RIVER), msg.Append(web.STORM), arg[1], arg[2:]...)

	case web.FIELD:
		if !_action_auth(m, msg) {
			break // 没有授权
		}

		if arg[0] = msg.Append(mdb.NAME); len(arg) == 1 {
			m.Push(TITLE, msg.Append(TITLE))
			m.Push(TOPIC, msg.Append(TOPIC))
			m.Push(ctx.INDEX, msg.Append(mdb.NAME))
			m.Push(ctx.ARGS, msg.Append(mdb.TEXT))
			break // 命令列表
		}

		if _action_domain(m, arg[1]); m.Option(ice.MSG_UPLOAD) != "" {
			_action_upload(m) // 上传文件
		}

		m.Cmdy(arg[1:]) // 执行命令
	}
}
func _action_domain(m *ice.Message, cmd string, arg ...string) (domain string) {
	m.Option(ice.MSG_LOCAL, "")
	m.Option(ice.MSG_DOMAIN, "")
	if m.Config(kit.Keys(DOMAIN, cmd)) != ice.TRUE {
		return "" // 公有命令
	}

	river := kit.Select(m.Option(ice.MSG_RIVER), arg, 0)
	storm := kit.Select(m.Option(ice.MSG_STORM), arg, 1)
	m.Richs(RIVER, "", river, func(key string, value ice.Map) {
		switch kit.Value(kit.GetMeta(value), mdb.TYPE) {
		case PUBLIC: // 公有群
			return
		case PROTECTED: // 共有群
			m.Richs(RIVER, kit.Keys(mdb.HASH, river, STORM), storm, func(key string, value ice.Map) {
				switch r := "R" + river; kit.Value(kit.GetMeta(value), mdb.TYPE) {
				case PUBLIC: // 公有组
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys(r))
				case PROTECTED: // 共有组
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys(r, "S"+storm))
				case PRIVATE: // 私有组
					domain = m.Option(ice.MSG_DOMAIN, kit.Keys(r, "U"+m.Option(ice.MSG_USERNAME)))
				}
			})
		case PRIVATE: // 私有群
			domain = m.Option(ice.MSG_DOMAIN, kit.Keys("U"+m.Option(ice.MSG_USERNAME)))
		}
		m.Option(ice.MSG_LOCAL, path.Join(m.Config(nfs.PATH), domain))
	})
	m.Log_AUTH(RIVER, river, STORM, storm, DOMAIN, domain)
	return
}
func _action_upload(m *ice.Message) {
	msg := m.Cmd(web.CACHE, web.UPLOAD)
	m.Option(ice.MSG_UPLOAD, msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
}

const (
	DOMAIN    = "domain"
	PUBLIC    = "public"
	PROTECTED = "protected"
	PRIVATE   = "private"
)

const ACTION = "action"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(ACTION): {Name: "/action river storm action arg...", Help: "工作台", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{
					"web.chat.meet.miss",
					"web.chat.meet.mate",
					"web.chat.location",
					"web.chat.paste",
					"web.chat.scan",
					"web.wiki.feel",
					"web.wiki.draw",
					"web.wiki.data",
					"web.wiki.word",
					"web.team.task",
					"web.team.plan",
					"web.mall.asset",
					"web.mall.salary",
				} {
					m.Config(kit.Keys(DOMAIN, cmd), ice.TRUE)
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _action_key(m), mdb.LIST, m.OptionSimple(mdb.ID), arg)
			}},
			SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_header_share(m, arg...)
			}},
			"_share": {Name: "_share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
				_action_share(m, arg...)
			}},
		}, ctx.CmdAction(nfs.PATH, ice.USR_LOCAL_RIVER)), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg) {
				return // 没有登录
			}
			if m.Option(ice.MSG_USERPOD) == "" && m.Warn(!_action_right(m, arg[0], arg[1]), ice.ErrNotRight, arg) {
				return // 没有授权
			}

			if len(arg) == 2 {
				m.Option(MENUS, m.Config(MENUS))
				_action_list(m, arg[0], arg[1])
				return //命令列表
			}

			// 执行命令
			_action_exec(m, arg[0], arg[1], arg[2], arg[3:]...)
		}},
	})
}
