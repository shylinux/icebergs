package crx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"encoding/csv"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"io/ioutil"
	"os"
	"path"
	"sort"
)

const CHROME = "chrome"
const (
	SPIDED = "spided"
	CACHED = "cached"
)

var Index = &ice.Context{Name: "chrome", Help: "浏览器",
	Configs: map[string]*ice.Config{
		CHROME: {Name: "chrome", Help: "浏览器", Value: kit.Data(
			kit.MDB_SHORT, "name", "history", "url.history",
		)},
		SPIDED: {Name: "spided", Help: "网页爬虫", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_PATH, "usr/spide",
		)},
		CACHED: {Name: "spided", Help: "网页爬虫", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_PATH, "usr/spide",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(SPIDED, CACHED)
		}},
		CHROME: {Name: "chrome wid=auto url auto 编译:button 下载:button", Help: "浏览器", Action: map[string]*ice.Action{
			"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(web.SPACE, CHROME, CHROME, arg)
		}},
		SPIDED: {Name: "spided wid=auto tid=auto cmd auto", Help: "网页爬虫", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				if m.Richs(CACHED, "", m.Option("link"), func(key string, value map[string]interface{}) {
					if _, e := os.Stat(path.Join(m.Conf(CACHED, kit.META_PATH), m.Option("name"))); e == nil {
						m.Push(key, value)
					}
				}) != nil && len(m.Appendv("name")) > 0 {
					return
				}

				m.Cmd(mdb.INSERT, m.Prefix(CACHED), "", mdb.HASH,
					kit.MDB_LINK, m.Option("link"),
					kit.MDB_TYPE, m.Option("type"),
					kit.MDB_NAME, m.Option("name"),
					kit.MDB_TEXT, m.Option("text"),
				)

				// 进度
				m.Richs(CACHED, "", m.Option("link"), func(key string, value map[string]interface{}) {
					m.Optionv("progress", func(size int, total int) {
						p := size * 100 / total
						if p != value["progress"] {
							m.Log_IMPORT(kit.MDB_FILE, m.Option("name"), "per", size*100/total, kit.MDB_SIZE, kit.FmtSize(int64(size)), "total", kit.FmtSize(int64(total)))
						}
						value["progress"], value["size"], value["total"] = p, size, total
					})
				})

				// 下载
				msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option("link"))
				p := path.Join(m.Conf(CACHED, kit.META_PATH), m.Option("name"))
				m.Cmdy(nfs.LINK, p, msg.Append("file"))

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
			"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			msg := m.Cmd(web.SPACE, CHROME, CHROME, arg)
			switch kit.Select("spide", arg, 2) {
			case "cache":
				m.Option("fields", "time,type,progress,size,total,name,text,link")
				m.Cmdy(mdb.SELECT, m.Prefix(SPIDED), "", mdb.HASH)
			case "spide":
				if len(arg) > 1 {
					msg.PushAction("下载")
					msg.Table(func(index int, value map[string]string, head []string) {
						m.Push("time", value["time"])
						m.Push("type", value["type"])
						m.Push("action", value["action"])
						m.Push("name", value["name"])
						switch value["type"] {
						case "img":
							m.Push("text", m.Cmdx(mdb.RENDER, web.RENDER.IMG, value["text"]))
						case "video":
							m.Push("text", m.Cmdx(mdb.RENDER, web.RENDER.Video, value["text"]))
						default:
							m.Push("text", value["text"])
						}
						m.Push("link", value["link"])
					})
					break
				}
				fallthrough
			default:
				m.Copy(msg)
			}
		}},
		CACHED: {Name: "cached hash=auto auto 清理:button 导出:button", Help: "网页爬虫", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Richs(CACHED, "", m.Option("link"), func(key string, value map[string]interface{}) {
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

		"/crx": {Name: "/crx", Help: "插件", Action: map[string]*ice.Action{
			web.HISTORY: {Name: "history", Help: "历史记录", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, web.SPIDE_DEV, "/code/chrome/favor", "cmds", mdb.INSERT, "sid", m.Option("sid"),
					"tab", m.Conf(CHROME, "meta.history"), "name", arg[1], "note", arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.FAVOR, mdb.INSERT, m.Option("tab"), web.SPIDE, m.Option("name"), m.Option("note"))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
