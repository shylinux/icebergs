package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
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
		FAVOR: {Help: "收藏夹", Icon: "favor.png", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					m.Cmds("", func(value ice.Maps) {
						if arg[1] == "" || arg[1] == value[mdb.TYPE] || strings.Contains(value[mdb.TEXT], arg[1]) {
							m.PushSearch(value)
						}
					})
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], mdb.TEXT, ctx.INDEX, cli.OPENS)
				case mdb.NAME:
					switch m.Option(mdb.TYPE) {
					case ctx.INDEX:
						m.Copy(mdb.HashInputs(m.Spawn(), ctx.INDEX).CutTo(ctx.INDEX, arg[0]))
						return
					}
				case mdb.TEXT:
					switch m.Option(mdb.TYPE) {
					case ctx.INDEX:
						m.Option(ctx.INDEX, m.Option(mdb.NAME))
						m.Copy(mdb.HashInputs(m.Spawn(), ctx.ARGS).CutTo(ctx.ARGS, arg[0]))
						return
					}
				}
				gdb.Event(m, "", arg)
			}},
			"getClipboardData": {Name: "favor create", Help: "粘贴"},
			"getLocation":      {Name: "favor create", Help: "定位"},
			"scanQRCode":       {Name: "favor create", Help: "扫码"},
			"record1":          {Name: "favor upload", Help: "截图"},
			"record2":          {Name: "favor upload", Help: "录屏"},
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(mdb.TEXT), ice.HTTP) {
					m.OptionDefault(mdb.TYPE, mdb.LINK, mdb.NAME, kit.ParseURL(m.Option(mdb.TEXT)).Host)
				}
				mdb.HashCreate(m, m.OptionSimple())
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT))
			}},
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(web.MergeURL2(m, web.SHARE_LOCAL+m.Option(mdb.TEXT), nfs.FILENAME, m.Option(mdb.NAME)))
			}},
			ctx.INDEX: {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if kit.HasPrefixList(arg, ctx.RUN) {
					msg := mdb.HashSelects(m.Spawn(), arg[1])
					ctx.ProcessField(m, msg.Append(mdb.NAME), kit.Split(msg.Append(mdb.TEXT)), kit.Simple(ctx.RUN, arg[2:])...)
				} else {
					msg := mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH))
					ctx.ProcessField(m, msg.Append(mdb.NAME), kit.Split(msg.Append(mdb.TEXT)), arg...)
					m.Option(ice.FIELD_PREFIX, ctx.ACTION, m.ActionKey(), ctx.RUN, m.Option(mdb.HASH))
				}
			}},
			cli.OPENS: {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, m.Option(mdb.TEXT)) }},
		}, FavorAction(), mdb.ExportHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				if m.Option(ice.MSG_INDEX) == m.PrefixKey() {
					m.Option(mdb.TYPE, mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH)).Append(mdb.TYPE))
					gdb.Event(m, FAVOR_ACTION, arg)
				} else if aaa.Right(m, m.Option(ice.MSG_INDEX), arg[3:]) {
					m.Cmdy(m.Option(ice.MSG_INDEX), arg[3:])
				}
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				text := m.Append(mdb.TEXT)
				m.PushQRCode(cli.QRCODE, text)
				m.PushScript(text)
			}
			if len(arg) == 0 {
				if m.IsMobileUA() {
					m.Action(mdb.CREATE, web.UPLOAD, "getClipboardData", "getLocation", "scanQRCode")
				} else {
					m.Action(mdb.CREATE, web.UPLOAD, "getClipboardData", "record1", "record2")
				}
			}
			m.Table(func(value ice.Maps) {
				delete(value, ctx.ACTION)
				if msg := gdb.Event(m.Spawn(value), FAVOR_TABLES, mdb.TYPE, value[mdb.TYPE]); msg.Append(ctx.ACTION) != "" {
					m.PushButton(msg.Append(ctx.ACTION), mdb.REMOVE)
					return
				}
				switch value[mdb.TYPE] {
				case ctx.INDEX:
					m.PushButton(ctx.INDEX, mdb.REMOVE)
				case cli.OPENS:
					m.PushButton(cli.OPENS, mdb.REMOVE)
				default:
					m.PushButton(mdb.REMOVE)
				}
			})
		}},
	})
}

func FavorAction() ice.Actions { return gdb.EventsAction(FAVOR_INPUTS, FAVOR_TABLES, FAVOR_ACTION) }
