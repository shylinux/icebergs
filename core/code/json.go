package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	DEMO_JSON = "demo.json"
)

const JSON = "json"

func init() {
	Index.MergeCommands(ice.Commands{
		JSON: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if path.Base(arg[1]) == "portal.json" {
					if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
						ctx.ProcessField(m, cmd, kit.Simple("table"))
						return
					}
				}
				m.FieldsSetDetail()
				kit.For(kit.KeyValue(nil, "", kit.UnMarshal(m.Cmdx(nfs.CAT, path.Join(arg[2], arg[1])))), func(key, value string) {
					m.Push(key, value)
				})
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1]))
				ctx.DisplayStoryJSON(m)
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Format(nfs.Template(m, DEMO_JSON)))
			}},
		}, PlugAction())},
	})
}
