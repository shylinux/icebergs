package code

import (
	ice "shylinux.com/x/icebergs"
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
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Format(nfs.Template(m, DEMO_JSON)))
			}},
		}, PlugAction())},
	})
}
