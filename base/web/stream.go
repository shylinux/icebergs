package web

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _stream_subkey(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = append(arg, kit.Hashs(kit.Fields(m.Option(ice.MSG_SPACE), m.Option(ice.MSG_INDEX)))) })
	return m.Options(mdb.SUBKEY, kit.Keys(mdb.HASH, arg[0]), ice.MSG_FIELDS, mdb.Config(m, mdb.FIELDS))
}

const STREAM = "stream"

func init() {
	Index.MergeCommands(ice.Commands{
		STREAM: {Name: "stream hash daemon auto", Help: "推送流", Actions: ice.MergeActions(ice.Actions{
			ONLINE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_DAEMON) == "" {
					return
				}
				mdb.HashCreate(m, SPACE, m.Option(ice.MSG_SPACE), ctx.INDEX, m.Option(ice.MSG_INDEX), mdb.SHORT, cli.DAEMON, mdb.FIELD, mdb.Config(m, mdb.FIELDS))
				mdb.HashCreate(_stream_subkey(m), ParseUA(m))
				mdb.HashSelect(m)
			}},
			"push": {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(_stream_subkey(m)).Table(func(value ice.Maps) {
					if value[cli.DAEMON] != m.Option(ice.MSG_DAEMON) {
						m.Options(mdb.SUBKEY, "").Cmd(SPACE, value[cli.DAEMON], arg)
					}
				})
			}},
			PORTAL_CLOSE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m).Table(func(value ice.Maps) {
					mdb.HashSelect(_stream_subkey(m, value[mdb.HASH]).Spawn()).Table(func(value ice.Maps) {
						if strings.HasPrefix(value[cli.DAEMON], m.Option(mdb.NAME)) {
							mdb.HashRemove(m, mdb.HASH, kit.Hashs(value[cli.DAEMON]))
						}
					})
				})
			}},
		}, gdb.EventsAction(PORTAL_CLOSE), mdb.ClearOnExitHashAction(), mdb.HashAction(
			mdb.SHORT, "space,index", mdb.FIELD, "time,hash,space,index",
			mdb.FIELDS, "time,daemon,userrole,username,usernick,avatar,icons,agent,system,ip,ua",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m)
			} else {
				mdb.HashSelect(_stream_subkey(m, arg[0]), arg[1:]...)
			}
		}},
	})
}
func StreamPush(m *ice.Message, arg ...string) {
	if ice.Info.NodeType == WORKER {
		m.Option(ice.MSG_SPACE, m.Option(ice.MSG_USERPOD))
	} else {
		m.Option(ice.MSG_SPACE, "")
	}
	m.Option(ice.MSG_INDEX, m.ShortKey())
	AdminCmd(m, STREAM, "push", arg)
}
func StreamPushRefreshConfirm(m *ice.Message, arg ...string) {
	StreamPush(m.Spawn(ice.Maps{"space.noecho": "true"}), kit.Simple(html.REFRESH, html.CONFIRM, arg)...)
}
