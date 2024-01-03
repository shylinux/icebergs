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
	if m.Warn(m.Cmd(STORM, index, arg, func(value ice.Maps) {
		if index = value[ctx.INDEX]; value[web.SPACE] != "" {
			m.Option(ice.POD, value[web.SPACE])
		}
	}).Length() == 0, ice.ErrNotRight, index, arg) {
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
	if m.Warn(!_river_right(m, msg.Append(web.RIVER)), ice.ErrNotRight) {
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
		}, web.ApiAction(""), aaa.WhiteAction("", web.SHARE)), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg) {
				return
			} else if m.Warn(!_river_right(m, arg[0]), ice.ErrNotRight, arg) {
				return
			}
			if web.PodCmd(m, web.SPACE, arg...) {
				m.Table(func(value ice.Maps) { m.Push(web.SPACE, m.Option(ice.MSG_USERPOD)) })
			} else if len(arg) == 2 {
				ctx.OptionFromConfig(m, MENUS)
				_action_list(m, arg[0], arg[1])
			} else {
				_action_exec(m, arg[0], arg[1], arg[2], arg[3:]...)
			}
		}},
	})
}
