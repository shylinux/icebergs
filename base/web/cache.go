package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"io"
	"net/http"
	"os"
	"path"
)

func _cache_name(m *ice.Message, h string) string {
	return path.Join(m.Conf(CACHE, "meta.path"), h[:2], h)
}

func _cache_list(m *ice.Message, key string) {
	if key == "" {
		m.Grows(CACHE, nil, "", "", func(index int, value map[string]interface{}) {
			// 缓存列表
			m.Push("", value, []string{kit.MDB_TIME, kit.MDB_ID, DATA, kit.MDB_TYPE})
			m.Push(kit.MDB_SIZE, kit.FmtSize(kit.Int64(value[kit.MDB_SIZE])))
			m.Push("", value, []string{kit.MDB_NAME, kit.MDB_TEXT})
		})
		m.Sort(kit.MDB_ID, "int_r")
		return
	}
	m.Richs(CACHE, nil, key, func(key string, value map[string]interface{}) {
		// 缓存详情
		m.Push("detail", value)
	})
}
func _cache_save(m *ice.Message, kind, name, text string, arg ...string) { // file size
	if name == "" {
		return
	}
	if len(text) > 512 {
		// 存入文件
		p := m.Cmdx(nfs.SAVE, _cache_name(m, kit.Hashs(text)), text)
		text, arg = p, kit.Simple(p, len(text))
	}

	// 添加数据
	size := kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	file := kit.Select("", arg, 0)
	text = kit.Select(file, text)
	h := m.Rich(CACHE, nil, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_FILE, file, kit.MDB_SIZE, size,
	))
	m.Log_CREATE(CACHE, h, kit.MDB_TYPE, kind, kit.MDB_NAME, name)

	// 添加记录
	m.Grow(CACHE, nil, kit.Dict(
		kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
		kit.MDB_SIZE, size, DATA, h,
	))

	// 返回结果
	m.Push(kit.MDB_TIME, m.Time())
	m.Push(kit.MDB_TYPE, kind)
	m.Push(kit.MDB_NAME, name)
	m.Push(kit.MDB_TEXT, text)
	m.Push(kit.MDB_SIZE, size)
	m.Push(kit.MDB_FILE, file)
	m.Push(DATA, h)
}
func _cache_watch(m *ice.Message, key, file string) {
	m.Richs(CACHE, nil, key, func(key string, value map[string]interface{}) {
		if value[kit.MDB_FILE] == nil {
			m.Cmdy(nfs.SAVE, file, value[kit.MDB_TEXT])
		} else {
			m.Cmdy(nfs.LINK, file, value[kit.MDB_FILE])
		}
	})
}

