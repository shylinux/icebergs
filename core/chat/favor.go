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
	"shylinux.com/x/icebergs/base/web/html"
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
		FAVOR: {Name: "favor hash auto", Help: "收藏夹", Icon: "favor.png", Actions: ice.MergeActions(ice.Actions{
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
			html.GetLocation:      {Name: "favor create", Help: "定位", Icon: "bi bi-geo-alt"},
			html.GetClipboardData: {Name: "favor create", Help: "粘贴", Icon: "bi bi-copy"},
			html.ScanQRCode:       {Name: "favor create", Help: "扫码", Icon: "bi bi-qr-code-scan"},
			html.Record1:          {Name: "favor upload", Help: "截图"},
			html.Record2:          {Name: "favor upload", Help: "录屏"},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(mdb.TEXT), ice.HTTP) {
					m.OptionDefault(mdb.TYPE, mdb.LINK, mdb.NAME, kit.ParseURL(m.Option(mdb.TEXT)).Host)
				}
				mdb.HashCreate(m, m.OptionSimple())
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(!web.PodCmd(m, web.SPACE, kit.Simple(ctx.ACTION, m.ActionKey(), arg)...), func() { mdb.HashRemove(m, arg) })
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT))
			}},
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergeLink(web.SHARE_LOCAL+m.Option(mdb.TEXT), nfs.FILENAME, m.Option(mdb.NAME)))
			}},
			web.PREVIEW: {Hand: FavorPreview},
			cli.OPENS:   {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, m.Option(mdb.TEXT)) }},
			web.PAGES:   {Name: "favor.js"},
		}, FavorAction(), mdb.ExportHashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				if m.Option(ice.MSG_INDEX) == m.PrefixKey() {
					m.Option(mdb.TYPE, mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH)).Append(mdb.TYPE))
					gdb.Event(m, FAVOR_ACTION, arg)
				} else if aaa.Right(m, m.Option(ice.MSG_INDEX), arg[3:]) {
					m.Cmdy(m.Option(ice.MSG_INDEX), arg[3:])
				}
				return
			} else if mdb.HashSelect(m, arg...); len(arg) == 0 {
				defer web.PushPodCmd(m, "", arg...)
				if m.SortStrR(mdb.TIME); m.IsMobileUA() {
					m.Action(mdb.CREATE, web.UPLOAD, html.GetClipboardData, html.GetLocation, html.ScanQRCode)
				} else {
					m.Action(mdb.CREATE, web.UPLOAD, html.GetClipboardData, html.Record1, html.Record2)
				}
			} else {
				m.PushQRCode(cli.QRCODE, m.Append(mdb.TEXT))
				m.PushScript(m.Append(mdb.TEXT))
			}
			m.Table(func(value ice.Maps) {
				delete(value, ctx.ACTION)
				if msg := gdb.Event(m.Spawn(value), FAVOR_TABLES, mdb.TYPE, value[mdb.TYPE]); msg.Append(ctx.ACTION) != "" {
					m.PushButton(msg.Append(ctx.ACTION), mdb.REMOVE)
					return
				}
				switch value[mdb.TYPE] {
				case cli.OPENS:
					m.PushButton(cli.OPENS, mdb.REMOVE)
				default:
					m.PushButton(web.PREVIEW, mdb.REMOVE)
				}
			})
		}},
	})
}

func FavorAction() ice.Actions {
	return gdb.EventsAction(FAVOR_INPUTS, FAVOR_TABLES, FAVOR_ACTION)
}
func FavorPreview(m *ice.Message, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) {
		web.ProcessPodCmd(m, "", "", nil, arg...)
	} else {
		msg := m
		if m.Option(web.SPACE) == "" {
			msg = mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH))
		} else {
			msg = m.Cmd(web.SPACE, m.Option(web.SPACE), m.PrefixKey(), m.Option(mdb.HASH))
		}
		index, args := msg.Append(mdb.TYPE), kit.Split(msg.Append(mdb.TEXT))
		switch msg.Append(mdb.TYPE) {
		case ctx.INDEX:
			index = msg.Append(mdb.NAME)
		case nfs.SHY:
			index = web.WORD
		}
		web.ProcessPodCmd(m, m.Option(web.SPACE), index, args, arg...)
	}
}
