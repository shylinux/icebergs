package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _cache_name(m *ice.Message, h string) string { return path.Join(ice.VAR_FILE, h[:2], h) }
func _cache_mime(m *ice.Message, mime, name string) string {
	if mime == "application/octet-stream" {
		if kit.ExtIsImage(name) {
			mime = "image/" + kit.Ext(name)
		} else if kit.ExtIsVideo(name) {
			mime = "video/" + kit.Ext(name)
		}
	} else if mime == "" {
		return kit.Ext(name)
	}
	return mime
}
func _cache_save(m *ice.Message, mime, name, text string, arg ...string) {
	if m.Warn(name == "", ice.ErrNotValid, mdb.NAME) {
		return
	} else if len(text) > 512 {
		p := m.Cmdx(nfs.SAVE, _cache_name(m, kit.Hashs(text)), text)
		text, arg = p, kit.Simple(p, len(text))
	}
	file, size := kit.Select("", arg, 0), kit.Int(kit.Select(kit.Format(len(text)), arg, 1))
	mime, text = _cache_mime(m, mime, name), kit.Select(file, text)
	m.Push(mdb.TIME, m.Time()).Push(mdb.HASH, mdb.HashCreate(m.Spawn(), kit.SimpleKV("", mime, name, text), nfs.FILE, file, nfs.SIZE, size))
	m.Push(mdb.TYPE, mime).Push(mdb.NAME, name).Push(mdb.TEXT, text).Push(nfs.FILE, file).Push(nfs.SIZE, size)
}
func _cache_watch(m *ice.Message, key, path string) {
	mdb.HashSelect(m.Spawn(), key).Tables(func(value ice.Maps) {
		if value[nfs.FILE] == "" {
			m.Cmdy(nfs.SAVE, path, value[mdb.TEXT])
		} else {
			m.Cmdy(nfs.LINK, path, value[nfs.FILE])
		}
	})
}
func _cache_catch(m *ice.Message, path string) (file string, size string) {
	if msg := m.Cmd(nfs.DIR, path, "hash,size"); msg.Length() > 0 {
		return m.Cmdx(nfs.LINK, _cache_name(m, msg.Append(mdb.HASH)), path), msg.Append(nfs.SIZE)
	}
	return "", "0"
}
func _cache_upload(m *ice.Message, r *http.Request) (mime, name, file, size string) {
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
func _cache_download(m *ice.Message, r *http.Response, file string, cb ice.Any) string {
	if f, p, e := nfs.CreateFile(m, file); m.Assert(e) {
		// if f, p, e := miss.CreateFile(file); !m.Warn(e, ice.ErrNotValid, DOWNLOAD) {
		defer f.Close()
		last, base := 0, 10
		nfs.CopyFile(m, f, r.Body, base*ice.MOD_BUFS, kit.Int(kit.Select("100", r.Header.Get(ContentLength))), func(count, total, step int) {
			if step/base != last {
				m.Logs(mdb.EXPORT, nfs.FILE, p, mdb.COUNT, count, mdb.TOTAL, total, mdb.VALUE, step)
				switch cb := cb.(type) {
				case func(int, int, int):
					cb(count, total, step)
				case nil:
				default:
					m.ErrorNotImplement(cb)
				}
			}
			last = step / base
		})
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
)
const CACHE = "cache"

func init() {
	Index.MergeCommands(ice.Commands{
		CACHE: {Name: "cache hash auto write catch upload", Help: "缓存池", Actions: ice.MergeActions(ice.Actions{
			WATCH: {Name: "watch hash* path*", Help: "释放", Hand: func(m *ice.Message, arg ...string) {
				_cache_watch(m, m.Option(mdb.HASH), m.Option(nfs.PATH))
			}},
			WRITE: {Name: "write type name* text*", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_cache_save(m, m.Option(mdb.TYPE), m.Option(mdb.NAME), m.Option(mdb.TEXT))
			}},
			CATCH: {Name: "catch path* type", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				file, size := _cache_catch(m, m.Option(nfs.PATH))
				_cache_save(m, m.Option(mdb.TYPE), m.Option(nfs.PATH), "", file, size)
			}},
			UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				mime, name, file, size := _cache_upload(m, m.R)
				_cache_save(m, mime, name, "", file, size)
			}},
			DOWNLOAD: {Name: "download type name*", Hand: func(m *ice.Message, arg ...string) {
				if r, ok := m.Optionv(RESPONSE).(*http.Response); !m.Warn(!ok, ice.ErrNotValid, RESPONSE) {
					file, size := _cache_catch(m, _cache_download(m, r, path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ)), m.OptionCB("")))
					_cache_save(m, m.Option(mdb.TYPE), m.Option(mdb.NAME), "", file, size)
				}
			}},
			ice.RENDER_DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				p := kit.Select(arg[0], arg, 1)
				p = kit.Select("", SHARE_LOCAL, !strings.HasPrefix(p, ice.PS) && !strings.HasPrefix(p, ice.HTTP)) + p
				args := []string{ice.POD, m.Option(ice.MSG_USERPOD), "filename", kit.Select("", arg[0], len(arg) > 1)}
				m.Echo(fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, MergeURL2(m, p, args), path.Base(arg[0]), arg[0]))
			}},
			ice.PS: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectDetail(m, arg[0], func(value ice.Map) {
					if kit.Format(value[nfs.FILE]) == "" {
						m.RenderResult(value[mdb.TEXT])
					} else {
						m.RenderDownload(value[nfs.FILE])
					}
				})
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,size,type,name,text,file", ctx.ACTION, WATCH), ice.RenderAction(ice.RENDER_DOWNLOAD)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || m.R.Method == http.MethodGet {
				return
			}
			if m.Append(nfs.FILE) == "" {
				m.PushScript(mdb.TEXT, m.Append(mdb.TEXT))
			} else {
				PushDisplay(m, m.Append(mdb.TYPE), m.Append(mdb.NAME), MergeURL2(m, SHARE_CACHE+arg[0]))
			}
		}},
	})
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		switch sub {
		case UPLOAD:
			if c.Name == WEB && key == CACHE {
				break
			}
			watch := action.Hand == nil
			action.Hand = ice.MergeHand(func(m *ice.Message, arg ...string) {
				up := Upload(m)
				m.Assert(len(up) > 1)
				m.Cmd(CACHE, m.Option(ice.MSG_UPLOAD)).Tables(func(value ice.Maps) { m.Options(value) })
				if m.Options(mdb.HASH, up[0], mdb.NAME, up[1]); watch {
					m.Cmdy(CACHE, WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH), m.Option(mdb.NAME)))
				}
			}, action.Hand)
		}
		return nil, nil
	})
	ctx.Upload = Upload
}
func RenderCache(m *ice.Message, h string) {
	if msg := m.Cmd(CACHE, h); msg.Append(nfs.FILE) == "" {
		m.RenderResult(msg.Append(mdb.TEXT))
	} else {
		m.RenderDownload(msg.Append(mdb.FILE), msg.Append(mdb.TYPE), msg.Append(mdb.NAME))
	}
}
func Upload(m *ice.Message) []string {
	if up := kit.Simple(m.Optionv(ice.MSG_UPLOAD)); len(up) == 1 {
		if m.Cmdy(CACHE, UPLOAD).Optionv(ice.MSG_UPLOAD, kit.Simple(m.Append(mdb.HASH), m.Append(mdb.NAME), m.Append(nfs.SIZE))); m.Option(ice.MSG_USERPOD) != "" {
			m.Cmd(SPACE, m.Option(ice.MSG_USERPOD), SPIDE, ice.DEV, SPIDE_CACHE, http.MethodGet, tcp.PublishLocalhost(m, MergeURL2(m, path.Join(SHARE_CACHE, m.Append(mdb.HASH)))))
		}
		return kit.Simple(m.Optionv(ice.MSG_UPLOAD))
	} else {
		return up
	}
}
func Download(m *ice.Message, link string, cb func(count, total, value int)) *ice.Message {
	return m.Cmdy("web.spide", ice.DEV, SPIDE_CACHE, http.MethodGet, link, cb)
}
func PushDisplay(m *ice.Message, mime, name, link string) {
	if strings.HasPrefix(mime, "image/") || kit.ExtIsImage(name) {
		m.PushImages(nfs.FILE, link)
	} else if strings.HasPrefix(mime, "video/") || kit.ExtIsImage(name) {
		m.PushVideos(nfs.FILE, link)
	} else {
		m.PushDownload(nfs.FILE, name, link)
	}
}
