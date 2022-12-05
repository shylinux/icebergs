package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _website_url(m *ice.Message, file string) string {
	return strings.Split(web.MergePodWebSite(m, "", file), "?")[0]
}

const ZML = nfs.ZML

func init() {
	const (
		SRC_WEBSITE = "src/website/"
	)
	Index.MergeCommands(ice.Commands{
		ZML: {Name: "zml", Help: "网页", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				switch kit.Ext(m.Option(mdb.FILE)) {
				case ZML:
					m.Echo(`
{
	username
	系统
		命令 index cli.system
		共享 index cli.qrcode
	代码
		趋势 index web.code.git.trend args icebergs action auto 
		状态 index web.code.git.status args icebergs
	脚本
		终端 index hi/hi.sh
		文档 index hi/hi.shy
		数据 index hi/hi.py
		后端 index hi/hi.go
		前端 index hi/hi.js
}
`)
				case nfs.IML:
					m.Echo(`
系统
	命令
		cli.system
	环境
		cli.runtime
开发
	模块
		hi/hi.go
	脚本
		hi/hi.sh
		hi/hi.shy
		hi/hi.py
		hi/hi.go
		hi/hi.js
`)
				}
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.FOREACH {
					switch m.Option(ctx.ACTION) {
					case web.WEBSITE:
						m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, SRC_WEBSITE), nfs.PATH)
					}
					return
				}

				switch kit.Select("", kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t \n`"), -1), 0) {
				case mdb.TYPE:
					m.Push(mdb.NAME, "menu")

				case ctx.INDEX:
					m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields("index,name,text"))

				case ctx.ACTION:
					m.Push(mdb.NAME, "auto")
					m.Push(mdb.NAME, "push")
					m.Push(mdb.NAME, "open")

				default:
					if strings.HasSuffix(m.Option(mdb.TEXT), ice.SP) {
						m.Push(mdb.NAME, "index")
						m.Push(mdb.NAME, "action")
						m.Push(mdb.NAME, "args")
						m.Push(mdb.NAME, "type")
					} else if m.Option(mdb.TEXT) == "" {
						m.Push(mdb.NAME, "head")
						m.Push(mdb.NAME, "left")
						m.Push(mdb.NAME, "main")
						m.Push(mdb.NAME, "foot")
					}
				}
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoIFrame(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_website_url(m, strings.TrimPrefix(path.Join(arg[2], arg[1]), SRC_WEBSITE)))
			}},
		}, PlugAction())},
	})
}
