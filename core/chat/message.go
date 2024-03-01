package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const MESSAGE = "message"

func init() {
	Index.MergeCommands(ice.Commands{
		MESSAGE: {Name: "message", Help: "聊天", Icon: "Messages.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				messageCreate(m, web.DREAM, "usr/icons/Launchpad.png")
				messageCreate(m, cli.SYSTEM, "usr/icons/System Settings.png")
				messageInsert(m, cli.SYSTEM, mdb.TYPE, "text", mdb.NAME, cli.RUNTIME, mdb.TEXT, m.Cmdx(cli.RUNTIME), ctx.DISPLAY, "/plugin/story/json.js")
			}},
			mdb.CREATE: {Name: "create type=tech,void title icons target", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(web.TARGET), "from.") {
					m.Option(web.TARGET, strings.Replace(m.Option(web.TARGET), "from", m.Option(ice.FROM_SPACE), 1))
				}
				mdb.ZoneCreate(m, kit.Simple(arg, web.TARGET, m.Option(web.TARGET), mdb.ZONE, kit.Select(kit.Hashs(mdb.UNIQ), m.Option(web.TARGET))))
			}},
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, append(arg, "direct", tcp.SEND, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.AVATAR, m.Option(ice.MSG_AVATAR)))
				kit.If(mdb.HashSelectField(m, arg[0], web.TARGET), func(p string) { m.Cmd(web.SPACE, p, MESSAGE, tcp.RECV, arg[1:]) })
				mdb.HashSelectUpdate(m, arg[0], func(value ice.Map) { kit.Value(value, mdb.TIME, m.Time()) })
				web.StreamPushRefreshConfirm(m, m.Trans("refresh for new message ", "刷新列表查看新消息 "))
			}},
			tcp.RECV: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, kit.Simple(mdb.ZONE, m.Option(ice.FROM_SPACE), arg, "direct", tcp.RECV, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.AVATAR, m.Option(ice.MSG_AVATAR)))
				mdb.HashSelectUpdate(m, m.Option(ice.FROM_SPACE), func(value ice.Map) { kit.Value(value, web.TARGET, m.Option(ice.FROM_SPACE)) })
				mdb.HashSelectUpdate(m, m.Option(ice.FROM_SPACE), func(value ice.Map) { kit.Value(value, mdb.TIME, m.Time()) })
				web.StreamPushRefreshConfirm(m, m.Trans("refresh for new message ", "刷新列表查看新消息 "))
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if ice.Info.Important {
					messageInsert(m, web.DREAM, mdb.TYPE, "plug", ctx.INDEX, IFRAME, ctx.ARGS, web.S(m.Option(mdb.NAME)))
				}
			}},
			web.OPEN: {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.MergePod(m.Option(web.TARGET))) }},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option("direct") == "recv" {
					m.Cmdy(web.Space(m, m.Option(web.TARGET)), ctx.COMMAND, arg[0]).ProcessField(ctx.ACTION, ctx.RUN, m.Option(web.TARGET), arg[0])
				} else {
					m.Cmdy(ctx.COMMAND, arg[0]).ProcessField(ctx.ACTION, ctx.RUN, "", arg[0])
				}
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.Space(m, arg[0]), arg[1], arg[2:]) }},
		}, web.DreamAction(), web.DreamTablesAction(), mdb.ZoneAction(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,hash,type,zone,icons,title,count,target",
			mdb.FIELDS, "time,id,type,name,text,space,index,args,style,display,username,usernick,avatar,direct",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.ZoneSelect(m.Display("").Spawn(), arg...).Table(func(value ice.Maps) {
					if kit.IsIn(m.Option(ice.MSG_USERROLE), value[mdb.TYPE], aaa.TECH, aaa.ROOT) {
						m.PushRecord(value, mdb.Config(m, mdb.FIELD))
					}
					if value[web.TARGET] == "" {
						m.PushButton(mdb.REMOVE)
					} else {
						m.PushButton(web.OPEN, mdb.REMOVE)
					}
				})
				m.Sort(mdb.TIME, ice.STR_R)
			} else {
				mdb.ZoneSelect(m, arg...).Sort(mdb.ID, ice.INT)
			}
		}},
	})
}
func messageCreate(m *ice.Message, zone, icons string) {
	kit.Value(m.Target().Configs[m.CommandKey()].Value, kit.Keys(mdb.HASH, zone, mdb.META), kit.Dict(
		mdb.TIME, m.Time(), mdb.TYPE, aaa.TECH, mdb.ZONE, zone, mdb.ICONS, icons,
	))
}
func messageInsert(m *ice.Message, zone string, arg ...string) {
	mdb.ZoneInsert(m, kit.Simple(mdb.ZONE, zone, arg, "direct", tcp.RECV, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.AVATAR, m.Option(ice.MSG_AVATAR)))
}
