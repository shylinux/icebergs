package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
)

const SPIDED = "spided"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SPIDED: {Name: "spided", Help: "网页爬虫", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_LINK, kit.MDB_PATH, "usr/spide",
			)},
		},
		Commands: map[string]*ice.Command{
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
		},
	}, nil)
}
