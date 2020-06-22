package web

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/toolkits"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

func _cache_list(m *ice.Message, key string) {
	if key == "" {
		m.Grows(CACHE, nil, "", "", func(index int, value map[string]interface{}) {
			m.Push("", value, []string{kit.MDB_TIME, kit.MDB_ID, kit.MDB_TYPE})
			m.Push(kit.MDB_SIZE, kit.FmtSize(kit.Int64(value[kit.MDB_SIZE])))
			m.Push("", value, []string{kit.MDB_NAME, kit.MDB_TEXT, kit.MDB_DATA})
		})
		m.Sort(kit.MDB_ID, "int_r")
		return
	}
	m.Richs(CACHE, nil, key, func(key string, value map[string]interface{}) {
		m.Push("detail", value)
		m.Push(kit.MDB_KEY, "操作")
		m.Push(kit.MDB_VALUE, `<input type="button" value="运行">`)
	})
}
func _cache_show(m *ice.Message, kind, name, text string, arg ...string) {
	if kind == "" && name == "" {
		m.Richs(CACHE, nil, m.Option(kit.MDB_DATA), func(key string, value map[string]interface{}) {
			kind = kit.Format(value[kit.MDB_TYPE])
			name = kit.Format(value[kit.MDB_NAME])
			text = kit.Format(value[kit.MDB_TEXT])
			arg = kit.Simple(value[kit.MDB_EXTRA])
			m.Log_EXPORT(kit.MDB_META, CACHE, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		})
	}
	FavorShow(m, kind, name, text, kit.Simple(arg)...)
}
func _cache_save(m *ice.Message, method, kind, name, text string, arg ...string) {
	size := kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	if method == "add" && size > 512 {
		file := kit.Hashs(text)

		// 创建文件
		if o, p, e := kit.Create(path.Join(m.Conf(CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
			defer o.Close()

			// 导入数据
			if n, e := o.WriteString(text); m.Assert(e) {
				m.Log_EXPORT(kit.MDB_FILE, p, kit.MDB_SIZE, kit.FmtSize(int64(n)))
				text, arg = p, kit.Simple(p, n)
			}
		}
	}

	// 添加数据
	h := m.Rich(CACHE, nil, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_FILE, kit.Select("", arg, 0), kit.MDB_SIZE, size,
	))
	m.Log_CREATE(CACHE, h, kit.MDB_TYPE, kind, kit.MDB_NAME, name)

	// 添加记录
	m.Grow(CACHE, nil, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_SIZE, size, kit.MDB_DATA, h,
	))

	// 返回结果
	m.Push(kit.MDB_TIME, m.Time())
	m.Push(kit.MDB_TYPE, kind)
	m.Push(kit.MDB_NAME, name)
	m.Push(kit.MDB_TEXT, text)
	m.Push(kit.MDB_SIZE, size)
	m.Push(kit.MDB_DATA, h)
}
func _cache_watch(m *ice.Message, key, file string) {
	if m.Richs(CACHE, nil, key, func(key string, value map[string]interface{}) {
		if value["file"] == "" {
			m.Cmdy(nfs.SAVE, file, value["text"])
		} else {
			m.Cmdy(nfs.LINK, file, value[kit.MDB_FILE])
		}
	}) == nil {
		m.Cmdy(SPIDE, "dev", "cache", "/cache/"+key)
		m.Cmdy(nfs.LINK, file, m.Append(kit.MDB_FILE))
	}
	m.Echo(file)
}

func _cache_catch(m *ice.Message, arg ...string) []string {
	if r, ok := m.Optionv("response").(*http.Response); ok {
		return _cache_download(m, r, arg...)
	} else if r, ok := m.Optionv("request").(*http.Request); ok {
		return _cache_upload(m, r, arg...)
	}

	if f, e := os.Open(arg[2]); m.Assert(e) {
		defer f.Close()

		// 创建文件
		h := kit.Hashs(f)
		if o, p, e := kit.Create(path.Join(m.Conf(CACHE, "meta.path"), h[:2], h)); m.Assert(e) {
			defer o.Close()

			// 导入数据
			f.Seek(0, os.SEEK_SET)
			if n, e := io.Copy(o, f); m.Assert(e) {
				m.Log_IMPORT(kit.MDB_FILE, p, kit.MDB_SIZE, kit.FmtSize(n))
				arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
			}
		}
	}
	return arg
}
func _cache_upload(m *ice.Message, r *http.Request, arg ...string) []string {
	if f, h, e := r.FormFile(kit.Select("upload", arg, 1)); e == nil {
		defer f.Close()

		// 创建文件
		file := kit.Hashs(f)
		if o, p, e := kit.Create(path.Join(m.Conf(CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
			defer o.Close()
			f.Seek(0, os.SEEK_SET)

			// 导入数据
			if n, e := io.Copy(o, f); m.Assert(e) {
				m.Log(ice.LOG_IMPORT, "%s: %s", kit.FmtSize(n), p)
				arg = kit.Simple(arg[0], h.Header.Get("Content-Type"), h.Filename, p, p, n)
			}
		}
	}
	return arg
}
func _cache_download(m *ice.Message, r *http.Response, arg ...string) []string {
	if buf, e := ioutil.ReadAll(r.Body); m.Assert(e) {
		defer r.Body.Close()

		// 创建文件
		file := kit.Hashs(string(buf))
		if o, p, e := kit.Create(path.Join(m.Conf(CACHE, "meta.path"), file[:2], file)); m.Assert(e) {
			defer o.Close()

			// 导入数据
			if n, e := o.Write(buf); m.Assert(e) {
				m.Log_IMPORT(kit.MDB_FILE, p, kit.MDB_SIZE, kit.FmtSize(int64(n)))
				arg = kit.Simple(arg[0], arg[1], arg[2], p, p, n)
			}
		}
	}
	return arg
}

const CACHE = "cache"
const (
	CATCH = "catch"
	WATCH = "watch"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CACHE: {Name: "cache", Help: "缓存池", Value: kit.Data(
				kit.MDB_SHORT, "text", "path", "var/file", "store", "var/data", "fsize", "100000", "limit", "50", "least", "30",
			)},
		},
		Commands: map[string]*ice.Command{
			CACHE: {Name: "cache data=auto auto", Help: "缓存池", Action: map[string]*ice.Action{
				CATCH: {Name: "catch type file", Help: "捕获", Hand: func(m *ice.Message, arg ...string) {
					arg = _cache_catch(m, "catch", arg[0], arg[1])
					_cache_save(m, arg[0], arg[1], arg[2], arg[3], arg[4:]...)
				}},
				WATCH: {Name: "watch key file", Help: "查看", Hand: func(m *ice.Message, arg ...string) {
					_cache_watch(m, arg[0], arg[1])
				}},
				kit.MDB_SHOW: {Name: "show type name text arg...", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 2 {
						_cache_show(m, arg[0], arg[1], arg[2], arg[3:]...)
					} else {
						_cache_show(m, "", "", "")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_cache_list(m, "")
					return
				}

				// TODO remove
				switch arg[0] {
				case "download", "upload":
					arg = _cache_catch(m, arg...)
				case "add":
					_cache_save(m, arg[0], arg[1], arg[2], arg[3], arg[4:]...)
				default:
					_cache_list(m, arg[0])
				}
			}},
			"/cache/": {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(CACHE, nil, arg[0], func(key string, value map[string]interface{}) {
					m.Render(ice.RENDER_DOWNLOAD, value["file"])
				})
			}},
		}}, nil)
}
