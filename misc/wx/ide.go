package wx

import (
	"net/url"
	"path"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _ide_args(m *ice.Message) []string {
	return []string{"--project", kit.Path(mdb.Config(m, PROJECT)), "--compile-condition", kit.Format(kit.Dict(
		"pathName", m.Option(nfs.PATH), "query", kit.JoinKV("=", "&", append(kit.Split(m.Option(ctx.ARGS), "=&"), kit.Simple(kit.Dict(
			web.SERVE, url.QueryEscape(web.UserHost(m)), m.OptionSimple(web.SPACE, ctx.INDEX, log.DEBUG),
		))...)...),
	))}
}

const (
	PROJECT = "project"
)
const IDE = "ide"

func init() {
	const (
		PREVIEW      = "preview"
		AUTO_PREVIEW = "auto-preview"
		CMDS_PREVIEW = "cmds-preview"
	)
	Index.MergeCommands(ice.Commands{
		IDE: {Name: "ide hash auto app login", Help: "集成开发环境", Actions: ice.MergeActions(ice.Actions{
			ice.APP: {Help: "应用", Hand: func(m *ice.Message, arg ...string) {
				IdeCli(m, "open", "--project", kit.Path(mdb.Config(m, PROJECT)))
			}},
			aaa.LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				p := m.Cmdx(nfs.SAVE, path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ)), "")
				m.Go(func() { IdeCli(m, "", "--qr-format", nfs.IMAGE, "--qr-output", kit.Path(p), _ide_args(m)) }).Sleep("1s")
				m.EchoImages(web.SHARE_LOCAL + p).ProcessInner()
			}},
			CMDS_PREVIEW: {Name: "cmds-preview space index*=web.dream args='debug=true'", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", AUTO_PREVIEW)
			}},
			AUTO_PREVIEW: {Help: "自动", Hand: func(m *ice.Message, arg ...string) {
				IdeCli(m, "", _ide_args(m)).ProcessInner()
			}},
			PREVIEW: {Help: "预览", Hand: func(m *ice.Message, arg ...string) {
				p := m.Cmdx(nfs.SAVE, path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ)), "")
				IdeCli(m, "", "--qr-format", nfs.IMAGE, "--qr-output", kit.Path(p), _ide_args(m))
				m.EchoImages(web.SHARE_LOCAL + p).ProcessInner()
			}},
		}, mdb.ExportHashAction(mdb.FIELD, "time,hash,name,path,args", cli.DARWIN, "/Applications/wechatwebdevtools.app/Contents/MacOS/cli", PROJECT, "usr/volcanos/publish/client/mp/")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(CMDS_PREVIEW, AUTO_PREVIEW, PREVIEW, mdb.REMOVE)
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
