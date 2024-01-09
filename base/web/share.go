package web

import (
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
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
		return kit.MergeURL2(UserHost(m), strings.TrimPrefix(p, ice.USR), arg...)
	}
	return kit.MergeURL2(UserHost(m), kit.Select("", PP(SHARE, LOCAL), !strings.HasPrefix(p, nfs.PS) && !strings.HasPrefix(p, HTTP))+p, arg...)
}
func _share_cache(m *ice.Message, arg ...string) {
	if m.Cmdy(CACHE, arg[0]); m.Length() == 0 {
		if pod := m.Option(ice.POD); pod != "" {
			msg := m.Options(ice.POD, "", ice.MSG_USERROLE, aaa.TECH).Cmd(SPACE, pod, CACHE, arg[0])
			kit.If(kit.Format(msg.Append(nfs.FILE)), func() {
				m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, pod, msg.Append(nfs.FILE)))
			}, func() {
				m.RenderResult(msg.Append(mdb.TEXT))
			})
		}
	} else if m.Append(nfs.FILE) != "" {
		m.RenderDownload(m.Append(nfs.FILE), m.Append(mdb.TYPE), m.Append(mdb.NAME))
	} else {
		m.RenderResult(m.Append(mdb.TEXT))
	}
}
func _share_proxy(m *ice.Message) {
	if m.Warn(m.Option(SHARE) == "", ice.ErrNotValid) {
		return
	}
	msg := m.Cmd(SHARE, m.Option(SHARE))
	defer m.Cmd(SHARE, mdb.REMOVE, mdb.HASH, m.Option(SHARE))
	if m.Warn(msg.Append(mdb.TEXT) == "", ice.ErrNotValid) {
		return
	}
	p := path.Join(ice.VAR_PROXY, msg.Append(mdb.TEXT), msg.Append(mdb.NAME))
	if _, _, e := m.R.FormFile(UPLOAD); e == nil {
		m.Cmdy(CACHE, UPLOAD).Cmdy(CACHE, WATCH, m.Option(mdb.HASH), p)
	}
	m.RenderResult(m.Option(nfs.PATH))
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
	SHARE_CACHE = "/share/cache/"
)
const SHARE = "share"

