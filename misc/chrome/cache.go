package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"

	"encoding/csv"
	"io/ioutil"
	"os"
	"path"
	"sort"
)

const CACHED = "cached"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CACHED: {Name: "spided", Help: "网页爬虫", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_PATH, "usr/spide",
			)},
		},
		Commands: map[string]*ice.Command{
			CACHED: {Name: "cached hash=auto auto 清理:button 导出:button", Help: "网页爬虫", Action: map[string]*ice.Action{
				"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(CACHED, "", m.Option("link"), func(key string, value map[string]interface{}) {
						value = value[kit.MDB_META].(map[string]interface{})
						m.Optionv("progress", func(size int, total int) {
							value["progress"], value["size"], value["total"] = size*100/total, size, total
							m.Log_IMPORT(kit.MDB_FILE, m.Option("name"), "per", size*100/total, kit.MDB_SIZE, kit.FmtSize(int64(size)), "total", kit.FmtSize(int64(total)))
						})
					})

					msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option("link"))
					p := path.Join(m.Conf(CACHED, kit.META_PATH), m.Option("name"))
					m.Cmdy(nfs.LINK, p, msg.Append("file"))

					// 完成
					if file, e := avutil.Open(p); m.Assert(e) {
						defer file.Close()
						if streams, e := file.Streams(); m.Assert(e) {
							for _, stream := range streams {
								if stream.Type().IsAudio() {

								} else if stream.Type().IsVideo() {
									vstream := stream.(av.VideoCodecData)
									if vstream.Width() > vstream.Height() {
										m.Cmdy(nfs.LINK, path.Join(m.Conf(CACHED, kit.META_PATH), "横屏", m.Option("name")), p)
									} else {
										m.Cmdy(nfs.LINK, path.Join(m.Conf(CACHED, kit.META_PATH), "竖屏", m.Option("name")), p)
									}
								}
							}
						}
					}
				}},
				"prune": {Name: "prune", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(CACHED, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if kit.Int(value["progress"]) == 100 {
							dir := path.Join("var/data", m.Prefix(CACHED), "")
							name := path.Join(dir, kit.Keys(key, "json"))
							if f, p, e := kit.Create(name); e == nil {
								defer f.Close()
								// 保存数据
								if n, e := f.WriteString(kit.Format(value)); e == nil {
									m.Log_EXPORT("file", p, kit.MDB_SIZE, n)
								}
							}
							m.Conf(CACHED, kit.Keys(kit.MDB_HASH, key), "")
						}
					})
				}},
				"export": {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					f, p, e := kit.Create(path.Join("usr/export", m.Prefix(CACHED), "list.csv"))
					m.Assert(e)
					defer f.Close()

					w := csv.NewWriter(f)
					defer w.Flush()

					count := 0
					head := []string{}
					m.Cmd(nfs.DIR, path.Join("var/data", m.Prefix(CACHED))+"/").Table(func(index int, v map[string]string, h []string) {

						f, e := os.Open(v["path"])
						m.Assert(e)
						defer f.Close()

						b, e := ioutil.ReadAll(f)
						m.Assert(e)

						value, ok := kit.UnMarshal(string(b)).(map[string]interface{})
						if !ok {
							return
						}

						if index == 0 {
							// 输出表头
							for k := range value {
								head = append(head, k)
							}
							sort.Strings(head)
							w.Write(head)
						}

						// 输出数据
						data := []string{}
						for _, k := range head {
							data = append(data, kit.Format(value[k]))
						}
						w.Write(data)
						count++
					})
					m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_COUNT, count)
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("cache.limit", 100)
				m.Option("fields", "time,hash,type,progress,size,total,name,text,link")
				m.Cmdy(mdb.SELECT, m.Prefix(CACHED), "", mdb.HASH)
				m.Sort("time", "time_r")
				m.PushAction("下载")
				m.Appendv(ice.MSG_APPEND, "time", "type", "name", "text",
					"action", "progress", "size", "total", "hash", "link")
			}},
		},
	}, nil)
}
