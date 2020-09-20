package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"

	"path"
)

const CACHE = "cache"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CACHE: {Name: CACHE, Help: "爬虫缓存", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_FIELD, "time,step,size,total,action,text,name,type,link",
				kit.MDB_PATH, "usr/spide",
			)},
		},
		Commands: map[string]*ice.Command{
			CACHE: {Name: "cache hash auto 添加 清理", Help: "爬虫缓存", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create link type name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option("_process", "_progress")
					if m.Cmdy(mdb.SELECT, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_LINK, m.Option(kit.MDB_LINK)); len(m.Appendv(kit.MDB_TOTAL)) > 0 {
						return // 已经下载
					}
					m.Cmd(mdb.INSERT, m.Prefix(CACHE), "", mdb.HASH,
						kit.MDB_LINK, m.Option(kit.MDB_LINK),
						kit.MDB_TYPE, m.Option(kit.MDB_TYPE),
						kit.MDB_NAME, m.Option(kit.MDB_NAME),
						kit.MDB_TEXT, m.Option(kit.MDB_TEXT),
					)

					m.Option("progress", m.Prefix(CACHE), "", m.Option(kit.MDB_LINK))
					msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(kit.MDB_LINK))

					p := path.Join(m.Conf(m.Prefix(CACHE), kit.META_PATH), m.Option(kit.MDB_NAME))
					m.Cmdy(nfs.LINK, p, msg.Append(kit.MDB_FILE))

					if file, e := avutil.Open(p); e == nil {
						defer file.Close()

						if streams, e := file.Streams(); e == nil {
							for _, stream := range streams {

								switch {
								case stream.Type().IsAudio():
								case stream.Type().IsVideo():
									// 自动分屏
									vstream := stream.(av.VideoCodecData)
									m.Cmd(nfs.LINK, path.Join(m.Conf(m.Prefix(CACHE), kit.META_PATH),
										kit.Select("竖屏", "横屏", vstream.Width() > vstream.Height()), m.Option(kit.MDB_NAME)), p)
								}
							}
						}
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_STEP, "100")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(ice.MSG_ACTION, "删除")
				m.Option(mdb.FIELDS, m.Conf(m.Prefix(CACHE), kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(CACHE), "", mdb.HASH)
			}},
		},
	}, nil)
}
