package location

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const AMAP = "amap"

func init() {
	web.Index.MergeCommands(ice.Commands{
		"/_AMapService/": {Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(web.SPIDE, ice.DEV, web.SPIDE_RAW, m.R.Method, kit.MergeURL("https://restapi.amap.com/"+path.Join(arg...)+"?"+m.R.URL.RawQuery,
				"jscode", mdb.Conf(m, chat.Prefix(AMAP), kit.Keym(aaa.SECRET))),
			).RenderResult()
		}},
	})
	chat.Index.MergeCommands(ice.Commands{
		AMAP: {Help: "高德地图", Hand: func(m *ice.Message, arg ...string) {
			m.Display("", nfs.SCRIPT, mdb.Config(m, nfs.SCRIPT))
		}},
	})
}
