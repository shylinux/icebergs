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
	"shylinux.com/x/toolkits/miss"
)

func _cache_name(m *ice.Message, h string) string {
	return path.Join(ice.VAR_FILE, h[:2], h)
}
func _cache_save(m *ice.Message, kind, name, text string, arg ...string) { // file size
	if name == "" {
		return
	}
	if len(text) > 512 || kind == nfs.GO { // 存入文件
		p := m.Cmdx(nfs.SAVE, _cache_name(m, kit.Hashs(text)), text)
		text, arg = p, kit.Simple(p, len(text))
	}

	// 添加数据
	size := kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	file := kit.Select("", arg, 0)
	text = kit.Select(file, text)
	h := mdb.HashCreate(m, kit.SimpleKV("", kind, name, text), nfs.FILE, file, nfs.SIZE, size).Result()

	// 返回结果
	m.Push(mdb.TIME, m.Time())
	m.Push(mdb.TYPE, kind)
	m.Push(mdb.NAME, name)
	m.Push(mdb.TEXT, text)
	m.Push(nfs.SIZE, size)
	m.Push(nfs.FILE, file)
	m.Push(mdb.HASH, h)
	m.Push(mdb.DATA, h)
}
func _cache_watch(m *ice.Message, key, file string) {
	mdb.HashSelect(m.Spawn(), key).Tables(func(value ice.Maps) {
		if value[nfs.FILE] == "" {
			m.Cmdy(nfs.SAVE, file, value[mdb.TEXT])
		} else {
			m.Cmdy(nfs.LINK, file, value[nfs.FILE])
		}
	})
}
func _cache_catch(m *ice.Message, name string) (file, size string) {
	if f, e := nfs.OpenFile(m, name); m.Assert(e) {
		defer f.Close()

		if s, e := nfs.StatFile(m, name); m.Assert(e) {
			return m.Cmdx(nfs.LINK, _cache_name(m, kit.Hashs(f)), name), kit.Format(s.Size())
		}
	}
	return "", "0"
}
func _cache_upload(m *ice.Message, r *http.Request) (kind, name, file, size string) {
	if b, h, e := r.FormFile(UPLOAD); e == nil {
		defer b.Close()

		// 创建文件
		if f, p, e := miss.CreateFile(_cache_name(m, kit.Hashs(b))); m.Assert(e) {
			defer f.Close()

			// 导入数据
			b.Seek(0, os.SEEK_SET)
			if n, e := io.Copy(f, b); m.Assert(e) {
				m.Log_IMPORT(nfs.FILE, p, nfs.SIZE, kit.FmtSize(int64(n)))
				return h.Header.Get(ContentType), h.Filename, p, kit.Format(n)
			}
		}
	}
	return "", "", "", "0"
}
func _cache_download(m *ice.Message, r *http.Response) (file, size string) {
	defer r.Body.Close()

	if f, p, e := miss.CreateFile(path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ))); m.Assert(e) {
		defer f.Close()

		step, total := 0, kit.Int(kit.Select("100", r.Header.Get(ContentLength)))
		size, buf := 0, make([]byte, ice.MOD_BUFS)

		for {
			if n, e := r.Body.Read(buf); n > 0 && e == nil {
				size += n
				f.Write(buf[0:n])
				s := size * 100 / total

				switch cb := m.OptionCB(SPIDE).(type) {
				case func(int, int, int):
					cb(size, total, s)
				case func(int, int):
					cb(size, total)
				default:
					if s != step && s%10 == 0 {
						m.Log_IMPORT(nfs.FILE, p, mdb.VALUE, s, mdb.COUNT, kit.FmtSize(int64(size)), mdb.TOTAL, kit.FmtSize(int64(total)))
					}
				}

				step = s
				continue
			}
			return p, kit.Format(size)
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
	Index.MergeCommands(ice.Commands{
		CACHE: {Name: "cache hash auto", Help: "缓存池", Actions: ice.MergeAction(ice.Actions{
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
					file, size = _cache_catch(m, file)
					_cache_save(m, arg[0], arg[1], "", file, size)
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,size,type,name,text,file")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				return
			}
			if m.Append(nfs.FILE) == "" {
				m.PushScript("inner", m.Append(mdb.TEXT))
			} else {
				m.PushDownload(m.Append(mdb.NAME), m.MergeURL2(SHARE_CACHE+arg[0]))
			}
		}},
		PP(CACHE): {Name: "/cache/", Help: "缓存池", Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelectDetail(m, arg[0], func(value ice.Map) {
				if kit.Format(value[nfs.FILE]) == "" {
					m.RenderResult(value[mdb.TEXT])
				} else {
					m.RenderDownload(value[nfs.FILE])
				}
			})
		}},
	})
}
