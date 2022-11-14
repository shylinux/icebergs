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

func _action_list(m *ice.Message, river, storm string) {
	m.Cmdy(STORM, kit.Dict(ice.MSG_RIVER, river, ice.MSG_STORM, storm))
}
func _action_exec(m *ice.Message, river, storm, index string, arg ...string) {
	m.Options(ice.MSG_RIVER, river, ice.MSG_STORM, storm)
	if m.Warn(m.Cmd(STORM, index, func(value ice.Maps) {
		if index = value[ctx.INDEX]; value[web.SPACE] != "" {
			m.Option(ice.POD, value[web.SPACE])
		}
	}).Length() == 0, ice.ErrNotFound, index) {
		return
	}
	if m.Option(ice.MSG_UPLOAD) != "" {
		_action_upload(m)
	}
	if !ctx.PodCmd(m, index, arg) {
		m.Cmdy(index, arg)
	}
}
func _action_auth(m *ice.Message, share string) *ice.Message {
	msg := m.Cmd(web.SHARE, share)
	if m.Warn(msg.Append(mdb.TIME) < m.Time(), ice.ErrNotValid) {
		msg.Append(mdb.TYPE, "")
		return msg
	}
	m.Tables(func(value ice.Maps) {
		aaa.SessAuth(m, value, RIVER, m.Option(ice.MSG_RIVER, msg.Append(RIVER)), STORM, m.Option(ice.MSG_STORM, msg.Append(STORM)))
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
			m.Push(TITLE, msg.Append(TITLE))
			m.Push(TOPIC, msg.Append(TOPIC))
			_action_list(m, msg.Append(web.RIVER), msg.Append(web.STORM))
			break
		}
		_action_exec(m, msg.Append(web.RIVER), msg.Append(web.STORM), arg[1], arg[2:]...)
	case web.FIELD:
		if len(arg) == 1 {
			m.Push(TITLE, msg.Append(TITLE))
			m.Push(TOPIC, msg.Append(TOPIC))
			m.Push(ctx.ARGS, msg.Append(mdb.TEXT))
			m.Cmdy(ctx.COMMAND, msg.Append(mdb.NAME))
			break
		}
		if arg[1] = msg.Append(mdb.NAME); m.Option(ice.MSG_UPLOAD) != "" {
			_action_upload(m)
		}
		m.Cmdy(arg[1:])
	}
}
func _action_upload(m *ice.Message) {
	msg := m.Cmd(web.CACHE, web.UPLOAD)
	m.Option(ice.MSG_UPLOAD, msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
}

const ACTION = "action"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(ACTION): {Name: "/action river storm action arg...", Help: "工作台", Actions: ice.MergeActions(ice.Actions{
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _storm_key(m), mdb.LIST, m.OptionSimple(mdb.ID), arg)
			}},
			web.SHARE: {Hand: func(m *ice.Message, arg ...string) { _action_share(m, arg...) }},
		}, ctx.CmdAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin, arg) {
				return
			}
			if m.Option(ice.MSG_USERPOD) == "" && m.Warn(!_river_right(m, arg[0]), ice.ErrNotRight, arg) {
				return
			}
			if len(arg) == 2 {
				m.OptionFromConfig(MENUS)
				_action_list(m, arg[0], arg[1])
			} else {
				_action_exec(m, arg[0], arg[1], arg[2], arg[3:]...)
			}
		}},
	})
}
