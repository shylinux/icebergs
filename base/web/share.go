package web

import (
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

func _share_link(m *ice.Message, p string) string {
	return tcp.PublishLocalhost(m, MergeLink(m, kit.Select("", SHARE_LOCAL, !strings.HasPrefix(p, ice.PS) && !strings.HasPrefix(p, HTTP))+p))
}
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(ice.POD); ctx.PodCmd(m, CACHE, arg[0]) {
		if m.Append(nfs.FILE) == "" {
			m.RenderResult(m.Append(mdb.TEXT))
		} else {
			m.Option(ice.POD, pod)
			ShareLocalFile(m, m.Append(nfs.FILE))
		}
	} else {
		if m.Cmdy(CACHE, arg[0]); m.Append(nfs.FILE) == "" {
			m.RenderResult(m.Append(mdb.TEXT))
		} else {
			m.RenderDownload(m.Append(nfs.FILE), m.Append(mdb.TYPE), m.Append(mdb.NAME))
		}
	}
}
func _share_proxy(m *ice.Message) {
	switch p := path.Join(ice.VAR_PROXY, m.Option(ice.POD), m.Option(nfs.PATH)); m.R.Method {
	case http.MethodGet:
		m.RenderDownload(p, m.Option(mdb.TYPE), m.Option(mdb.NAME))
	case http.MethodPost:
		if _, _, e := m.R.FormFile(UPLOAD); e == nil {
			m.Cmdy(CACHE, UPLOAD).Cmdy(CACHE, WATCH, m.Option(mdb.HASH), p)
		}
		m.RenderResult(m.Option(nfs.PATH))
	}
}

const (
	THEME = "tospic"
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
	FIELD = "field"

	SHARE_CACHE = "/share/cache/"
	SHARE_LOCAL = "/share/local/"
	SHARE_PROXY = "/share/proxy/"
	SHARE_TOAST = "/share/toast/"

	SHARE_LOCAL_AVATAR     = "/share/local/avatar/"
	SHARE_LOCAL_BACKGROUND = "/share/local/background/"
)
const SHARE = "share"

func init() {
	Index.MergeCommands(ice.Commands{
		SHARE: {Name: "share hash auto login prunes", Help: "共享链", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, arg, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
				m.Option(mdb.LINK, _share_link(m, P(SHARE, m.Result())))
			}},
			LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoQRCode(m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN).Option(mdb.LINK)).ProcessInner()
			}},
			ice.PS: {Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(len(arg) == 0 || arg[0] == "", ice.ErrNotValid, SHARE) {
					return
				}
				msg := m.Cmd(SHARE, m.Option(SHARE, arg[0]))
				if IsNotValidShare(m, msg.Append(mdb.TIME)) {
					m.RenderResult(kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s",
						msg.Append(aaa.USERNAME), msg.Append(aaa.USERNICK), msg.Append(mdb.TYPE), msg.Append(mdb.NAME)))
					return
				}
				switch msg.Append(mdb.TYPE) {
				case LOGIN:
					m.RenderRedirect(ice.PS, ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
				default:
					RenderMain(m)
				}
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,username,usernick,userrole,river,storm,type,name,text", mdb.EXPIRE, mdb.DAYS), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if ctx.PodCmd(m, SHARE, arg) {
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				link := _share_link(m, P(SHARE, arg[0]))
				m.PushQRCode(cli.QRCODE, link)
				m.PushScript(nfs.SCRIPT, link)
				m.PushAnchor(link)
			}
		}},
		SHARE_CACHE: {Hand: func(m *ice.Message, arg ...string) { _share_cache(m, arg...) }},
		SHARE_LOCAL: {Hand: func(m *ice.Message, arg ...string) { ShareLocalFile(m, arg...) }},
		SHARE_LOCAL_AVATAR: {Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmdv(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.AVATAR), SHARE_LOCAL))
		}},
		SHARE_LOCAL_BACKGROUND: {Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmdv(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.BACKGROUND), SHARE_LOCAL))
		}},
		SHARE_PROXY: {Hand: func(m *ice.Message, arg ...string) { _share_proxy(m) }},
		SHARE_TOAST: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPACE, arg[0], kit.UnMarshal(m.Option("arg"))) }},
	})
}
func IsNotValidShare(m *ice.Message, time string) bool {
	return m.Warn(time < m.Time(), ice.ErrNotValid, m.Option(SHARE), time, m.Time(), logs.FileLineMeta(2))
}
func ShareLocalFile(m *ice.Message, arg ...string) {
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
		m.Option(ice.MSG_USERROLE, aaa.TECH)
	}
	m.Cmd(SPACE, m.Option(ice.POD), SPIDE, ice.DEV, SPIDE_RAW, MergeLink(m, SHARE_PROXY), SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, p, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)
	if file.ExistsFile(pp) {
		m.RenderDownload(pp)
	} else {
		m.RenderDownload(p)
	}
}
