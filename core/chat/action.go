package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _action_right(m *ice.Message, river string, storm string) (ok bool) {
	return m.Option(ice.MSG_USERROLE) != aaa.VOID || m.Cmd(OCEAN, m.Option(ice.MSG_USERNAME)).Length() > 0
}
func _action_key(m *ice.Message, arg ...string) string {
	return kit.KeyHash(kit.Select(m.Option(RIVER), arg, 0), STORM, kit.KeyHash(kit.Select(m.Option(STORM), arg, 1)))
}
func _action_list(m *ice.Message, river, storm string) {
	m.Cmdy(STORM, storm, kit.Dict(ice.MSG_RIVER, river))
}
func _action_exec(m *ice.Message, river, storm, index string, arg ...string) {
	m.Option(ice.MSG_RIVER, river)
	m.Option(ice.MSG_STORM, storm)

	if m.Cmd(STORM, storm, index, func(value ice.Maps) {
		if index = value[ctx.INDEX]; value[web.SPACE] != "" {
			m.Option(ice.POD, value[web.SPACE])
		}
	}).Length() == 0 && m.Option(ice.MSG_USERPOD) == "" && !aaa.Right(m, index) {
		return // 没有授权
	}

	if m.Option(ice.MSG_UPLOAD) != "" {
		_action_upload(m) // 上传文件
	}

	if index == m.Prefix(NODES) || !ctx.PodCmd(m, index, arg) {
		m.Cmdy(index, arg) // 执行命令
	}
}
func _action_auth(m *ice.Message, share string) *ice.Message {
	msg := m.Cmd(web.SHARE, share)
	if m.Warn(kit.Time(msg.Append(mdb.TIME)) < kit.Time(m.Time()), ice.ErrNotValid) {
		msg.Append(mdb.TYPE, "")
		return msg // 共享过期
	}
	m.Auth(
		aaa.USERROLE, m.Option(ice.MSG_USERROLE, msg.Append(aaa.USERROLE)),
		aaa.USERNAME, m.Option(ice.MSG_USERNAME, msg.Append(aaa.USERNAME)),
		aaa.USERNICK, m.Option(ice.MSG_USERNICK, msg.Append(aaa.USERNICK)),
		RIVER, m.Option(ice.MSG_RIVER, msg.Append(RIVER)),
		STORM, m.Option(ice.MSG_STORM, msg.Append(STORM)),
	)
	if m.Warn(!_action_right(m, msg.Append(web.RIVER), msg.Append(web.STORM)), ice.ErrNotRight) {
		msg.Append(mdb.TYPE, "")
		return msg // 没有权限
	}
	return msg
}
func _action_share(m *ice.Message, arg ...string) {
	switch msg := _action_auth(m, arg[0]); msg.Append(mdb.TYPE) {
	case web.STORM:
		if len(arg) == 1 {
			_action_list(m, msg.Append(web.RIVER), msg.Append(web.STORM))
			break // 命令列表
		}

		// 执行命令
		_action_exec(m, msg.Append(web.RIVER), msg.Append(web.STORM), arg[1], arg[2:]...)

	case web.FIELD:
		if len(arg) == 1 {
			m.Push(TITLE, msg.Append(TITLE))
			m.Push(TOPIC, msg.Append(TOPIC))
			m.Push(ctx.INDEX, msg.Append(mdb.NAME))
			m.Push(ctx.ARGS, msg.Append(mdb.TEXT))
			break // 命令列表
		}
		if arg[1] = msg.Append(mdb.NAME); m.Option(ice.MSG_UPLOAD) != "" {
			_action_upload(m) // 上传文件
		}
		m.Cmdy(arg[1:]) // 执行命令
	}
}
func _action_upload(m *ice.Message) {
	msg := m.Cmd(web.CACHE, web.UPLOAD)
	m.Option(ice.MSG_UPLOAD, msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
}

const (
	PUBLIC    = "public"
	PROTECTED = "protected"
	PRIVATE   = "private"
)

const ACTION = "action"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(ACTION): {Name: "/action river storm action arg...", Help: "工作台", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.CommandKey())
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _action_key(m), mdb.LIST, m.OptionSimple(mdb.ID), arg)
			}},
			web.SHARE: {Name: "share", Help: "共享", Hand: func(m *ice.Message, arg ...string) {
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
