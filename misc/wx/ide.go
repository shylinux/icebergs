package wx

import (
	"net/url"
	"path"
	"runtime"

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
	kit "shylinux.com/x/toolkits"
)

func _ide_args(m *ice.Message) (args []string) {
	args = append(kit.Split(m.Option(ctx.ARGS), "=&"), kit.Simple(kit.Dict(m.OptionSimple(web.SPACE, ctx.INDEX, log.DEBUG)))...)
	kit.If(m.Option(tcp.WIFI), func(p string) { args = append(args, m.Cmd(tcp.WIFI, p).AppendSimple(tcp.SSID, aaa.PASSWORD)...) })
	return
}
func _ide_args_cli(m *ice.Message) []string {
	return []string{"--project", kit.Path(mdb.Config(m, PROJECT)), "--compile-condition", kit.Format(kit.Dict(
		"pathName", m.Option(PAGES), "query", kit.JoinKV("=", "&", kit.Simple(kit.Dict(web.SERVE, url.QueryEscape(web.UserHost(m)), _ide_args(m)))...),
	))}
}
func _ide_args_qrcode(m *ice.Message, p string) []string {
	return []string{"--qr-format", nfs.IMAGE, "--qr-output", kit.Path(p)}
}

const (
	PAGES_ACTION = "pages/action/action"
)
const (
	PROJECT = "project"
	PAGES   = "pages"
	ENV     = "env"
)
const IDE = "ide"

func init() {
	const (
		AUTO_PREVIEW = "auto-preview"
		PREVIEW      = "preview"
		PUSH         = "push"
		DOC          = "doc"

		APP_JSON = "app.json"
		CURRENT  = "current"
		ISLOGIN  = "islogin"
	)
	Index.MergeCommands(ice.Commands{
		IDE: {Name: "ide hash auto", Help: "集成开发环境", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			ice.APP: {Help: "应用", Hand: func(m *ice.Message, arg ...string) {
				IdeCli(m, cli.OPEN, "--project", kit.Path(mdb.Config(m, PROJECT)))
			}},
			aaa.LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				p := nfs.TempName(m)
				m.Go(func() { web.PushNoticeGrow(m.Sleep("1s"), ice.Render(m, ice.RENDER_IMAGES, web.SHARE_LOCAL+p)) })
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p)).ProcessRefresh()
			}},
			web.ADMIN: {Help: "后台", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen("https://mp.weixin.qq.com/")
			}},
			DOC: {Help: "文档", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen("https://developers.weixin.qq.com/miniprogram/dev/api/")
			}},
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
			cli.MAKE: {Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				msg := m.Cmd("", kit.Select(mdb.Config(m, CURRENT), arg, 0))
				m.Options(msg.AppendSimple()).Cmd("", AUTO_PREVIEW)
			}},
			AUTO_PREVIEW: {Help: "预览", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				IdeCli(m, "", _ide_args_cli(m)).ProcessInner()
			}},
			PREVIEW: {Help: "体验", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, CURRENT, p) })
				p := nfs.TempName(m)
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p))
				m.EchoImages(web.SHARE_LOCAL + p).ProcessInner()
			}},
			PUSH: {Name: "push weixin", Help: "推送", Hand: func(m *ice.Message, arg ...string) {
				defer m.ProcessHold()
				defer web.ToastProcess(m)()
				m.Cmd(web.SPACE, m.Option(web.WEIXIN), lex.PARSE, m.Cmdx("", m.Option(mdb.HASH)))
			}},
		}, mdb.ExportHashAction(
			mdb.FIELD, "time,hash,name,pages,space,index,args,wifi",
			cli.DARWIN, "/Applications/wechatwebdevtools.app/Contents/MacOS/cli",
			PROJECT, "usr/volcanos/publish/client/mp/",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 && tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) && kit.Value(kit.UnMarshal(IdeCli(m.Spawn(), ISLOGIN).Append(cli.CMD_OUT)), aaa.LOGIN) != true {
				m.EchoInfoButton("请登录: ", aaa.LOGIN)
				return
			}
			if mdb.HashSelect(m, arg...); tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
				m.PushAction(AUTO_PREVIEW, PREVIEW, PUSH, mdb.REMOVE).Action(mdb.CREATE, ice.APP, aaa.LOGIN, web.ADMIN, DOC)
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
