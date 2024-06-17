package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _action_list(m *ice.Message, river, storm string) {
	m.Cmdy(STORM, kit.Dict(ice.MSG_RIVER, river, ice.MSG_STORM, storm))
}
func _action_exec(m *ice.Message, river, storm, index string, arg ...string) {
	m.Options(ice.MSG_RIVER, river, ice.MSG_STORM, storm)
	if m.WarnNotRight(m.Cmd(STORM, index, arg, func(value ice.Maps) {
		if index = value[ctx.INDEX]; value[web.SPACE] != "" {
			m.Option(ice.POD, value[web.SPACE])
		}
	}).Length() == 0, index, arg) {
		return
	}
	kit.If(!ctx.PodCmd(m, index, arg), func() { m.Cmdy(index, arg) })
}
func _action_auth(m *ice.Message, share string) *ice.Message {
	msg := m.Cmd(web.SHARE, share)
	if web.IsNotValidShare(m, msg.Append(mdb.TIME)) {
		msg.Append(mdb.TYPE, "")
		return msg
	}
	m.Table(func(value ice.Maps) {
		aaa.SessAuth(m, kit.Dict(value), RIVER, m.Option(ice.MSG_RIVER, msg.Append(RIVER)), STORM, m.Option(ice.MSG_STORM, msg.Append(STORM)))
	})
	if m.WarnNotRight(!_river_right(m, msg.Append(web.RIVER))) {
		msg.Append(mdb.TYPE, "")
		return msg
	}
	return msg
}
func _action_share(m *ice.Message, arg ...string) {
	switch msg := _action_auth(m, arg[0]); msg.Append(mdb.TYPE) {
	case web.STORM:
		if len(arg) == 1 {
			m.Push(TITLE, msg.Append(TITLE)).Push(THEME, msg.Append(THEME))
			_action_list(m, msg.Append(web.RIVER), msg.Append(web.STORM))
			break
		}
		_action_exec(m, msg.Append(web.RIVER), msg.Append(web.STORM), arg[1], arg[2:]...)
	case web.FIELD:
		m.Option(ice.MSG_USERPOD, kit.Keys(m.Option(ice.MSG_USERPOD), msg.Append(ice.POD)))
		if len(arg) == 1 {
			m.Push(TITLE, msg.Append(TITLE)).Push(THEME, msg.Append(THEME))
			m.Cmdy(web.Space(m, msg.Append(ice.POD)), ctx.COMMAND, msg.Append(mdb.NAME))
			m.Push(ctx.ARGS, msg.Append(mdb.TEXT))
			break
		}
		m.Cmdy(web.Space(m, msg.Append(ice.POD)), msg.Append(mdb.NAME), arg[2:])
	}
}

const ACTION = "action"

func init() {
	Index.MergeCommands(ice.Commands{
		ACTION: {Help: "工作区", Actions: ice.MergeActions(ice.Actions{
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m), mdb.LIST, m.OptionSimple(mdb.ID), arg)
			}},
			web.SHARE: {Hand: func(m *ice.Message, arg ...string) { _action_share(m, arg...) }},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(REWRITE, kit.Hashs(kit.Fields(m.Option(ice.POD), arg[0]))); msg.Length() > 0 {
					kit.If(msg.Append("to_space"), func(p string) { m.Option(ice.POD, p) })
					kit.If(msg.Append("to_index"), func(p string) { arg[0] = p })
					defer m.Push(web.SPACE, m.Option(ice.POD))
				}
				ctx.Command(m, arg...)
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.Run(m, arg...)
			}},
		}, web.ApiAction(""), aaa.WhiteAction("", web.SHARE)), Hand: func(m *ice.Message, arg ...string) {
			if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "", arg) {
				return
			} else if m.WarnNotRight(!_river_right(m, arg[0]), arg) {
				return
			}
			if len(arg) == 2 {
				ctx.OptionFromConfig(m, MENUS)
				_action_list(m, arg[0], arg[1])
			} else {
				_action_exec(m, arg[0], arg[1], arg[2], arg[3:]...)
			}
		}},
	})
}
