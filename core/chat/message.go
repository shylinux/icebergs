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
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const MESSAGE = "message"

func init() {
	Index.MergeCommands(ice.Commands{
		MESSAGE: {Name: "message refresh", Help: "聊天", Icon: "Messages.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				MessageCreate(m, aaa.APPLY, html.ICONS_MAIL)
				MessageCreate(m, web.DREAM, html.ICONS_DREAM)
				MessageCreate(m, cli.SYSTEM, html.ICONS_SETTINGS)
				web.MessageInsertJSON(m, cli.SYSTEM, cli.BOOTINFO, m.Cmdx(cli.RUNTIME), ctx.ARGS, m.Cmdx(cli.RUNTIME, "boot.time"))
			}},
			mdb.CREATE: {Name: "create type*=tech,void title icons target zone", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(web.TARGET), "from.") {
					m.Option(web.TARGET, strings.Replace(m.Option(web.TARGET), "from", m.Option(ice.FROM_SPACE), 1))
				}
				if m.OptionDefault(mdb.ZONE, m.Option(web.TARGET)) == "" {
					m.Option(mdb.ZONE, kit.Hashs(mdb.UNIQ))
				}
				mdb.ZoneCreate(m, kit.Simple(arg, web.TARGET, m.Option(web.TARGET), mdb.ZONE, m.Option(mdb.ZONE)))
			}},
			mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, kit.Simple(arg[0], tcp.DIRECT, tcp.SEND, arg[1:], aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.AVATAR, m.Option(ice.MSG_AVATAR)))
				mdb.HashSelectUpdate(m, arg[0], func(value ice.Map) { kit.Value(value, mdb.TIME, m.Time()) })
				web.StreamPushRefresh(m)
			}},
			tcp.SEND: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.INSERT, arg, tcp.DIRECT, tcp.SEND)
				kit.If(mdb.HashSelectField(m, arg[0], web.TARGET), func(p string) { m.Cmd(web.SPACE, p, MESSAGE, tcp.RECV, arg[1:]) })
			}},
			tcp.RECV: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.INSERT, m.Option(ice.FROM_SPACE), arg, tcp.DIRECT, tcp.RECV)
				mdb.HashSelectUpdate(m, m.Option(ice.FROM_SPACE), func(value ice.Map) { kit.Value(value, web.TARGET, m.Option(ice.FROM_SPACE)) })
			}},
			cli.CLEAR: {Hand: func(m *ice.Message, arg ...string) {}},
			web.OPEN:  {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(m.MergePod(m.Option(web.TARGET))) }},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if !m.IsCliUA() {
					MessageInsertPlug(m, web.DREAM, "", "", web.DREAM, m.Option(mdb.NAME))
				}
			}},
			web.DREAM_REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, web.DREAM, "", "", web.DREAM, m.Option(mdb.NAME))
			}},
			web.SPACE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, aaa.APPLY, "", "", web.CHAT_GRANT, m.Option(mdb.NAME))
			}},
			aaa.OFFER_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, aaa.APPLY, "", "", aaa.OFFER, m.Option(mdb.HASH))
			}},
			aaa.OFFER_ACCEPT: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, aaa.APPLY, "", "", aaa.OFFER, m.Option(mdb.HASH))
			}},
			aaa.USER_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, aaa.APPLY, "", "", aaa.USER, m.Option(aaa.USERNAME))
			}},
			aaa.USER_REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				MessageInsertPlug(m, aaa.APPLY, "", "", aaa.USER, m.Option(aaa.USERNAME))
			}},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(tcp.DIRECT) == tcp.RECV {
					m.Cmdy(web.Space(m, m.Option(web.TARGET)), ctx.COMMAND, arg[0]).ProcessField(ctx.ACTION, ctx.RUN, m.Option(web.TARGET), arg[0])
				} else {
					m.Cmdy(ctx.COMMAND, arg[0]).ProcessField(ctx.ACTION, ctx.RUN, "", arg[0])
				}
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(web.Space(m, arg[0]), arg[1], arg[2:]) }},
		}, web.DreamTablesAction(), web.DreamAction(), aaa.OfferAction(), mdb.ZoneAction(
			mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,hash,type,zone,icons,title,count,target",
			mdb.FIELDS, "time,id,type,name,text,space,index,args,style,display,username,usernick,avatar,direct",
			web.ONLINE, ice.TRUE,
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.ZoneSelect(m.Display("").Spawn(), arg...).Table(func(value ice.Maps) {
					if !kit.IsIn(m.Option(ice.MSG_USERROLE), value[mdb.TYPE], aaa.TECH, aaa.ROOT) {
						return
					}
					m.PushRecord(value, mdb.Config(m, mdb.FIELD))
					list := []ice.Any{}
					if value[web.TARGET] != "" {
						list = append(list, web.OPEN)
					}
					if aaa.IsTechOrRoot(m) {
						list = append(list, cli.CLEAR, mdb.REMOVE)
					}
					m.PushButton(list...)
				})
				m.Sort(mdb.TIME, ice.STR_R)
				ctx.Toolkit(m)
			} else {
				if msg := mdb.ZoneSelects(m.Spawn(), arg[0]); !kit.IsIn(m.Option(ice.MSG_USERROLE), msg.Append(mdb.TYPE), aaa.TECH, aaa.ROOT) {
					return
				}
				mdb.ZoneSelect(m, arg...).Sort(mdb.ID, ice.INT)
			}
		}},
	})
}
func MessageCreate(m *ice.Message, zone, icons string) {
	if _, ok := m.CmdMap(MESSAGE, mdb.ZONE)[zone]; !ok {
		m.Cmd(MESSAGE, mdb.CREATE, mdb.TYPE, aaa.TECH, mdb.ICONS, icons, mdb.ZONE, zone)
	}
}
func MessageInsert(m *ice.Message, zone string, arg ...string) {
	if ice.Info.Important {
		m.Cmd(MESSAGE, mdb.INSERT, zone, tcp.DIRECT, tcp.RECV, arg)
	}
}
func MessageInsertPlug(m *ice.Message, zone, name, text, index, args string, arg ...string) {
	kit.If(text == "", func() {
		msg := m.Cmds(index, args, ice.Maps{ice.MSG_USERUA: html.Mozilla})
		kit.If(msg.Option(ice.MSG_STATUS) == "", func() { msg.StatusTimeCount() })
		text = msg.FormatMeta()
	})
	MessageInsert(m, zone, kit.Simple(mdb.TYPE, html.PLUG, mdb.NAME, kit.Select(m.ActionKey(), name), mdb.TEXT, text, ctx.INDEX, kit.Select(m.ShortKey(), index), ctx.ARGS, args, arg)...)
}
