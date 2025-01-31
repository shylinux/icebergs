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
	if m.WarnNotValid(m.Option(SHARE) == "") {
		return
	}
	msg := m.Cmd(SHARE, m.Option(SHARE))
	defer m.Cmd(SHARE, mdb.REMOVE, mdb.HASH, m.Option(SHARE))
	if m.WarnNotValid(msg.Append(mdb.TEXT) == "") {
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

	PROXY = "proxy"
	LOCAL = "local"

	SHARE_CACHE = "/share/cache/"
	SHARE_LOCAL = "/share/local/"
)
const SHARE = "share"

func init() {
	Index.MergeCommands(ice.Commands{
		SHARE: {Name: "share hash auto login", Help: "共享链", Icon: "Freeform.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text space", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == LOGIN && m.Option(mdb.TEXT) == "", func() { arg = append(arg, mdb.TEXT, tcp.PublishLocalhost(m, m.Option(ice.MSG_USERWEB))) })
				mdb.HashCreate(m, m.OptionSimple("type,name,text,space"), arg, aaa.USERNICK, m.Option(ice.MSG_USERNICK), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERROLE, m.Option(ice.MSG_USERROLE))
				m.Option(mdb.LINK, tcp.PublishLocalhost(m, m.MergeLink(P(SHARE, m.Result()))))
				Count(m, "", m.Option(mdb.TYPE))
			}},
			LOGIN: {Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				m.EchoQRCode(m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN).Option(mdb.LINK)).ProcessInner()
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergeLink(P(SHARE, m.Option(mdb.HASH))))
			}},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn(), m.Option(SHARE)); !IsNotValidFieldShare(m, msg) {
					m.Cmdy(Space(m, msg.Append(SPACE)), ctx.COMMAND, msg.Append(mdb.NAME), kit.Dict(ice.MSG_USERPOD, msg.Append(SPACE)))
				}
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.HashSelects(m.Spawn(), m.Option(SHARE)); !IsNotValidFieldShare(m, msg) {
					aaa.SessAuth(m, kit.Dict(msg.AppendSimple(aaa.USERNICK, aaa.USERNAME, aaa.USERROLE)))
					m.Cmdy(Space(m, msg.Append(SPACE)), msg.Append(mdb.NAME), kit.UnMarshal(msg.Append(mdb.TEXT)), arg[1:], kit.Dict(ice.MSG_USERPOD, msg.Append(SPACE)))
				}
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) {
				if m.WarnNotValid(len(arg) == 0 || arg[0] == "", SHARE) {
					return
				}
				msg := m.Cmd(SHARE, m.Option(SHARE, arg[0]))
				if IsNotValidShare(m, msg.Append(mdb.TIME)) {
					m.RenderResult(kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s", msg.Append(aaa.USERNICK), msg.Append(aaa.USERNAME), msg.Append(mdb.TYPE), msg.Append(mdb.NAME)))
					return
				}
				switch msg.Append(mdb.TYPE) {
				case LOGIN:
					u := kit.ParseURL(m.Option(ice.MSG_USERHOST))
					m.RenderRedirect(kit.MergeURL(msg.Append(mdb.TEXT), ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME))))
					break
					if u.Scheme == ice.HTTP {
						m.RenderRedirect(kit.MergeURL(msg.Append(mdb.TEXT), ice.MSG_SESSID, aaa.SessCreate(m, msg.Append(aaa.USERNAME))))
					} else {
						RenderCookie(m, aaa.SessCreate(m, msg.Append(aaa.USERNAME)))
						m.RenderRedirect(msg.Append(mdb.TEXT))
					}
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
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,space,usernick,username,userrole", mdb.EXPIRE, mdb.DAYS)), Hand: func(m *ice.Message, arg ...string) {
			if aaa.IsTechOrRoot(m) || len(arg) > 0 && arg[0] != "" {
				mdb.HashSelect(m, arg...).PushAction(OPEN, mdb.REMOVE)
			}
		}},
		PP(SHARE, PROXY): {Hand: func(m *ice.Message, arg ...string) { _share_proxy(m) }},
		PP(SHARE, CACHE): {Hand: func(m *ice.Message, arg ...string) { _share_cache(m, arg...) }},
		PP(SHARE, LOCAL): {Hand: func(m *ice.Message, arg ...string) { ShareLocalFile(m, arg...) }},
	})
}
func IsNotValidShare(m *ice.Message, time string) bool {
	return m.WarnNotValid(time < m.Time(), ice.ErrNotValid, m.Option(SHARE), time, m.Time(), logs.FileLineMeta(2))
}
func IsNotValidFieldShare(m *ice.Message, msg *ice.Message) bool {
	if m.Warn(IsNotValidShare(m, msg.Append(mdb.TIME)), kit.Format("共享超时, 请联系 %s(%s), 重新分享 %s %s %s", msg.Append(aaa.USERNICK), msg.Append(aaa.USERNAME), msg.Append(mdb.TYPE), msg.Append(mdb.NAME), msg.Append(mdb.TEXT))) {
		return true
	}
	if m.WarnNotValid(msg.Append(mdb.NAME) == "") {
		return true
	}
	return false
}
func SharePath(m *ice.Message, p string) string {
	kit.If(!kit.HasPrefix(p, nfs.PS, ice.HTTP), func() {
		if kit.HasPrefix(p, nfs.SRC, nfs.USR) && !kit.HasPrefix(p, nfs.USR_LOCAL) {
			p = m.MergeLink(path.Join(nfs.P, p), ice.POD, m.Option(ice.MSG_USERPOD))
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
		if m.WarnNotRight(m.Option(ice.MSG_USERROLE) == aaa.VOID, p) {
			return
		}
	default:
		if m.Option(ice.MSG_USERNAME) != "" && strings.HasPrefix(p, nfs.USR_LOCAL_IMAGE+m.Option(ice.MSG_USERNAME)) {

		} else if m.Option(ice.POD) == "" && !aaa.Right(m, ls) {
			return
		} else {
			if m.Option(ice.POD) != "" && !strings.Contains(p, "/src/") && !strings.HasPrefix(p, "src/") {
				if strings.HasPrefix(p, "usr/local/storage/") {
					if m.Cmd(SPACE, "20240903-operation", "web.team.storage.file", aaa.RIGHT, ls[3:]).IsErr() {
						return
					}
				} else if m.WarnNotRight(m.Cmdx(SPACE, m.Option(ice.POD), aaa.ROLE, aaa.RIGHT, aaa.VOID, p) != ice.OK) {
					return
				}
			}
		}
	}
	if m.Option(ice.POD) != "" && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(ice.POD))) {
		if pp := kit.Path(ice.USR_LOCAL_WORK, m.Option(ice.POD), p); nfs.Exists(m, pp) {
			m.RenderDownload(pp)
			return
		} else if nfs.Exists(m, p) {
			m.RenderDownload(p)
			return
		}
	}
	if m.Option(ice.POD) == "" || (kit.HasPrefix(p, ice.USR_ICONS, ice.USR_VOLCANOS, ice.USR_ICEBERGS, ice.USR_INTSHELL) && nfs.Exists(m, p)) {
		m.RenderDownload(p)
	} else if pp := kit.Path(ice.USR_LOCAL_WORK, m.Option(ice.POD), p); nfs.Exists(m, pp) {
		m.RenderDownload(pp)
	} else if pp := ProxyUpload(m, m.Option(ice.POD), p); nfs.Exists(m, pp) {
		m.RenderDownload(pp)
	} else {
		m.RenderDownload(p)
	}
}
func ShareLocal(m *ice.Message, p string) string {
	if kit.HasPrefix(p, nfs.PS, HTTP) {
		return p
	}
	return m.MergeLink(PP(SHARE, LOCAL, p))
}
func ShareField(m *ice.Message, cmd string, arg ...ice.Any) *ice.Message {
	return m.EchoQRCode(tcp.PublishLocalhost(m, m.MergeLink(P(SHARE, AdminCmd(m, SHARE, mdb.CREATE, mdb.TYPE, FIELD, mdb.NAME, kit.Select(m.ShortKey(), cmd), mdb.TEXT, kit.Format(kit.Simple(arg...)), SPACE, m.Option(ice.MSG_USERPOD)).Result()))))
}
func ProxyUpload(m *ice.Message, pod string, p string) string {
	pp := path.Join(ice.VAR_PROXY, pod, p)
	size, cache := int64(0), time.Now().Add(-time.Hour*24)
	if s, e := file.StatFile(pp); e == nil {
		size, cache = s.Size(), s.ModTime()
	} else if s, e := file.StatFile(p); e == nil {
		size, cache = s.Size(), s.ModTime()
	}
	if m.Cmdv(SPACE, pod, mdb.TYPE) == ORIGIN {
		m.Cmd(SPIDE, pod, SPIDE_SAVE, pp, "/p/"+p)
	} else {
		kit.If(p == ice.BIN_ICE_BIN, func() { m.Option(ice.MSG_USERROLE, aaa.TECH) })
		share := m.Cmdx(SHARE, mdb.CREATE, mdb.TYPE, PROXY, mdb.NAME, p, mdb.TEXT, pod)
		defer m.Cmd(SHARE, mdb.REMOVE, mdb.HASH, share)
		url := tcp.PublishLocalhost(m, m.MergeLink(PP(SHARE, PROXY), SHARE, share))
		m.Cmd(SPACE, pod, SPIDE, PROXY, URL, url, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, mdb.AT+p, kit.Dict(ice.MSG_USERROLE, aaa.TECH))
	}
	return kit.Select(p, pp, file.ExistsFile(pp))
}
