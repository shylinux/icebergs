package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	FAVOR_INPUTS = "favor.inputs"
	FAVOR_TABLES = "favor.tables"
	FAVOR_ACTION = "favor.action"
)
const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor hash auto create getClipboardData getLocation scanQRCode upload", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], mdb.TEXT, ctx.INDEX)
				case mdb.NAME:
					switch m.Option(mdb.TYPE) {
					case ctx.INDEX:
						m.Copy(m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, arg[1:], ice.OptionFields(ctx.INDEX)).RenameAppend(ctx.INDEX, arg[0]))
					}
				}
				gdb.Event(m, "", arg)
			}},
			"getClipboardData": {Name: "favor create", Help: "粘贴"},
			"getLocation":      {Name: "favor create", Help: "定位"},
			"scanQRCode":       {Name: "favor create", Help: "扫码"},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.TYPE, mdb.LINK, mdb.NAME, kit.ParseURL(m.Option(mdb.TEXT)).Host)
				mdb.HashCreate(m, m.OptionSimple())
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				web.Upload(m).Cmd("", mdb.CREATE, m.AppendSimple(mdb.TYPE, mdb.NAME, mdb.TEXT))
			}},
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessOpen(m, web.MergeURL2(m, web.SHARE_LOCAL+m.Option(mdb.TEXT), "filename", m.Option(mdb.NAME)))
			}},
			web.DISPLAY: {Help: "预览", Hand: func(m *ice.Message, arg ...string) {
				m.EchoImages(web.SHARE_LOCAL + m.Option(mdb.TEXT)).ProcessInner()
			}},
			ctx.INDEX: {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.Cmd("", m.Option(mdb.HASH)).Append(mdb.NAME), kit.Simple(kit.UnMarshal(m.Option(mdb.TEXT))), arg...)
			}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TYPE, m.Cmd("", m.Option(mdb.HASH)).Append(mdb.TYPE))
				ctx.Run(m, arg...)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Option(mdb.TYPE, m.Cmd("", m.Option(mdb.HASH)).Append(mdb.TYPE))
				gdb.Event(m, FAVOR_ACTION, arg)
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.Tables(func(value ice.Maps) {
					if msg := gdb.Event(m.Spawn(), FAVOR_TABLES, mdb.TYPE, value[mdb.TYPE]); msg.Append(ctx.ACTION) != "" {
						m.PushButton(msg.Append(ctx.ACTION))
						return
					}
					switch value[mdb.TYPE] {
					case ctx.INDEX:
						m.PushButton(ctx.INDEX, mdb.REMOVE)
					default:
						if strings.HasPrefix(value[mdb.TEXT], ice.VAR_FILE) {
							if kit.ExtIsImage(value[mdb.NAME]) {
								m.PushButton(web.DISPLAY, web.DOWNLOAD, mdb.REMOVE)
							} else {
								m.PushButton(web.DOWNLOAD, mdb.REMOVE)
							}
						} else {
							m.PushButton(mdb.REMOVE)
						}
					}
				})
			} else {
				if strings.HasPrefix(m.Append(mdb.TEXT), ice.VAR_FILE) {
					link := web.SHARE_LOCAL + m.Append(mdb.TEXT)
					if m.PushDownload(mdb.LINK, m.Append(mdb.NAME), link); len(arg) > 0 && kit.ExtIsImage(m.Append(mdb.NAME)) {
						m.PushImages(web.DISPLAY, link)
					}
				}
				m.PushScript(ssh.SCRIPT, m.Append(mdb.TEXT))
				m.PushQRCode(cli.QRCODE, m.Append(mdb.TEXT))
				m.PushAction(mdb.REMOVE)
			}
		}},
	})
}

func FavorAction() ice.Actions { return gdb.EventAction(FAVOR_INPUTS, FAVOR_TABLES, FAVOR_ACTION) }
