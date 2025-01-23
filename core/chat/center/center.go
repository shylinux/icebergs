package center

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

type center struct {
	list string `name:"list list" help:"云游"`
}

func (s center) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		m.Cmd(web.SPACE).Table(func(value ice.Maps) {
			if value[mdb.TYPE] == web.SERVER {
				m.PushRecord(value, mdb.NAME, mdb.ICONS, nfs.MODULE, nfs.VERSION)
			}
		})
		m.Display("/plugin/story/spides.js?split=.").Option(nfs.DIR_ROOT, ice.Info.NodeName)
	} else {
		m.Cmdy(web.SPACE, arg[0], m.PrefixKey()).Table(func(value ice.Maps) {
			m.Push(nfs.FILE, kit.Keys(arg[0], value[mdb.NAME]))
		})
		if m.Length() == 0 {
			m.Push(web.SPACE, arg[0]).Push(ctx.INDEX, web.DESKTOP)
		}
	}
}

func init() { ice.Cmd("web.chat.center.center", center{}) }
