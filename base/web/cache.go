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
func _cache_type(m *ice.Message, kind, name string) string {
	if kind == "application/octet-stream" {
		if kit.ExtIsImage(name) {
			kind = "image/"+kit.Ext(name)
		} else if kit.ExtIsVideo(name) {
			kind = "video/"+kit.Ext(name)
		}
	}
	return kind
}
func _cache_save(m *ice.Message, kind, name, text string, arg ...string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	}
	if len(text) > 512 || kind == nfs.GO {
		p := m.Cmdx(nfs.SAVE, _cache_name(m, kit.Hashs(text)), text)
		text, arg = p, kit.Simple(p, len(text))
	}
	file, size := kit.Select("", arg, 0), kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	kind, text = _cache_type(m, kind, name), kit.Select(file, text)
	h := mdb.HashCreate(m, kit.SimpleKV("", kind, name, text), nfs.FILE, file, nfs.SIZE, size)
	m.Push(mdb.TIME, m.Time())
	m.Push(mdb.TYPE, kind)
	m.Push(mdb.NAME, name)
	m.Push(mdb.TEXT, text)
	m.Push(nfs.FILE, file)
	m.Push(nfs.SIZE, size)
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
	if msg := m.Cmd(nfs.DIR, name, "hash,size"); msg.Length() > 0 {
		return m.Cmdx(nfs.LINK, _cache_name(m, msg.Append(mdb.HASH)), name), msg.Append(nfs.SIZE)
	}
	return "", "0"
}
func _cache_upload(m *ice.Message, r *http.Request) (kind, name, file, size string) {
	if b, h, e := r.FormFile(UPLOAD); !m.Warn(e, ice.ErrNotValid, UPLOAD) {
		defer b.Close()
		if f, p, e := miss.CreateFile(_cache_name(m, kit.Hashs(b))); !m.Warn(e, ice.ErrNotValid, UPLOAD) {
			defer f.Close()
			b.Seek(0, os.SEEK_SET)
			if n, e := io.Copy(f, b); !m.Warn(e, ice.ErrNotValid, UPLOAD) {
				m.Logs(mdb.IMPORT, nfs.FILE, p, nfs.SIZE, kit.FmtSize(int64(n)))
				return h.Header.Get(ContentType), h.Filename, p, kit.Format(n)
			}
		}
	}
	return "", "", "", "0"
}
func _cache_download(m *ice.Message, r *http.Response, file string) string {
	defer r.Body.Close()
	if f, p, e := miss.CreateFile(file); m.Warn(e, ice.ErrNotValid, DOWNLOAD) {
		defer f.Close()
		nfs.CopyFile(m, f, r.Body, kit.Int(kit.Select("100", r.Header.Get(ContentLength))), m.OptionCB(SPIDE))
		return p
	}
	return ""
}

const (
	WATCH    = "watch"
	CATCH    = "catch"
	WRITE    = "write"
	UPLOAD   = "upload"
	DOWNLOAD = "download"
	DISPLAY  = "display"

	UPLOAD_WATCH = "upload_watch"
)
const CACHE = "cache"

func init() {
	Index.MergeCommands(ice.Commands{
		CACHE: {Name: "cache hash auto", Help: "缓存池", Actions: ice.MergeActions(ice.Actions{
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
					file, size := _cache_catch(m, _cache_download(m, r, path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ))))
					_cache_save(m, arg[0], arg[1], "", file, size)
				}
			}},
			UPLOAD_WATCH: {Name: "upload_watch", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				if len(up) < 2 {
					msg := m.Cmd(CACHE, UPLOAD)
					up = kit.Simple(msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
				}
				if p := path.Join(arg[0], up[1]); m.Option(ice.MSG_USERPOD) == "" {
					m.Cmdy(CACHE, WATCH, up[0], p)
				} else {
					m.Cmdy(SPIDE, ice.DEV, nfs.SAVE, p, SPIDE_GET, MergeURL2(m, path.Join(SHARE_CACHE, up[0])))
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,size,type,name,text,file")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				return
			}
			if m.Append(nfs.FILE) == "" {
				m.PushScript("inner", m.Append(mdb.TEXT))
			} else {
				m.PushDownload(m.Append(mdb.NAME), MergeURL2(m, SHARE_CACHE+arg[0]))
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
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		switch sub {
		case UPLOAD:
			if key == CACHE {
				break
			}
			hand := action.Hand
			action.Hand = func(m *ice.Message, arg ...string) {
				if len(kit.Simple(m.Optionv(ice.MSG_UPLOAD))) == 1 {
					m.Cmdy(CACHE, UPLOAD)
				}
				hand(m, arg...)
			}
		}
		return nil, nil
	})
}
