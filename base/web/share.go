package web

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _share_link(m *ice.Message, p string, arg ...interface{}) string {
	p = kit.Select("", "/share/local/", !strings.HasPrefix(p, ice.PS)) + p
	return tcp.ReplaceLocalhost(m, m.MergeURL2(p, arg...))
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	if repos == ice.Info.Make.Module && kit.FileExists(path.Join(arg...)) {
		m.RenderDownload(path.Join(arg...))
		return
	}
	prefix := kit.Path(m.Conf(SERVE, kit.Keym(ice.REQUIRE)))
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil { // 克隆代码
		m.Cmd("web.code.git.repos", mdb.CREATE, nfs.REPOS, "https://"+repos, nfs.PATH, path.Join(prefix, repos))
	}
	m.RenderDownload(path.Join(prefix, repos, path.Join(arg...)))
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
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(ice.POD); m.PodCmd(CACHE, arg[0]) {
		if m.Append(nfs.FILE) == "" {
			m.RenderResult(m.Append(mdb.TEXT))
		} else {
			m.Option(ice.POD, pod)
			_share_local(m, m.Append(nfs.FILE))
		}
		return
	}
	msg := m.Cmd(CACHE, arg[0])
	m.RenderDownload(msg.Append(nfs.FILE), msg.Append(mdb.TYPE), msg.Append(mdb.NAME))
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
		if !m.Right(ls) {
			m.Render(STATUS, http.StatusUnauthorized, ice.ErrNotRight)
			return // 没有权限
		}
	}

	if m.Option(ice.POD) != "" { // 远程文件
		pp := path.Join(ice.VAR_PROXY, m.Option(ice.POD), p)
		cache := time.Now().Add(-time.Hour * 240000)
		if s, e := os.Stat(pp); e == nil {
			cache = s.ModTime()
		}

		// 上传文件
		m.Cmdy(SPACE, m.Option(ice.POD), SPIDE, ice.DEV, SPIDE_RAW, m.MergeURL2(SHARE_PROXY, nfs.PATH, ""),
			SPIDE_PART, m.OptionSimple(ice.POD), nfs.PATH, p, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)

		if s, e := os.Stat(pp); e == nil && !s.IsDir() {
			p = pp
		}
	}
	if strings.HasSuffix(p, path.Join(ice.USR_PUBLISH, ice.ORDER_JS)) {
		if _, e := os.Stat(p); os.IsNotExist(e) {
			m.RenderResult("")
			return
		}
	}
	m.RenderDownload(p)
}

const (
	TOPIC = "topic"
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
	FIELD = "field"

	SHARE_TOAST = "/share/toast/"
	SHARE_CACHE = "/share/cache/"
	SHARE_REPOS = "/share/repos/"
	SHARE_PROXY = "/share/proxy/"
	SHARE_LOCAL = "/share/local/"

	SHARE_LOCAL_AVATAR     = "/share/local/avatar/"
	SHARE_LOCAL_BACKGROUND = "/share/local/background/"
)
const SHARE = "share"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SHARE: {Name: SHARE, Help: "共享链", Value: kit.Data(
			mdb.EXPIRE, "72h", mdb.FIELD, "time,hash,userrole,username,river,storm,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		SHARE: {Name: "share hash auto prunes", Help: "共享链", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_DOWNLOAD, func(msg *ice.Message, cmd string, args ...interface{}) string {
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
			}},
			mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, mdb.TIME, m.Time(m.Config(mdb.EXPIRE)),
					aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.USERNICK, m.Option(ice.MSG_USERNICK),
					RIVER, m.Option(ice.MSG_RIVER), STORM, m.Option(ice.MSG_STORM), arg)
				m.Option(mdb.LINK, _share_link(m, "/share/"+m.Result()))
			}},
			LOGIN: {Name: "login userrole=void,tech username", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(SHARE, mdb.CREATE, mdb.TYPE, LOGIN, m.OptionSimple(aaa.USERROLE, aaa.USERNAME))
				m.EchoQRCode(msg.Option(mdb.LINK))
				m.ProcessInner()
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.PodCmd(SHARE, arg) && m.Length() > 0 {
				return
			}
			if mdb.HashSelect(m, arg...); len(arg) > 0 {
				link := _share_link(m, "/share/"+arg[0])
				m.PushQRCode(cli.QRCODE, link)
				m.PushScript(ssh.SCRIPT, link)
				m.PushAnchor(link)
			} else {
				m.Action(LOGIN)
			}
			m.PushAction(mdb.REMOVE)
			m.StatusTimeCount()
		}},
		"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(SHARE, kit.Select(m.Option(SHARE), arg, 0))
			if msg := m.Cmd(SHARE, m.Option(SHARE)); kit.Int(msg.Append(mdb.TIME)) < kit.Int(msg.FormatTime()) {
				m.RenderResult("共享超时")
				return
			} else {
				switch msg.Append(mdb.TYPE) {
				case LOGIN:
					if sessid := aaa.SessCreate(m, msg.Append(aaa.USERNAME)); m.Option(ice.MSG_USERWEB) == "" {
						m.RenderRedirect(ice.PS, ice.MSG_SESSID, sessid)
					} else {
						RenderCookie(m, sessid)
						RenderRedirect(m, ice.PS)
					}
				default:
					m.RenderIndex(SERVE, ice.VOLCANOS)
				}
			}
		}},

		SHARE_TOAST: {Name: "/share/toast/", Help: "推送流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(SPACE, m.Option(ice.POD), m.Optionv("cmds"))
		}},
		SHARE_CACHE: {Name: "/share/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_share_cache(m, arg...)
		}},
		SHARE_REPOS: {Name: "/share/repos/", Help: "代码库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
		}},
		SHARE_PROXY: {Name: "/share/proxy/", Help: "文件流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_share_proxy(m)
		}},
		SHARE_LOCAL: {Name: "/share/local/", Help: "文件夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_share_local(m, arg...)
		}},
		SHARE_LOCAL_AVATAR: {Name: "avatar", Help: "头像", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// RenderType(m.W, "", "image/svg+xml")
			// m.RenderResult(`<svg font-size="32" text-anchor="middle" dominant-baseline="middle" width="80" height="60" xmlns="http://www.w3.org/2000/svg"><text x="40" y="30" stroke="red">hello</text></svg>`)
			m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.AVATAR), SHARE_LOCAL))
		}},
		SHARE_LOCAL_BACKGROUND: {Name: "background", Help: "壁纸", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.BACKGROUND), SHARE_LOCAL))
		}},
	}})
}
