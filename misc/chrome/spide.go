package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SPIDE: {Name: SPIDE, Help: "网页爬虫", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_PATH, "usr/spide",
			)},
		},
		Commands: map[string]*ice.Command{
			SPIDE: {Name: "spide wid tid cmd auto", Help: "网页爬虫", Action: map[string]*ice.Action{
				web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(CACHE, mdb.CREATE, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch msg := m.Cmd(web.SPACE, CHROME, CHROME, arg); kit.Select(SPIDE, arg, 2) {
				case SPIDE:
					if len(arg) > 1 {
						msg.Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_TIME, value[kit.MDB_TIME])
							m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
							m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
							m.PushButton(web.DOWNLOAD)
							m.PushRender(kit.MDB_TEXT, value[kit.MDB_TYPE], value[kit.MDB_LINK])
							m.Push(kit.MDB_LINK, value[kit.MDB_LINK])
						})
						break
					}
					fallthrough
				default:
					m.Copy(msg)
				}
			}},
		},
	})
}
