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

func _share_link(m *ice.Message, p string, arg ...ice.Any) string {
	if strings.HasPrefix(p, ice.USR_PUBLISH) {
		return tcp.PublishLocalhost(m, MergeLink(m, strings.TrimPrefix(p, ice.USR), arg...))
	}
	return tcp.PublishLocalhost(m, MergeLink(m, kit.Select("", PP(SHARE, LOCAL), !strings.HasPrefix(p, nfs.PS) && !strings.HasPrefix(p, HTTP))+p, arg...))
}
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(ice.POD); ctx.PodCmd(m, CACHE, arg[0]) {
		if m.Append(nfs.FILE) == "" {
			m.RenderResult(m.Append(mdb.TEXT))
		} else {
			ShareLocalFile(m.Options(ice.POD, pod), m.Append(nfs.FILE))
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
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
	FIELD = "field"

	LOCAL = "local"
	PROXY = "proxy"
	TOAST = "toast"

	SHARE_LOCAL = "/share/local/"
)
const SHARE = "share"

func init() {
	Index.MergeCommands(ice.Commands{
		SHARE: {Name: "share hash auto login prunes", Help: "共享链", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, arg, aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
				m.Option(mdb.LINK, _share_link(m, P(SHARE, m.Result())))
			}},
			LOGIN: {Hand: func(m *ice.Message, arg ...string) {
				m.EchoQRCode(m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN).Option(mdb.LINK)).ProcessInner()
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) {
				if m.Warn(len(arg) == 0 || arg[0] == "", ice.ErrNotValid, SHARE) {
					return
				}
				msg := m.Cmd(SHARE, m.Option(SHARE, arg[0]))
				if IsNotValidShare(m, msg.Append(mdb.TIME)) {
					m.RenderResult(kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s", msg.Append(aaa.USERNICK), msg.Append(aaa.USERNAME), msg.Append(mdb.TYPE), msg.Append(mdb.NAME)))
					return
				}
				switch msg.Append(mdb.TYPE) {
				case LOGIN:
					m.RenderRedirect(nfs.PS, ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
				default:
					RenderMain(m)
				}
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,river,storm,usernick,username,userrole", mdb.EXPIRE, mdb.DAYS), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
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
		PP(SHARE, CACHE): {Hand: func(m *ice.Message, arg ...string) { _share_cache(m, arg...) }},
		PP(SHARE, LOCAL): {Hand: func(m *ice.Message, arg ...string) { ShareLocalFile(m, arg...) }},
		PP(SHARE, LOCAL, aaa.AVATAR): {Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmdv(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.AVATAR), PP(SHARE, LOCAL)))
		}},
		PP(SHARE, LOCAL, aaa.BACKGROUND): {Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmdv(aaa.USER, m.Option(ice.MSG_USERNAME), aaa.BACKGROUND), PP(SHARE, LOCAL)))
		}},
		PP(SHARE, PROXY): {Hand: func(m *ice.Message, arg ...string) { _share_proxy(m) }},
		PP(SHARE, TOAST): {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPACE, arg[0], kit.UnMarshal(m.Option(ice.ARG))) }},
	})
}
func IsNotValidShare(m *ice.Message, time string) bool {
	return m.Warn(time < m.Time(), ice.ErrNotValid, m.Option(SHARE), time, m.Time(), logs.FileLineMeta(2))
}
func ShareLocalFile(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, nfs.PS); ls[0] {
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
	} else if p := kit.Path(ice.USR_LOCAL_WORK, m.Option(ice.POD), p); nfs.Exists(m, p) {
		m.RenderDownload(p)
		return
	}
	pp := path.Join(ice.VAR_PROXY, m.Option(ice.POD), p)
	cache, size := time.Now().Add(-time.Hour*24), int64(0)
	if s, e := file.StatFile(pp); e == nil {
		cache, size = s.ModTime(), s.Size()
	}
	kit.If(p == ice.BIN_ICE_BIN, func() { m.Option(ice.MSG_USERROLE, aaa.TECH) })
	m.Cmd(SPACE, m.Option(ice.POD), SPIDE, ice.DEV, SPIDE_RAW, MergeLink(m, PP(SHARE, PROXY)), SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, p, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, mdb.AT+p)
	m.RenderDownload(kit.Select(p, pp, file.ExistsFile(pp)))
}
