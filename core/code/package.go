package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const PACKAGE = "package"

func init() {
	Index.MergeCommands(ice.Commands{
		PACKAGE: {Name: "package index auto", Help: "软件包", Actions: ice.MergeActions(ice.Actions{
			cli.START: {Name: "start port*=10000", Hand: func(m *ice.Message, arg ...string) {
				if cli.IsSuccess(m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)) {
					web.OpsCmd(m, tcp.PORT, mdb.CREATE, m.OptionSimple(tcp.PORT, mdb.NAME, mdb.TEXT, mdb.ICON, ctx.INDEX), web.SPACE, m.Option(ice.MSG_USERPOD), m.AppendSimple(cli.CMD, cli.PID))
				}
			}},
			cli.BUILD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)
				m.Cmdy(nfs.DIR, path.Join(_install_path(m, m.Option(web.LINK)), "_install/bin/nginx"))
				mdb.HashModify(m, mdb.TIME, m.Append(mdb.TIME), cli.CMD, m.Append(nfs.PATH))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(ice.USR_INSTALL, path.Base(m.Option(web.LINK))))
				nfs.Trash(m, _install_path(m, m.Option(web.LINK)))
				mdb.HashModify(m, cli.CMD, "")
			}},
		}, mdb.HashAction(mdb.SHORT, "index", mdb.FIELD, "time,index,type,name,text,icon,cmd,link")), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				m.Cmdy(m.Option(ctx.INDEX), arg)
				return
			}
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				button := []ice.Any{}
				switch value[mdb.TYPE] {
				case nfs.BINARY:
					kit.If(!nfs.Exists(m, _install_path(m, value[mdb.LINK])), func() {
						button = append(button, web.INSTALL)
					}, func() {
						button = append(button, nfs.TRASH)
					})
				case nfs.SOURCE:
					button = append(button, cli.START, cli.BUILD)
					kit.If(!nfs.Exists(m, _install_path(m, value[mdb.LINK])), func() {
						button = append(button, web.DOWNLOAD)
					}, func() {
						button = append(button, nfs.TRASH)
					})
				}
				m.PushButton(button...)
			})
			web.PushPodCmd(m, "", arg...)
			kit.If(m.Option(ice.MSG_USERPOD) == "", func() { m.RenameAppend(web.SPACE, ice.POD) })
		}},
	})
}
func PackageCreate(m *ice.Message, kind, name, text, icon, link string) {
	if m.Cmd(PACKAGE, m.PrefixKey()).Length() > 0 {
		return
	}
	m.Cmd(PACKAGE, mdb.CREATE, ctx.INDEX, m.PrefixKey(),
		mdb.TYPE, kind, mdb.NAME, name, mdb.TEXT, kit.Select(path.Base(link), text),
		mdb.ICON, icon, web.LINK, link,
	)
}
