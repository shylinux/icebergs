package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	_DOWNLOAD = "_download"
)

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(web.DOWNLOAD): {Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Cmdy(FAVOR, _DOWNLOAD).TableEcho()
			} else {
				m.Cmdy(web.CACHE, m.Cmd(FAVOR, _DOWNLOAD, arg[0]).Append(mdb.TEXT))
				m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append(nfs.FILE) == ""), m.Append(mdb.TEXT))
			}
		}},
		web.P(web.UPLOAD): {Hand: func(m *ice.Message, arg ...string) {
			m.Optionv(ice.MSG_UPLOAD, web.UPLOAD)
			up := web.Upload(m)
			m.Cmd(FAVOR, mdb.INSERT, _DOWNLOAD, mdb.TYPE, kit.Ext(up[1]), mdb.NAME, up[1], mdb.TEXT, up[0], m.OptionSimple(cli.PWD, aaa.USERNAME, tcp.HOSTNAME))
			m.Echo(up[0]).Echo(ice.NL)
		}},
	})
}
