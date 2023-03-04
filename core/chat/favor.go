package chat

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _favor_is_image(m *ice.Message, name, mime string) bool {
	return strings.HasPrefix(mime, "image/") || kit.ExtIsImage(name)
}
func _favor_is_video(m *ice.Message, name, mime string) bool {
	return strings.HasPrefix(mime, "video/") || kit.ExtIsVideo(name)
}
func _favor_is_audio(m *ice.Message, name, mime string) bool {
	return strings.HasPrefix(mime, "audio/")
}

const (
	FAVOR_INPUTS = "favor.inputs"
	FAVOR_TABLES = "favor.tables"
	FAVOR_ACTION = "favor.action"
)
const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor hash auto create getClipboardData getLocation scanQRCode record1 record2 upload", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], web.LINK, nfs.FILE, mdb.TEXT, ctx.INDEX, ssh.SHELL)
				case mdb.NAME:
					switch m.Option(mdb.TYPE) {
					case ctx.INDEX:
						m.Copy(m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, arg[1:], ice.OptionFields(ctx.INDEX)).RenameAppend(ctx.INDEX, arg[0]))
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
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(mdb.TEXT), ice.HTTP) {
					m.OptionDefault(mdb.TYPE, mdb.LINK, mdb.NAME, kit.ParseURL(m.Option(mdb.TEXT)).Host)
				}
				mdb.HashCreate(m, m.OptionSimple())
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || arg[0] == mdb.FOREACH && arg[1] == "" {
					m.Cmd("", ice.Maps{ice.MSG_FIELDS: ""}, func(values ice.Maps) { m.PushSearch(values) })
				}
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT))
			}},
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessOpen(m, web.MergeURL2(m, web.SHARE_LOCAL+m.Option(mdb.TEXT), "filename", m.Option(mdb.NAME)))
			}},
			web.DISPLAY: {Help: "预览", Hand: func(m *ice.Message, arg ...string) {
				if link := web.SHARE_LOCAL + m.Option(mdb.TEXT); _favor_is_image(m, m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
					m.EchoImages(link)
				} else if _favor_is_video(m, m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
					m.EchoVideos(link)
				} else {
					m.Echo("<audio src=%s autoplay controls/>", link)
				}
				m.ProcessInner()
			}},
			ctx.INDEX: {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH))
				ls := kit.Split(msg.Option(mdb.TEXT))
				ctx.ProcessField(m, ls[0], ls[1:], arg...)
			}},
			"vimer": {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				args := []string{}
				if len(arg) == 0 || arg[0] != ice.RUN {
					args = nfs.SplitPath(m, m.Option(mdb.TEXT))
				}
				ctx.ProcessField(m, web.CODE_VIMER, args, arg...)
			}},
			"xterm": {Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, web.CODE_XTERM, []string{m.Option(mdb.TEXT)}, arg...)
			}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TYPE, mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH)).Append(mdb.TYPE))
				ctx.Run(m, arg...)
			}},
		}, mdb.HashAction(), ctx.CmdAction(), KeyboardAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Option(mdb.TYPE, mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH)).Append(mdb.TYPE))
				gdb.Event(m, FAVOR_ACTION, arg)
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				text := m.Append(mdb.TEXT)
				if strings.HasPrefix(m.Append(mdb.TEXT), ice.VAR_FILE) {
					text = web.SHARE_LOCAL + m.Append(mdb.TEXT)
					if m.PushDownload(mdb.LINK, m.Append(mdb.NAME), text); len(arg) > 0 && _favor_is_image(m, m.Append(mdb.NAME), m.Append(mdb.TYPE)) {
						m.PushImages(web.DISPLAY, text)
					} else if _favor_is_video(m, m.Append(mdb.NAME), m.Append(mdb.TYPE)) {
						m.PushVideos(web.DISPLAY, text)
					}
					text = tcp.PublishLocalhost(m, web.MergeLink(m, text))
				}
				m.PushScript(nfs.SCRIPT, text)
				m.PushQRCode(cli.QRCODE, text)
			}
			m.Tables(func(value ice.Maps) {
				if msg := gdb.Event(m.Spawn(), FAVOR_TABLES, mdb.TYPE, value[mdb.TYPE]); msg.Append(ctx.ACTION) != "" {
					m.PushButton(msg.Append(ctx.ACTION))
					return
				}
				switch value[mdb.TYPE] {
				case ssh.SHELL:
					m.PushButton("xterm", mdb.REMOVE)
				case ctx.INDEX:
					m.PushButton(ctx.INDEX, mdb.REMOVE)
				case nfs.FILE:
					m.PushButton("vimer", mdb.REMOVE)
				default:
					if strings.HasPrefix(value[mdb.TEXT], ice.VAR_FILE) {
						if _favor_is_image(m, value[mdb.NAME], value[mdb.TYPE]) || _favor_is_video(m, value[mdb.NAME], value[mdb.TYPE]) || _favor_is_audio(m, value[mdb.NAME], value[mdb.TYPE]) {
							m.PushButton(web.DISPLAY, web.DOWNLOAD, mdb.REMOVE)
						} else {
							m.PushButton(web.DOWNLOAD, mdb.REMOVE)
						}
					} else {
						m.PushButton(mdb.REMOVE)
					}
				}
			})
		}},
	})
}

func FavorAction() ice.Actions { return gdb.EventAction(FAVOR_INPUTS, FAVOR_TABLES, FAVOR_ACTION) }
