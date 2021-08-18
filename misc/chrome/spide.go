package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{
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

							switch m.PushButton(web.DOWNLOAD); value[kit.MDB_TYPE] {
							case "img":
								m.PushImages(kit.MDB_TEXT, value[kit.MDB_LINK])
							case "video":
								m.PushVideos(kit.MDB_TEXT, value[kit.MDB_LINK])
							default:
								m.Push(kit.MDB_TEXT, "")
							}
							m.Push(kit.MDB_LINK, value[kit.MDB_LINK])
						})
						m.StatusTimeCount()
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