func init() {
	Index.MergeCommands(ice.Commands{
		SHARE: {Name: "share hash auto login", Help: "共享链", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				// kit.If(m.Option(mdb.TYPE) == LOGIN, func() { arg = append(arg, mdb.TEXT, tcp.PublishLocalhost(m, m.Option(mdb.TEXT))) })
				kit.If(m.Option(mdb.TYPE) == LOGIN && m.Option(mdb.TEXT) == "", func() { arg = append(arg, mdb.TEXT, tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB))) })
				mdb.HashCreate(m, arg, SPACE, m.Option(ice.MSG_USERPOD), aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
				m.Option(mdb.LINK, tcp.PublishLocalhost(m, m.MergeLink(P(SHARE, m.Result()))))
			}},
			LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.EchoQRCode(m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN).Option(mdb.LINK)).ProcessInner()
			}},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn(), m.Option(SHARE)); !IsNotValidFieldShare(m, msg) {
					m.Cmdy(Space(m, msg.Append(SPACE)), ctx.COMMAND, msg.Append(mdb.NAME))
				}
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn(), m.Option(SHARE)); !IsNotValidFieldShare(m, msg) {
					aaa.SessAuth(m, kit.Dict(msg.AppendSimple(aaa.USERNICK, aaa.USERNAME, aaa.USERROLE)))
					m.Cmdy(Space(m, msg.Append(SPACE)), msg.Append(mdb.NAME), arg[1:])
				}
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
					RenderCookie(m, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
					m.RenderRedirect(msg.Append(mdb.TEXT))
				case STORM:
					RenderCookie(m, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
					m.RenderRedirect(m.MergeLink(kit.Select(nfs.PS, msg.Append(mdb.TEXT)), msg.AppendSimple(RIVER, STORM)))
				case FIELD:
					RenderPodCmd(m, msg.Append(SPACE), msg.Append(mdb.NAME), kit.UnMarshal(msg.Append(mdb.TEXT)))
				case DOWNLOAD:
					m.RenderDownload(msg.Append(mdb.TEXT))
				default:
					RenderMain(m)
				}
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,usernick,username,userrole", mdb.EXPIRE, mdb.DAYS)), Hand: func(m *ice.Message, arg ...string) {
			if kit.IsIn(m.Option(ice.MSG_USERROLE), aaa.ROOT, aaa.TECH) || len(arg) > 0 && arg[0] != "" {
				mdb.HashSelect(m, arg...)
			}
		}},
		PP(SHARE, CACHE): {Hand: func(m *ice.Message, arg ...string) { _share_cache(m, arg...) }},
		PP(SHARE, LOCAL): {Hand: func(m *ice.Message, arg ...string) { ShareLocalFile(m, arg...) }},
		PP(SHARE, PROXY): {Hand: func(m *ice.Message, arg ...string) { _share_proxy(m) }},
		PP(SHARE, TOAST): {Hand: func(m *ice.Message, arg ...string) {
			m.Options(ice.LOG_DISABLE, ice.TRUE).Cmdy(SPACE, arg[0], kit.UnMarshal(m.Option(ice.ARG)))
		}},
	})
}
func IsNotValidShare(m *ice.Message, time string) bool {
	return m.Warn(time < m.Time(), ice.ErrNotValid, m.Option(SHARE), time, m.Time(), logs.FileLineMeta(2))
}
func IsNotValidFieldShare(m *ice.Message, msg *ice.Message) bool {
	if m.Warn(IsNotValidShare(m, msg.Append(mdb.TIME)), kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s %s", msg.Append(aaa.USERNICK), msg.Append(aaa.USERNAME), msg.Append(mdb.TYPE), msg.Append(mdb.NAME), msg.Append(mdb.TEXT))) {
		return true
	}
	if m.Warn(msg.Append(mdb.NAME) == "") {
		return true
	}
	return false
}
func SharePath(m *ice.Message, p string) string {
	kit.If(!kit.HasPrefix(p, nfs.PS, ice.HTTP), func() {
		if kit.HasPrefix(p, nfs.SRC, nfs.USR) && !kit.HasPrefix(p, nfs.USR_LOCAL) {
			p = m.MergeLink(path.Join(nfs.REQUIRE, p), ice.POD, m.Option(ice.MSG_USERPOD))
		} else {
			p = m.MergeLink(path.Join(SHARE_LOCAL, p), ice.POD, m.Option(ice.MSG_USERPOD))
		}
	})
	return p
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
	} else if pp := kit.Path(ice.USR_LOCAL_WORK, m.Option(ice.POD), p); nfs.Exists(m, pp) {
		m.RenderDownload(pp)
	} else if pp := kit.Path(ice.USR_LOCAL_WORK, m.Option(ice.POD)); nfs.Exists(m, pp) {
		m.RenderDownload(p)
	} else {
		m.RenderDownload(ProxyUpload(m, m.Option(ice.POD), p))
	}
}
func ProxyUpload(m *ice.Message, pod string, p string) string {
	pp := path.Join(ice.VAR_PROXY, pod, p)
	size, cache := int64(0), time.Now().Add(-time.Hour*24)
	if s, e := file.StatFile(pp); e == nil {
		size, cache = s.Size(), s.ModTime()
	} else if s, e := file.StatFile(p); e == nil {
		size, cache = s.Size(), s.ModTime()
	}
	kit.If(p == ice.BIN_ICE_BIN, func() { m.Option(ice.MSG_USERROLE, aaa.TECH) })
	share := m.Cmdx(SHARE, mdb.CREATE, mdb.TYPE, PROXY, mdb.NAME, p, mdb.TEXT, pod)
	defer m.Cmd(SHARE, mdb.REMOVE, mdb.HASH, share)
	url := tcp.PublishLocalhost(m, m.MergeLink(PP(SHARE, PROXY), SHARE, share))
	m.Cmd(SPACE, pod, SPIDE, PROXY, URL, url, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, mdb.AT+p)
	return kit.Select(p, pp, file.ExistsFile(pp))
}
func ShareLocal(m *ice.Message, p string) string {
	if kit.HasPrefix(p, nfs.PS, HTTP) {
		return p
	}
	return m.MergeLink(PP(SHARE, LOCAL, p))
}
