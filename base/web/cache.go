package web

import (
	"io"
	"net/http"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _cache_name(m *ice.Message, h string) string {
	return path.Join(m.Config(nfs.PATH), h[:2], h)
}
func _cache_save(m *ice.Message, kind, name, text string, arg ...string) { // file size
	if name == "" {
		return
	}
	if len(text) > 512 || kind == "go" { // 存入文件
		p := m.Cmdx(nfs.SAVE, _cache_name(m, kit.Hashs(text)), text)
		text, arg = p, kit.Simple(p, len(text))
	}

	// 添加数据
	size := kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	file := kit.Select("", arg, 0)
	text = kit.Select(file, text)
	h := m.Cmdx(mdb.INSERT, CACHE, "", mdb.HASH,
		mdb.TYPE, kind, mdb.NAME, name, mdb.TEXT, text,
		nfs.FILE, file, nfs.SIZE, size)

	// 返回结果
	m.Push(mdb.TIME, m.Time())
	m.Push(mdb.TYPE, kind)
	m.Push(mdb.NAME, name)
	m.Push(mdb.TEXT, text)
	m.Push(nfs.SIZE, size)
	m.Push(nfs.FILE, file)
	m.Push(mdb.HASH, h)
	m.Push(DATA, h)
}
func _cache_watch(m *ice.Message, key, file string) {
	mdb.HashSelect(m.Spawn(), key).Table(func(index int, value map[string]string, head []string) {
		if value[nfs.FILE] == "" {
			m.Cmdy(nfs.SAVE, file, value[mdb.TEXT])
		} else {
			m.Cmdy(nfs.LINK, file, value[nfs.FILE])
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
				m.Log_IMPORT(nfs.FILE, p, nfs.SIZE, kit.FmtSize(int64(n)))
				return h.Header.Get(ContentType), h.Filename, p, kit.Format(n)
			}
		}
	}
	return "", "", "", "0"
}
func _cache_download(m *ice.Message, r *http.Response) (file, size string) {
	defer r.Body.Close()

	if f, p, e := kit.Create(path.Join(ice.VAR_TMP, kit.Hashs("uniq"))); m.Assert(e) {
		step, total := 0, kit.Int(kit.Select("1", r.Header.Get(ContentLength)))
		size, buf := 0, make([]byte, ice.MOD_BUFS)

		for {
			if n, _ := r.Body.Read(buf); n > 0 {
				size += n
				f.Write(buf[0:n])
				s := size * 100 / total

				switch cb := m.OptionCB(SPIDE).(type) {
				case func(int, int, int):
					cb(size, total, s)
				case func(int, int):
					cb(size, total)
				case []string:
					m.Richs(cb[0], cb[1], cb[2], func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						value[mdb.VALUE], value[mdb.COUNT], value[mdb.TOTAL] = kit.Format(s), size, total
					})
				default:
					if s != step && s%10 == 0 {
						m.Log_IMPORT(nfs.FILE, p, mdb.VALUE, s,
							mdb.COUNT, kit.FmtSize(int64(size)), mdb.TOTAL, kit.FmtSize(int64(total)))
					}
				}
				step = s
				continue
			}

			f.Close()
			break
		}

		if f, e := os.Open(p); m.Assert(e) {
			defer f.Close()

			m.Log_IMPORT(nfs.FILE, p, nfs.SIZE, kit.FmtSize(int64(size)))
			c := _cache_name(m, kit.Hashs(f))
			m.Cmd(nfs.LINK, c, p)
			return c, kit.Format(size)
		}
	}
	return "", "0"
}

const (
	WATCH    = "watch"
	CATCH    = "catch"
	WRITE    = "write"
	UPLOAD   = "upload"
	DOWNLOAD = "download"
)
const CACHE = "cache"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CACHE: {Name: CACHE, Help: "缓存池", Value: kit.Data(
			mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,size,type,name,text",
			kit.MDB_STORE, ice.VAR_DATA, nfs.PATH, ice.VAR_FILE, kit.MDB_FSIZE, "200000",
			mdb.LIMIT, "50", kit.MDB_LEAST, "30",
		)},
	}, Commands: map[string]*ice.Command{
		"/cache/": {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(CACHE, nil, arg[0], func(key string, value map[string]interface{}) {
				if kit.Format(value[nfs.FILE]) == "" {
					m.RenderResult(value[mdb.TEXT])
				} else {
					m.RenderDownload(value[nfs.FILE])
				}
			})
		}},
		CACHE: {Name: "cache hash auto", Help: "缓存池", Action: ice.MergeAction(map[string]*ice.Action{
			WATCH: {Name: "watch key file", Help: "释放", Hand: func(m *ice.Message, arg ...string) {
				_cache_watch(m, arg[0], arg[1])
			}},
			CATCH: {Name: "catch type name", Help: "捕获", Hand: func(m *ice.Message, arg ...string) {
				file, size := _cache_catch(m, arg[1])
				_cache_save(m, arg[0], arg[1], "", file, size)
			}},
			WRITE: {Name: "write type name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_cache_save(m, arg[0], arg[1], arg[2])
			}},
			UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				kind, name, file, size := _cache_upload(m, m.R)
				_cache_save(m, kind, name, "", file, size)
			}},
			DOWNLOAD: {Name: "download type name", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				if r, ok := m.Optionv(RESPONSE).(*http.Response); ok {
					file, size := _cache_download(m, r)
					_cache_save(m, arg[0], arg[1], "", file, size)
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				return
			}

			if m.Append(nfs.FILE) == "" {
				m.PushScript("inner", m.Append(mdb.TEXT))
			} else {
				m.PushDownload(m.Append(mdb.NAME), m.MergeURL2("/share/cache/"+arg[0]))
			}
		}},
	}})
}