func _cache_catch(m *ice.Message, name string) (file, size string) {
	if f, e := os.Open(name); m.Assert(e) {
		defer f.Close()

		if s, e := f.Stat(); m.Assert(e) {
			return m.Cmdx(nfs.LINK, _cache_name(m, kit.Hashs(f)), name), kit.Format(s.Size())
		}
	}
	return "", "0"
}
func _cache_upload(m *ice.Message, r *http.Request) (kind, name, file, size string) {
	if buf, h, e := r.FormFile(UPLOAD); e == nil {
		defer buf.Close()

		// 创建文件
		if f, p, e := kit.Create(_cache_name(m, kit.Hashs(buf))); m.Assert(e) {
			defer f.Close()

			// 导入数据
			buf.Seek(0, os.SEEK_SET)
			if n, e := io.Copy(f, buf); m.Assert(e) {
				m.Log_IMPORT(kit.MDB_FILE, p, kit.MDB_SIZE, kit.FmtSize(int64(n)))
				return h.Header.Get("Content-Type"), h.Filename, p, kit.Format(n)
			}
		}
	}
	return "", "", "", "0"
}
func _cache_download(m *ice.Message, r *http.Response) (file, size string) {
	defer r.Body.Close()

	total := kit.Int(kit.Select("1", r.Header.Get("Content-Length")))
	if f, p, e := kit.Create(path.Join("var/tmp", kit.Hashs("uniq"))); m.Assert(e) {
		size, buf := 0, make([]byte, 1024)
		for {
			if n, _ := r.Body.Read(buf); n > 0 {
				f.Write(buf[0:n])
				switch size += n; cb := m.Optionv("progress").(type) {
				case []string:
					m.Richs(cb[0], cb[1], cb[2], func(key string, value map[string]interface{}) {
						value = value[kit.MDB_META].(map[string]interface{})

						s := size * 100 / total
						if s != kit.Int(value[kit.SSH_STEP]) && s%10 == 0 {
							m.Log_IMPORT(kit.MDB_FILE, path.Base(cb[2]), kit.SSH_STEP, s, kit.MDB_SIZE, kit.FmtSize(int64(size)), kit.MDB_TOTAL, kit.FmtSize(int64(total)))
						}
						value[kit.SSH_STEP], value[kit.MDB_SIZE], value[kit.MDB_TOTAL] = kit.Format(s), size, total
					})
				case func(int, int):
					cb(size, total)
				default:
					m.Log_IMPORT(kit.MDB_FILE, p, "per", size*100/total, kit.MDB_SIZE, kit.FmtSize(int64(size)), "total", kit.FmtSize(int64(total)))
				}

			} else {
				f.Close()
				break
			}
		}
		if f, e := os.Open(p); m.Assert(e) {
			defer f.Close()
			m.Log_IMPORT(kit.MDB_FILE, p, kit.MDB_SIZE, kit.FmtSize(int64(size)))
			c := _cache_name(m, kit.Hashs(f))
			m.Cmd(nfs.LINK, c, p)
			return c, kit.Format(size)
		}
	}
	return "", "0"
}

const (
	WATCH = "watch"
	CATCH = "catch"
	WRITE = "write"

	UPLOAD   = "upload"
	DOWNLOAD = "download"
)
const CACHE = "cache"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CACHE: {Name: "cache", Help: "缓存池", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_TEXT,
				kit.MDB_PATH, "var/file",
				kit.MDB_STORE, "var/data",
				kit.MDB_FSIZE, "200000",
				kit.MDB_LIMIT, "50",
				kit.MDB_LEAST, "30",
			)},
		},
		Commands: map[string]*ice.Command{
			CACHE: {Name: "cache data=auto auto", Help: "缓存池", Action: map[string]*ice.Action{
				WATCH: {Name: "watch key file", Help: "查看", Hand: func(m *ice.Message, arg ...string) {
					_cache_watch(m, arg[0], arg[1])
				}},
				WRITE: {Name: "write type name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_cache_save(m, arg[0], arg[1], arg[2])
				}},
				CATCH: {Name: "catch type name", Help: "捕获", Hand: func(m *ice.Message, arg ...string) {
					file, size := _cache_catch(m, arg[1])
					_cache_save(m, arg[0], arg[1], "", file, size)
				}},
				UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					kind, name, file, size := _cache_upload(m, m.R)
					_cache_save(m, kind, name, "", file, size)
				}},
				DOWNLOAD: {Name: "download type name", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					if r, ok := m.Optionv("response").(*http.Response); ok {
						file, size := _cache_download(m, r)
						_cache_save(m, arg[0], arg[1], "", file, size)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,size,type,name,text,file")
				m.Cmdy(mdb.SELECT, m.Prefix(CACHE), "", mdb.HASH, kit.MDB_HASH, arg)
			}},
			"/cache/": {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(CACHE, nil, arg[0], func(key string, value map[string]interface{}) {
					if value[kit.MDB_FILE] == nil {
						m.Render(ice.RENDER_DOWNLOAD, value[kit.MDB_FILE])
					} else {
						m.Render(ice.RENDER_RESULT, value[kit.MDB_TEXT])
					}
				})
			}},
		}}, nil)
}
