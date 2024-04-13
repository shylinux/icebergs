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
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const PACKAGE = "package"

func init() {
	Index.MergeCommands(ice.Commands{
		PACKAGE: {Name: "package index auto", Help: "软件包", Actions: ice.MergeActions(ice.Actions{
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)
				m.Cmdy(nfs.DIR, path.Join(ice.USR_INSTALL, path.Base(m.Option(web.LINK))))
				mdb.HashModify(m, m.AppendSimple(mdb.TIME), mdb.TEXT, m.Append(nfs.PATH))
			}},
			cli.BUILD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)
				m.Cmdy(nfs.DIR, path.Join(_install_path(m, m.Option(web.LINK)), "_install/bin/nginx"))
				mdb.HashModify(m, m.AppendSimple(mdb.TIME), cli.CMD, m.Append(nfs.PATH))
			}},
			cli.START: {Name: "start port*=10000", Hand: func(m *ice.Message, arg ...string) {
				if cli.IsSuccess(m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)) {
					web.OpsCmd(m, tcp.PORT, mdb.CREATE, m.OptionSimple(tcp.PORT, mdb.NAME, mdb.TEXT, mdb.ICON, ctx.INDEX), web.SPACE, m.Option(ice.MSG_USERPOD),
						m.AppendSimple(cli.CMD, cli.PID))
					mdb.HashModify(m, m.AppendSimple(cli.PID), m.OptionSimple(tcp.PORT))
				}
			}},
			cli.STOP: {Hand: func(m *ice.Message, arg ...string) {
				if cli.IsSuccess(m.Cmdy(m.Option(ctx.INDEX), m.ActionKey(), arg)) {
					web.OpsCmd(m, tcp.PORT, mdb.MODIFY, m.OptionSimple(tcp.PORT), cli.PID, "")
					mdb.HashModify(m, cli.PID, "", tcp.PORT, "")
				}
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(ice.USR_INSTALL, path.Base(m.Option(web.LINK))))
				nfs.Trash(m, _install_path(m, m.Option(web.LINK)))
				mdb.HashModify(m, mdb.TEXT, "", cli.CMD, "")
			}},
		}, mdb.HashAction(mdb.SHORT, "index", mdb.FIELD, "time,index,type,name,text,icon,cmd,pid,port,link")), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				m.Cmdy(m.Option(ctx.INDEX), arg)
				return
			}
			mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
				button := []ice.Any{}
				switch value[mdb.TYPE] {
				case nfs.BINARY:
					if value[cli.PID] == "" {
						button = append(button, cli.START)
					} else {
						button = append(button, cli.STOP)
					}
					kit.If(!nfs.Exists(m, _install_path(m, value[mdb.LINK])), func() {
						button = append(button, web.INSTALL)
					}, func() {
						button = append(button, nfs.TRASH)
					})
				case nfs.SOURCE:
					if value[cli.PID] == "" {
						button = append(button, cli.START, cli.BUILD)
					} else {
						button = append(button, cli.STOP, cli.BUILD)
					}
					kit.If(!nfs.Exists(m, _install_path(m, value[mdb.LINK])), func() {
						button = append(button, web.DOWNLOAD)
					}, func() {
						button = append(button, nfs.TRASH)
					})
				}
				m.PushButton(button...)
			})
			web.PushPodCmd(m, "", arg...)
			kit.If(!m.IsWorker(), func() { m.RenameAppend(web.SPACE, ice.POD) })
			m.Action(html.FILTER)
		}},
	})
}
func PackageCreate(m *ice.Message, kind, name, text, icon, link string) {
	if m.Cmd(PACKAGE, m.PrefixKey()).Length() > 0 {
		return
	}
	m.Cmd(PACKAGE, mdb.CREATE, ctx.INDEX, m.PrefixKey(),
		mdb.TYPE, kind, mdb.NAME, name, mdb.TEXT, "",
		mdb.ICON, ctx.ResourceFile(m, kit.Select(name+".png", icon)), web.LINK, link,
	)
}
