package wx

import (
	"net/url"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _ide_args(m *ice.Message) (args []string) {
	kit.If(m.Option(tcp.WIFI), func(p string) { args = append(args, m.Cmd(tcp.WIFI, p).AppendSimple(tcp.SSID, aaa.PASSWORD)...) })
	return append(args, kit.Simple(kit.SplitQuery(m.Option(ctx.ARGS)), kit.Dict(m.OptionSimple(web.SPACE, ctx.INDEX, log.DEBUG)))...)
}
func _ide_args_cli(m *ice.Message) []string {
	return []string{"--project", kit.Path(mdb.Config(m, PROJECT)), "--compile-condition", kit.Format(kit.Dict(
		"pathName", m.Option(PAGES), "query", kit.JoinQuery(kit.Simple(_ide_args(m), web.SERVE, url.QueryEscape(web.UserHost(m)))...),
	))}
}
func _ide_args_qrcode(m *ice.Message, p string) []string {
	return []string{"--qr-format", nfs.IMAGE, "--qr-output", kit.Path(p)}
}
func _ide_autogen_utils(m *ice.Message) {
	to, from := ice.USR_PROGRAM, ice.USR_VOLCANOS+PUBLISH_CLIENT_MP
	nfs.DirDeepAll(m, from, "", func(value ice.Maps) {
		m.Cmd(nfs.COPY, to+value[nfs.PATH], path.Join(from, value[nfs.PATH]), ice.Maps{nfs.DIR_ROOT: ""})
	})
	to += "utils/"
	kit.For([]string{"const.js"}, func(from string) {
		m.Cmd(nfs.COPY, to+from, path.Join(ice.USR_VOLCANOS, from))
	})
	head, foot := nfs.TemplateText(m, "frame_head.js"), nfs.TemplateText(m, "frame_foot.js")
	kit.For([]string{"frame.js"}, func(from string) {
		m.Cmd(nfs.SAVE, to+from, head, lex.NL, m.Cmdx(nfs.CAT, path.Join(ice.USR_VOLCANOS, from)), lex.NL, foot)
	})
	head, foot = nfs.TemplateText(m, "lib_head.js"), nfs.TemplateText(m, "lib_foot.js")
	kit.For([]string{path.Join(ice.USR_VOLCANOS, ice.LIB), path.Join(nfs.SRC_TEMPLATE, web.CHAT_HEADER, aaa.LANGUAGE)}, func(from string) {
		nfs.DirDeepAll(m, from, "", func(value ice.Maps) {
			m.Cmd(nfs.SAVE, path.Join(to, ice.LIB, value[nfs.PATH]), head, lex.NL, m.Cmdx(nfs.CAT, value[nfs.PATH]), ice.Maps{nfs.DIR_ROOT: ""}, lex.NL, foot)
		})
	})
}
func _ide_autogen_pages(m *ice.Message) {
	to, list := ice.USR_PROGRAM, []string{}
	head, foot := nfs.TemplateText(m, "pages_head.js"), nfs.TemplateText(m, "pages_foot.js")
	head_wxml, foot_wxml, wxml := nfs.TemplateText(m, "pages_head.wxml"), nfs.TemplateText(m, "pages_foot.wxml"), nfs.TemplateText(m, "pages.wxml")
	m.Travel(func(_ *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if h, ok := cmd.Actions[PAGES]; ok {
			file := path.Join(path.Dir(strings.TrimPrefix(m.Resource(m.FileURI(cmd.FileLine())), nfs.REQUIRE)), kit.Select(key+".js", h.Name))
			if !nfs.Exists(m, file) {
				return
			}
			prefix := strings.ReplaceAll(s.Prefix(), nfs.PT, "-")
			list = append(list, path.Join(PAGES, prefix, kit.TrimExt(path.Base(file), nfs.JS)))
			kit.For([]string{WXML, WXSS, nfs.JS, nfs.JSON}, func(ext string) {
				if file = kit.ExtChange(file, ext); kit.Ext(file) == WXML {
					m.Cmd(nfs.SAVE, path.Join(to, PAGES, prefix, path.Base(file)), head_wxml, lex.NL, kit.Select(wxml, m.Cmdx(nfs.CAT, file)), lex.NL, foot_wxml)
				} else if kit.Ext(file) == nfs.JS {
					m.Cmd(nfs.SAVE, path.Join(to, PAGES, prefix, path.Base(file)), head, lex.NL, m.Cmdx(nfs.CAT, file), lex.NL, foot)
				} else {
					m.Cmd(nfs.COPY, path.Join(to, PAGES, prefix, path.Base(file)), file)
				}
			})
		}
	})
	app := kit.UnMarshal(m.Cmdx(nfs.CAT, to+APP_JSON))
	kit.Value(app, PAGES, kit.AddUniq(kit.Simple(kit.Value(app, PAGES)), list...))
	m.Cmd(nfs.SAVE, to+APP_JSON, kit.Formats(app))
}

const (
	PAGES_RIVER         = "pages/river/river"
	PAGES_ACTION        = "pages/action/action"
	PAGES_INSERT        = "pages/insert/insert"
	PUBLISH_CLIENT_MP   = "publish/client/mp"
	PROJECT_CONFIG_JSON = "project.config.json"
	APP_JSON            = "app.json"
	CONF_JS             = "conf.js"
)
const (
	ISLOGIN = "islogin"
	CURRENT = "current"
	PROJECT = "project"
	PAGES   = "pages"
	UTILS   = "utils"
)
const (
	AUTO_PREVIEW = "auto-preview"
	PREVIEW      = "preview"
	PUSH         = "push"
	DOC          = "doc"
	ENV          = "env"
)
const IDE = "ide"

func init() {
	Index.MergeCommands(ice.Commands{
		IDE: {Name: "ide hash auto", Help: "集成开发环境", Icon: "wxdev.png", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case PAGES:
					m.Push(arg[0], kit.Value(kit.UnMarshal(m.Cmdx(nfs.CAT, path.Join(mdb.Config(m, PROJECT), APP_JSON))), PAGES))
				case tcp.WIFI:
					m.Cmdy(tcp.WIFI).Cut(tcp.SSID)
				case web.WEIXIN:
					m.Cmd(web.SPACE).Table(func(value ice.Maps) {
						if value[mdb.TYPE] == web.WEIXIN {
							m.Push(arg[0], value[mdb.NAME])
							m.Push(aaa.IP, value[aaa.IP])
							m.Push(aaa.USERNICK, value[aaa.USERNICK])
							m.Push(aaa.USERNAME, value[aaa.USERNAME])
						}
					})
				}
			}},
			aaa.LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				p := nfs.TempName(m)
				m.GoSleep3s(func() { web.PushNoticeGrow(m, ice.Render(m, ice.RENDER_IMAGES, web.SHARE_LOCAL+p)) })
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p)).ProcessRefresh()
			}},
			code.AUTOGEN: {Name: "autogen projectname*='终端工具链' appid*='wxf4e5104d83476ed6' serve*='https://2021.shylinux.com'", Help: "生成", Icon: "bi bi-tools", Hand: func(m *ice.Message, arg ...string) {
				_ide_autogen_utils(m)
				_ide_autogen_pages(m)
				p := ice.USR_PROGRAM
				m.Cmd(nfs.SAVE, p+CONF_JS, `module.exports = `+kit.Formats(kit.Dict(m.OptionSimple(APPID, web.SERVE), nfs.MODULE, ice.Info.Make.Module, nfs.VERSION, ice.Info.Make.Versions())))
				m.Cmd(nfs.DEFS, p+PROJECT_CONFIG_JSON, kit.Formats(kit.Dict(m.OptionSimple(APPID, "projectname"))))
				IdeCli(m.Sleep3s(), cli.OPEN, "--project", kit.Path(mdb.Config(m, PROJECT, p))).ProcessInner()
			}},
			web.ADMIN: {Help: "后台", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen("https://mp.weixin.qq.com/")
			}},
			DOC: {Help: "文档", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen("https://developers.weixin.qq.com/miniprogram/dev/api/")
			}},
			cli.MAKE: {Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				m.Options(m.Cmd("", kit.Select(mdb.Config(m, CURRENT), arg, 0)).AppendSimple())
				kit.If(m.Option(cli.PWD), func(p string) {
					kit.If(p == kit.Path(ice.USR_VOLCANOS+PUBLISH_CLIENT_MP), func() { _ide_autogen_utils(m); _ide_autogen_pages(m) })
				})
				m.Cmd("", AUTO_PREVIEW)
			}},
			AUTO_PREVIEW: {Help: "预览", Icon: "bi bi-wechat", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				IdeCli(m, "", _ide_args_cli(m)).ProcessInner()
			}},
			PREVIEW: {Help: "体验", Icon: "bi bi-qr-code-scan", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				p := nfs.TempName(m)
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p)).EchoImages(web.SHARE_LOCAL + p).ProcessInner()
			}},
			PUSH: {Name: "push weixin", Help: "推送", Icon: "bi bi-arrow-up-right-square", Hand: func(m *ice.Message, arg ...string) {
				defer m.ProcessHold()
				defer web.ToastProcess(m)()
				m.Cmd(web.SPACE, m.Option(web.WEIXIN), lex.PARSE, m.Cmdx("", m.Option(mdb.HASH)))
			}},
		}, web.StatsAction("", "小程序场景数"), mdb.ExportHashAction(mdb.FIELD, "time,hash,name,text,pages,space,index,args,wifi", cli.DARWIN, "/Applications/wechatwebdevtools.app/Contents/MacOS/cli")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) && kit.Value(kit.UnMarshal(IdeCli(m.Spawn(), ISLOGIN).Append(cli.CMD_OUT)), aaa.LOGIN) != true {
				m.EchoInfoButton("请登录: ", aaa.LOGIN)
				return
			} else if !nfs.Exists(m, ice.USR_PROGRAM) {
				m.EchoInfoButton("请生成项目: ", code.AUTOGEN)
				return
			} else if mdb.HashSelect(m, arg...); tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
				m.PushAction(AUTO_PREVIEW, PREVIEW, PUSH, mdb.REMOVE).Action(aaa.LOGIN, code.AUTOGEN, mdb.CREATE, web.ADMIN, DOC)
			} else {
				m.PushAction(PUSH, mdb.REMOVE).Action(mdb.CREATE, web.ADMIN, DOC)
			}
			if len(arg) > 0 {
				m.Options(m.AppendSimple(web.SPACE, ctx.INDEX, ctx.ARGS, tcp.WIFI))
				p := kit.MergeURL2(kit.Select(web.UserHost(m), m.Option(web.SERVE)), path.Join(nfs.PS+m.Append(PAGES)), _ide_args(m))
				m.PushQRCode(cli.QRCODE, p).Push(web.LINK, p).Echo(p)
			}
			p := mdb.Config(m, CURRENT)
			m.Table(func(value ice.Maps) { m.Push(mdb.STATUS, kit.Select("", CURRENT, value[mdb.HASH] == p)) })
			m.StatusTimeCount(mdb.ConfigSimple(m, PROJECT))
		}},
	})
}

func IdeCli(m *ice.Message, action string, arg ...ice.Any) *ice.Message {
	defer web.ToastProcess(m)()
	switch runtime.GOOS {
	case cli.DARWIN:
		m.Cmdy(cli.SYSTEM, mdb.Config(m, runtime.GOOS), kit.Select(m.ActionKey(), action), arg)
	}
	return m
}
