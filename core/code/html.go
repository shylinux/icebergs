package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		HTML: {Name: "html path auto", Help: "网页", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				p := kit.MergeURL(path.Join("/require/", arg[2], arg[1]), "_", kit.Hashs("uniq"))
				m.Push(mdb.LINK, p).Echo(`<iframe src="%s"></iframe>`, p).StatusTime(web.LINK, p)
			}},
		}, PlugAction())},
	})
}
