package web

import (
	"io"
	"net/http"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _cache_name(m *ice.Message, h string) string { return path.Join(ice.VAR_FILE, h[:2], h) }
func _cache_mime(m *ice.Message, mime, name string) string {
	if mime == ApplicationOctet {
		if kit.ExtIsImage(name) {
			mime = IMAGE + nfs.PS + kit.Ext(name)
		} else if kit.ExtIsVideo(name) {
			mime = VIDEO + nfs.PS + kit.Ext(name)
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
	mdb.HashSelect(m.Spawn(), key).Table(func(value ice.Maps) {
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
				m.Logs(nfs.SAVE, nfs.FILE, p, nfs.SIZE, kit.FmtSize(int64(n)))
				return h.Header.Get(ContentType), h.Filename, p, kit.Format(n)
			}
		}
	}
	return "", "", "", "0"
}
func _cache_download(m *ice.Message, r *http.Response, file string, cb ice.Any) string {
	if f, p, e := miss.CreateFile(file); !m.Warn(e, ice.ErrNotValid, DOWNLOAD) {
		defer func() {
			if s, e := os.Stat(file); e == nil && s.Size() == 0 {
				nfs.Remove(m, file)
			}
		}()
		defer f.Close()
		last, base := 0, 10
		nfs.CopyStream(m, f, r.Body, base*ice.MOD_BUFS, kit.Int(kit.Select("100", r.Header.Get(ContentLength))), func(count, total, value int) {
			if value/base == last {
				return
			}
			last = value / base
			switch m.Logs(nfs.SAVE, nfs.FILE, p, mdb.COUNT, kit.FmtSize(int64(count)), mdb.TOTAL, kit.FmtSize(int64(total)), mdb.VALUE, value); cb := cb.(type) {
			case func(int, int, int):
				kit.If(cb != nil, func() { cb(count, total, value) })
			case nil:
			default:
				m.ErrorNotImplement(cb)
			}
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

	IMAGE = "image"
	VIDEO = "video"
)
const CACHE = "cache"

func init() {
	Index.MergeCommands(ice.Commands{
		CACHE: {Name: "cache hash auto write catch upload", Help: "缓存池", Actions: ice.MergeActions(ice.Actions{
			ice.RENDER_DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(_share_link(m, kit.Select(arg[0], arg, 1), ice.POD, m.Option(ice.MSG_USERPOD), nfs.FILENAME, kit.Select("", arg[0], len(arg) > 1)))
			}},
			WATCH: {Name: "watch hash* path*", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				_cache_watch(m, m.Option(mdb.HASH), m.Option(nfs.PATH))
			}},
			WRITE: {Name: "write type name* text*", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_cache_save(m, m.Option(mdb.TYPE), m.Option(mdb.NAME), m.Option(mdb.TEXT))
			}},
			CATCH: {Name: "catch path* type", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				file, size := _cache_catch(m, m.Option(nfs.PATH))
				_cache_save(m, m.Option(mdb.TYPE), m.Option(nfs.PATH), "", file, size)
			}},
			UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				mime, name, file, size := _cache_upload(m, m.R)
				_cache_save(m, mime, name, "", file, size)
			}},
			DOWNLOAD: {Name: "download type name*", Hand: func(m *ice.Message, arg ...string) {
				if res, ok := m.Optionv(RESPONSE).(*http.Response); !m.Warn(!ok, ice.ErrNotValid, RESPONSE) {
					p := path.Join(ice.VAR_TMP, kit.Hashs(mdb.UNIQ))
					defer os.Remove(p)
					file, size := _cache_catch(m, _cache_download(m, res, p, m.OptionCB("")))
					_cache_save(m, m.Option(mdb.TYPE), m.Option(mdb.NAME), "", file, size)
				}
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.HashSelectDetail(m, arg[0], func(value ice.Map) {
					kit.If(kit.Format(value[nfs.FILE]), func() { m.RenderDownload(value[nfs.FILE]) }, func() { m.RenderResult(value[mdb.TEXT]) })
				}) {
					return
				}
				if pod := m.Option(ice.POD); pod != "" {
					msg := m.Options(ice.POD, "").Cmd(SPACE, pod, CACHE, arg[0])
					kit.If(kit.Format(msg.Append(nfs.FILE)), func() {
						m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, pod, msg.Append(nfs.FILE)))
					}, func() { m.RenderResult(msg.Append(mdb.TEXT)) })
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,size,type,name,text,file", ctx.ACTION, WATCH), ice.RenderAction(ice.RENDER_DOWNLOAD)), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || m.R != nil && m.R.Method == http.MethodGet {
				m.Option(ice.MSG_ACTION, "")
			} else if m.Append(nfs.FILE) == "" {
				m.PushScript(mdb.TEXT, m.Append(mdb.TEXT))
			} else {
				PushDisplay(m, m.Append(mdb.TYPE), m.Append(mdb.NAME), MergeURL2(m, P(SHARE, CACHE, arg[0])))
			}
		}},
	})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) {
		switch sub {
		case UPLOAD:
			if kit.FileLines(action.Hand) == kit.FileLines(1) {
				break
			}
			watch := action.Hand == nil
			action.Hand = ice.MergeHand(func(m *ice.Message, arg ...string) {
				up := Upload(m)
				m.Assert(len(up) > 1)
				m.Cmd(CACHE, m.Option(ice.MSG_UPLOAD)).Table(func(value ice.Maps) { m.Options(value) })
				if m.Options(mdb.HASH, up[0], mdb.NAME, up[1]); watch {
					m.Cmdy(CACHE, WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH), up[1]))
				}
			}, action.Hand)
		}
	})
}
func Upload(m *ice.Message) []string {
	if up := kit.Simple(m.Optionv(ice.MSG_UPLOAD)); len(up) == 1 {
		if m.Cmdy(CACHE, UPLOAD).Optionv(ice.MSG_UPLOAD, kit.Simple(m.Append(mdb.HASH), m.Append(mdb.NAME), m.Append(nfs.SIZE))); m.Option(ice.MSG_USERPOD) != "" {
			m.Cmd(SPACE, m.Option(ice.MSG_USERPOD), SPIDE, ice.DEV, SPIDE_CACHE, http.MethodGet, tcp.PublishLocalhost(m, MergeURL2(m, PP(SHARE, CACHE, m.Append(mdb.HASH)))))
		}
		return kit.Simple(m.Optionv(ice.MSG_UPLOAD))
	} else {
		return up
	}
}
func Download(m *ice.Message, link string, cb func(count, total, value int)) *ice.Message {
	return m.Cmdy(Prefix(SPIDE), ice.DEV, SPIDE_CACHE, http.MethodGet, link, cb)
}
func PushDisplay(m *ice.Message, mime, name, link string) {
	if html.IsImage(name, mime) {
		m.PushImages(nfs.FILE, link)
	} else if html.IsVideo(name, mime) {
		m.PushVideos(nfs.FILE, link)
	} else if html.IsAudio(name, mime) {
		m.PushAudios(nfs.FILE, link)
	} else {
		m.PushDownload(nfs.FILE, name, link)
	}
}
func RenderCache(m *ice.Message, h string) {
	if msg := m.Cmd(CACHE, h); msg.Append(nfs.FILE) == "" {
		m.RenderResult(msg.Append(mdb.TEXT))
	} else {
		m.RenderDownload(msg.Append(mdb.FILE), msg.Append(mdb.TYPE), msg.Append(mdb.NAME))
	}
}
