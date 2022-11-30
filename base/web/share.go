package web

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
	"shylinux.com/x/toolkits/logs"
)

func _share_render(m *ice.Message, arg ...string) {
	ice.AddRender(ice.RENDER_DOWNLOAD, func(msg *ice.Message, arg ...ice.Any) string {
		args := kit.Simple(arg...)
		list := []string{ice.POD, msg.Option(ice.MSG_USERPOD), "filename", kit.Select("", args[0], len(args) > 1)}
		return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, _share_link(msg, kit.Select(args[0], args, 1), list), path.Base(args[0]), args[0])
	})
}
func _share_link(m *ice.Message, p string, arg ...ice.Any) string {
	p = kit.Select("", SHARE_LOCAL, !strings.HasPrefix(p, ice.PS) && !strings.HasPrefix(p, ice.HTTP)) + p
	return tcp.PublishLocalhost(m, MergeLink(m, p, arg...))
}
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(ice.POD); ctx.PodCmd(m, CACHE, arg[0]) {
		if m.Append(nfs.FILE) == "" {
			m.RenderResult(m.Append(mdb.TEXT))
		} else {
			m.Option(ice.POD, pod)
			_share_local(m, m.Append(nfs.FILE))
		}
	} else {
		msg := m.Cmd(CACHE, arg[0])
		m.RenderDownload(msg.Append(nfs.FILE), msg.Append(mdb.TYPE), msg.Append(mdb.NAME))
	}
}
func _share_local(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, ice.PS); ls[0] {
	case ice.ETC, ice.VAR:
		if m.Warn(m.Option(ice.MSG_USERROLE) == aaa.VOID, ice.ErrNotRight, p) {
			return
		}
	default:
		if m.Option(ice.POD) == "" && !aaa.Right(m, ls) {
			return
		}
	}
	if m.Option(ice.POD) == "" {
		m.RenderDownload(p)
		return
	}
	pp := path.Join(ice.VAR_PROXY, m.Option(ice.POD), p)
	cache, size := time.Now().Add(-time.Hour*24), int64(0)
	if s, e := file.StatFile(pp); e == nil {
		cache, size = s.ModTime(), s.Size()
	}
	if p == ice.BIN_ICE_BIN {
		aaa.UserRoot(m).Cmd(SPACE, m.Option(ice.POD), SPIDE, "submit", MergeURL2(m, SHARE_PROXY, nfs.PATH, ""), m.Option(ice.POD), p, size, cache.Format(ice.MOD_TIME))
	} else {
		m.Cmd(SPACE, m.Option(ice.POD), SPIDE, ice.DEV, SPIDE_RAW, MergeURL2(m, SHARE_PROXY, nfs.PATH, ""),
			SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, p, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)
	}
	if !m.Warn(!file.ExistsFile(pp), ice.ErrNotFound, pp) {
		m.RenderDownload(pp)
	}
}
func _share_proxy(m *ice.Message) {
	switch p := path.Join(ice.VAR_PROXY, m.Option(ice.POD), m.Option(nfs.PATH)); m.R.Method {
	case http.MethodGet:
		m.RenderDownload(path.Join(p, m.Option(mdb.NAME)))
	case http.MethodPost:
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option(mdb.DATA), p)
		m.RenderResult(m.Option(nfs.PATH))
	}
}

const (
	TOPIC = "topic"
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
	FIELD = "field"

	SHARE_TOAST = "/share/toast/"
	SHARE_CACHE = "/share/cache/"
	SHARE_LOCAL = "/share/local/"
	SHARE_PROXY = "/share/proxy/"
	SHARE_REPOS = "/share/repos/"

	SHARE_LOCAL_AVATAR     = "/share/local/avatar/"
	SHARE_LOCAL_BACKGROUND = "/share/local/background/"
)
const SHARE = "share"

func init() {
	Index.MergeCommands(ice.Commands{
		SHARE: {Name: "share hash auto prunes", Help: "共享链", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { _share_render(m) }},
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, arg, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
				m.Option(mdb.LINK, _share_link(m, PP(SHARE)+m.Result()))
			}},
			LOGIN: {Name: "login userrole=void,tech username", Hand: func(m *ice.Message, arg ...string) {
				m.EchoQRCode(m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN).Option(mdb.LINK)).ProcessInner()
			}},
			SERVE_PARSE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Select("", arg, 0) != SHARE {
					return
				}
				switch arg[1] {
				case "local":
				default:
					m.Logs("Refer", arg[0], arg[1])
					m.Option(arg[0], arg[1])
				}
			}},
			SERVE_LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ice.MSG_USERNAME) == "" && m.Option(SHARE) != "" {
					switch msg := m.Cmd(SHARE, m.Option(SHARE)); msg.Append(mdb.TYPE) {
					case STORM, FIELD:
						msg.Tables(func(value ice.Maps) { aaa.SessAuth(m, value) })
					}
				}
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,username,usernick,userrole,river,storm,type,name,text", mdb.EXPIRE, "72h"), ServeAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if ctx.PodCmd(m, SHARE, arg) {
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				link := _share_link(m, P(SHARE, arg[0]))
				m.PushQRCode(cli.QRCODE, link)
				m.PushScript(nfs.SCRIPT, link)
				m.PushAnchor(link)
			} else {
				m.Action(LOGIN)
			}
		}},
		PP(SHARE): {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(SHARE, m.Option(SHARE, kit.Select(m.Option(SHARE), arg, 0)))
			if IsNotValidShare(m, msg.Append(mdb.TIME)) {
				m.RenderResult(kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s",
					msg.Append(aaa.USERNAME), msg.Append(aaa.USERNICK), msg.Append(mdb.TYPE), msg.Append(mdb.NAME)))
				return
			}
			switch msg.Append(mdb.TYPE) {
			case LOGIN:
				m.RenderRedirect(ice.PS, ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
			default:
				RenderIndex(m, ice.VOLCANOS)
			}
		}},
		SHARE_TOAST: {Name: "/share/toast/", Help: "推送流", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(SPACE, m.Option(ice.POD), m.Optionv("cmds"))
		}},
		SHARE_CACHE: {Name: "/share/cache/", Help: "缓存池", Hand: func(m *ice.Message, arg ...string) {
			_share_cache(m, arg...)
		}},
		SHARE_LOCAL: {Name: "/share/local/", Help: "文件夹", Hand: func(m *ice.Message, arg ...string) {
			_share_local(m, arg...)
		}},
		SHARE_LOCAL_AVATAR: {Name: "avatar", Help: "头像", Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.CmdAppend(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.AVATAR), SHARE_LOCAL))
		}},
		SHARE_LOCAL_BACKGROUND: {Name: "background", Help: "背景", Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.CmdAppend(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.BACKGROUND), SHARE_LOCAL))
		}},
		SHARE_PROXY: {Name: "/share/proxy/", Help: "文件流", Hand: func(m *ice.Message, arg ...string) {
			_share_proxy(m)
		}},
	})
}

func IsNotValidShare(m *ice.Message, time string) bool {
	return m.Warn(time < m.Time(), ice.ErrNotValid, m.Option(SHARE), time, m.Time(), logs.FileLineMeta(2))
}
