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
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
	"shylinux.com/x/toolkits/logs"
)

func _share_render(m *ice.Message, arg ...string) {
	ice.AddRender(ice.RENDER_DOWNLOAD, func(msg *ice.Message, cmd string, args ...ice.Any) string {
		list := []string{}
		if msg.Option(ice.MSG_USERPOD) != "" {
			list = append(list, ice.POD, msg.Option(ice.MSG_USERPOD))
		}

		arg := kit.Simple(args...)
		if len(arg) > 1 {
			list = append(list, "filename", arg[0])
		}
		return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`,
			_share_link(msg, kit.Select(arg[0], arg, 1), list), path.Base(arg[0]), arg[0])
	})
}
func _share_link(m *ice.Message, p string, arg ...ice.Any) string {
	p = kit.Select("", SHARE_LOCAL, !strings.HasPrefix(p, ice.PS)) + p
	return tcp.ReplaceLocalhost(m, MergeLink(m, p, arg...))
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
	case ice.ETC, ice.VAR: // 私有文件
		if m.Option(ice.MSG_USERROLE) == aaa.VOID {
			m.Render(STATUS, http.StatusUnauthorized, ice.ErrNotRight)
			return // 没有权限
		}
	default:
		if !aaa.Right(m, ls) && m.Option(ice.POD) == "" {
			m.Render(STATUS, http.StatusUnauthorized, ice.ErrNotRight)
			return // 没有权限
		}
	}

	if m.Option(ice.POD) == "" {
		m.RenderDownload(p)
		return // 本地文件
	}

	pp := path.Join(ice.VAR_PROXY, m.Option(ice.POD), p)
	cache, size := time.Now().Add(-time.Hour*240000), int64(0)
	if s, e := file.StatFile(pp); e == nil {
		cache, size = s.ModTime(), s.Size()
	}

	// 上传文件
	if p == ice.BIN_ICE_BIN {
		aaa.UserRoot(m).Cmd(SPACE, m.Option(ice.POD), SPIDE, "submit", MergeURL2(m, SHARE_PROXY, nfs.PATH, ""), m.Option(ice.POD), p, size, cache.Format(ice.MOD_TIME))
	} else {
		m.Cmd(SPACE, m.Option(ice.POD), SPIDE, ice.DEV, SPIDE_RAW, MergeURL2(m, SHARE_PROXY, nfs.PATH, ""),
			SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, p, nfs.SIZE, size, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)
	}
	if s, e := file.StatFile(pp); e == nil && !s.IsDir() {
		p = pp
	}

	if m.Warn(!file.ExistsFile(p)) {
		m.RenderStatusNotFound()
		return
	}
	m.RenderDownload(p)
}
func _share_proxy(m *ice.Message) {
	switch p := path.Join(ice.VAR_PROXY, m.Option(ice.POD), m.Option(nfs.PATH)); m.R.Method {
	case http.MethodGet: // 下发文件
		m.RenderDownload(path.Join(p, m.Option(mdb.NAME)))

	case http.MethodPost: // 上传文件
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option(mdb.DATA), p)
		m.RenderResult(m.Option(nfs.PATH))
	}
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	if repos == ice.Info.Make.Module && nfs.ExistsFile(m, path.Join(arg...)) {
		m.RenderDownload(path.Join(arg...))
		return
	}
	if !nfs.ExistsFile(m, path.Join(ice.ISH_PLUGED, repos)) { // 克隆代码
		m.Cmd("web.code.git.repos", mdb.CREATE, nfs.REPOS, "https://"+repos, nfs.PATH, path.Join(ice.ISH_PLUGED, repos))
	}
	m.RenderDownload(path.Join(ice.ISH_PLUGED, repos, path.Join(arg...)))
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
			mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, arg, aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK))
				m.Option(mdb.LINK, _share_link(m, PP(SHARE)+m.Result()))
			}},
			LOGIN: {Name: "login userrole=void,tech username", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN, m.OptionSimple(aaa.USERROLE, aaa.USERNAME))
				m.EchoQRCode(msg.Option(mdb.LINK))
				m.ProcessInner()
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,userrole,username,usernick,river,storm,type,name,text", mdb.EXPIRE, "72h")), Hand: func(m *ice.Message, arg ...string) {
			if ctx.PodCmd(m, SHARE, arg) && m.Length() > 0 {
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				link := _share_link(m, P(SHARE, arg[0]))
				m.PushQRCode(cli.QRCODE, link)
				m.PushScript(ssh.SCRIPT, link)
				m.PushAnchor(link)
			} else {
				m.Action(LOGIN)
			}
		}},
		PP(SHARE): {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(SHARE, m.Option(SHARE, kit.Select(m.Option(SHARE), arg, 0)))
			if m.Warn(msg.Append(mdb.TIME) < msg.Time(), ice.ErrNotValid, kit.Select(m.Option(SHARE), arg, 0), msg.Append(mdb.TIME)) {
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
			m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.AVATAR), SHARE_LOCAL))
		}},
		SHARE_LOCAL_BACKGROUND: {Name: "background", Help: "壁纸", Hand: func(m *ice.Message, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.BACKGROUND), SHARE_LOCAL))
		}},
		SHARE_PROXY: {Name: "/share/proxy/", Help: "文件流", Hand: func(m *ice.Message, arg ...string) {
			_share_proxy(m)
		}},
		SHARE_REPOS: {Name: "/share/repos/", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
		}},
	})
}

func IsNotValidShare(m *ice.Message, value ice.Maps) bool {
	_source := logs.FileLineMeta(logs.FileLine(2))
	if m.Warn(value[mdb.TIME] < m.Time(), ice.ErrNotValid, m.Option(SHARE), value[mdb.TIME], m.Time(), _source) {
		return true
	}
	return false
}
