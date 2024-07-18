package wx

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		"ocr": {Name: "ocr access path type auto", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS)
			} else if len(arg) == 1 || strings.HasSuffix(arg[1], "/") {
				m.Cmdy(nfs.DIR, arg[1:])
			} else if arg[1] != "" {
				res := SpidePost(m, "/cv/ocr/"+kit.Select("idcard", arg, 2), web.SPIDE_PART, "img", "@"+arg[1])
				if len(arg) > 2 && arg[2] == "comm" {
					kit.For(kit.Value(res, "items"), func(value ice.Map) { m.Push(mdb.TEXT, value[mdb.TEXT]) })
				} else {
					m.PushDetail(res)
				}
				m.EchoImages(m.Resource(arg[1]))
			}
		}},
	})
}
