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
		"pathName", m.Option(nfs.PATH), "query", kit.JoinKV("=", "&", kit.Simple(kit.Dict(web.SERVE, url.QueryEscape(web.UserHost(m)), _ide_args(m)))...),
	))}
}
func _ide_args_qrcode(m *ice.Message, p string) []string {
	return []string{"--qr-format", nfs.IMAGE, "--qr-output", kit.Path(p)}
}

const (
	PROJECT = "project"
)
const IDE = "ide"

func init() {
	const (
		PREVIEW      = "preview"
		AUTO_PREVIEW = "auto-preview"
	)
	Index.MergeCommands(ice.Commands{
		IDE: {Name: "ide hash auto", Help: "集成开发环境", Actions: ice.MergeActions(ice.Actions{
			ice.APP: {Help: "应用", Hand: func(m *ice.Message, arg ...string) {
				IdeCli(m, cli.OPEN, "--project", kit.Path(mdb.Config(m, PROJECT)))
			}},
			aaa.LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				p := nfs.TempName(m)
				m.Go(func() { web.PushNoticeGrow(m.Sleep("1s"), ice.Render(m, ice.RENDER_IMAGES, web.SHARE_LOCAL+p)) })
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p)).ProcessRefresh()
			}},
			cli.MAKE: {Help: "调试", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.HASH), func(p string) { mdb.Config(m, cli.MAKE, p) })
				msg := m.Cmd("", kit.Select(mdb.Config(m, cli.MAKE), arg, 0))
				m.Options(msg.AppendSimple()).Cmd("", AUTO_PREVIEW)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case nfs.PATH:
					m.Push(arg[0], kit.Value(kit.UnMarshal(m.Cmdx(nfs.CAT, path.Join(mdb.Config(m, PROJECT), "app.json"))), "pages"))
				case tcp.WIFI:
					m.Cmdy(tcp.WIFI).Cut(tcp.SSID)
				}
			}},
			AUTO_PREVIEW: {Help: "预览", Hand: func(m *ice.Message, arg ...string) { IdeCli(m, "", _ide_args_cli(m)).ProcessInner() }},
			PREVIEW: {Help: "体验", Hand: func(m *ice.Message, arg ...string) {
				p := nfs.TempName(m)
				IdeCli(m, "", _ide_args_cli(m), _ide_args_qrcode(m, p))
				m.EchoImages(web.SHARE_LOCAL + p).ProcessInner()
			}},
		}, mdb.ExportHashAction(mdb.FIELD, "time,hash,name,path,space,index,args,wifi", cli.DARWIN, "/Applications/wechatwebdevtools.app/Contents/MacOS/cli", PROJECT, "usr/volcanos/publish/client/mp/")), Hand: func(m *ice.Message, arg ...string) {
			if kit.Value(kit.UnMarshal(IdeCli(m.Spawn(), "islogin").Append(cli.CMD_OUT)), aaa.LOGIN) != true {
				m.EchoInfoButton("请登录: ", aaa.LOGIN)
			} else if mdb.HashSelect(m, arg...).PushAction(AUTO_PREVIEW, PREVIEW, cli.MAKE, mdb.REMOVE).Action(mdb.CREATE, ice.APP); len(arg) > 0 {
				m.Options(m.AppendSimple(web.SPACE, ctx.INDEX, ctx.ARGS, tcp.WIFI))
				p := kit.MergeURL2(web.UserHost(m), path.Join(nfs.PS+m.Append(nfs.PATH)), _ide_args(m))
				m.EchoQRCode(p).Echo(lex.NL).EchoAnchor(p)
			}
			m.StatusTimeCount(kit.Dict(mdb.ConfigSimple(m, cli.MAKE)))
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
