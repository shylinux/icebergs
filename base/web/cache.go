package web

import (
	"io"
	"io/ioutil"
	"net/http"
	"path"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"os"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Value: kit.Data(
				kit.MDB_SHORT, "text", "path", "var/file", "store", "var/data", "fsize", "100000", "limit", "50", "least", "30",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_CACHE: {Name: "cache", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 缓存记录
					m.Grows(ice.WEB_CACHE, nil, "", "", func(index int, value map[string]interface{}) {
						m.Push("", value)
					})
					return
				}

				switch arg[0] {
				case "catch":
					// 导入文件
					if f, e := os.Open(arg[2]); m.Assert(e) {
						defer f.Close()

						// 创建文件
						h := kit.Hashs(f)
						if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), h[:2], h)); m.Assert(e) {
							defer o.Close()
							f.Seek(0, os.SEEK_SET)

							// 导入数据
							if n, e := io.Copy(o, f); m.Assert(e) {
								m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(n), p)
								arg = kit.Simple(arg[0], arg[1], path.Base(arg[2]), p, p, n)
							}
						}
					}
					fallthrough
				case "upload", "download":
					if m.R != nil {
						// 上传文件
						if f, h, e := m.R.FormFile(kit.Select("upload", arg, 1)); e == nil {
							defer f.Close()

							// 创建文件
							file := kit.Hashs(f)
							if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
								defer o.Close()
								f.Seek(0, os.SEEK_SET)

								// 导入数据
								if n, e := io.Copy(o, f); m.Assert(e) {
									m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(n), p)
									arg = kit.Simple(arg[0], h.Header.Get("Content-Type"), h.Filename, p, p, n)
								}
							}
						}
					} else if r, ok := m.Optionv("response").(*http.Response); ok {
						// 下载文件
						if buf, e := ioutil.ReadAll(r.Body); m.Assert(e) {
							defer r.Body.Close()

							// 创建文件
							file := kit.Hashs(string(buf))
							if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
								defer o.Close()

								// 导入数据
								if n, e := o.Write(buf); m.Assert(e) {
									m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(int64(n)), p)
									arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
								}
							}
						}
					}
					fallthrough
				case "add":
					size := kit.Int(kit.Select(kit.Format(len(arg[3])), arg, 5))
					if arg[0] == "add" && size > 512 {
						// 创建文件
						file := kit.Hashs(arg[3])
						if o, p, e := kit.Create(path.Join(m.Conf(ice.WEB_CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
							defer o.Close()

							// 导入数据
							if n, e := o.WriteString(arg[3]); m.Assert(e) {
								m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(int64(n)), p)
								arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
							}
						}
					}

					// 添加数据
					h := m.Rich(ice.WEB_CACHE, nil, kit.Dict(
						kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
						kit.MDB_FILE, kit.Select("", arg, 4), kit.MDB_SIZE, size,
					))
					m.Log(ice.LOG_CREATE, "cache: %s %s: %s", h, arg[1], arg[2])

					// 添加记录
					m.Grow(ice.WEB_CACHE, nil, kit.Dict(
						kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
						kit.MDB_SIZE, size, "data", h,
					))

					// 返回结果
					m.Push("time", m.Time())
					m.Push("type", arg[1])
					m.Push("name", arg[2])
					m.Push("text", arg[3])
					m.Push("size", size)
					m.Push("data", h)
				case "watch":
					if m.Richs(ice.WEB_CACHE, nil, arg[1], func(key string, value map[string]interface{}) {
						if value["file"] == "" {
							if f, _, e := kit.Create(arg[2]); m.Assert(e) {
								defer f.Close()
								f.WriteString(kit.Format(value["text"]))
							}
						} else {
							os.MkdirAll(path.Dir(arg[2]), 0777)
							os.Remove(arg[2])
							os.Link(kit.Format(value["file"]), arg[2])
						}
					}) == nil {
						m.Cmdy(ice.WEB_SPIDE, "dev", "cache", "/cache/"+arg[1])
						os.MkdirAll(path.Dir(arg[2]), 0777)
						os.Remove(arg[2])
						os.Link(m.Append("file"), arg[2])
					}
					m.Echo(arg[2])
				}
			}},
			"/cache/": {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(ice.WEB_CACHE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Render(ice.RENDER_DOWNLOAD, value["file"])
				})
			}},
		}}, nil)
}
