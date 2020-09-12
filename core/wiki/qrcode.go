package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const QRCODE = "qrcode"

func init() {
	Index.Register(&ice.Context{Name: QRCODE, Help: "二维码",
		Configs: map[string]*ice.Config{
			QRCODE: {Name: "qrcode", Help: "二维码", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			QRCODE: {Name: "qrcode", Help: "二维码", Action: map[string]*ice.Action{
				mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
					m.Grow(QRCODE, kit.Keys(m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), func(index int, value map[string]interface{}) {
						m.Push("", value, []string{kit.MDB_TIME, kit.MDB_TEXT})
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Grow(QRCODE, kit.Keys(m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), func(index int, value map[string]interface{}) {
					m.Push("", value, []string{kit.MDB_TIME, kit.MDB_TEXT})
				})
			}},
		},
	}, nil)
}
