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
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _share_link(m *ice.Message, p string, arg ...interface{}) string {
	p = kit.Select("", "/share/local/", !strings.HasPrefix(p, "/")) + p
	return tcp.ReplaceLocalhost(m, kit.MergeURL2(kit.Select(m.Option(ice.MSG_USERWEB), m.Conf(SHARE, kit.Keym(kit.MDB_DOMAIN))), p, arg...))
}
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(cli.POD); m.PodCmd(CACHE, arg[0]) {
		if m.Append(kit.MDB_FILE) == "" {
			m.RenderResult(m.Append(kit.MDB_TEXT))
		} else {
			m.Option(cli.POD, pod)
			_share_local(m, m.Append(kit.MDB_FILE))
		}
		return
	}
	msg := m.Cmd(CACHE, arg[0])
	m.RenderDownload(msg.Append(kit.MDB_FILE), msg.Append(kit.MDB_TYPE), msg.Append(kit.MDB_NAME))
}
func _share_local(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, "/"); ls[0] {
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

	if m.Option(cli.POD) != "" { // 远程文件
		pp := path.Join(ice.VAR_PROXY, m.Option(cli.POD), p)
		cache := time.Now().Add(-time.Hour * 240000)
		if s, e := os.Stat(pp); e == nil {
			cache = s.ModTime()
		}

		// 上传文件
		m.Cmdy(SPACE, m.Option(cli.POD), SPIDE, ice.DEV, SPIDE_RAW, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/proxy/"),
			SPIDE_PART, m.OptionSimple(cli.POD), kit.MDB_PATH, p, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)

		if s, e := os.Stat(pp); e == nil && !s.IsDir() {
			p = pp
		}
	}
	if p == path.Join(ice.USR_PUBLISH, ice.ORDER_JS) {
		if _, e := os.Stat(p); os.IsNotExist(e) {
			m.RenderResult("")
			return
		}
	}
	m.RenderDownload(p)
}
func _share_proxy(m *ice.Message, arg ...string) {
	switch p := path.Join(ice.VAR_PROXY, m.Option(cli.POD), m.Option(kit.MDB_PATH)); m.R.Method {
	case http.MethodGet: // 下发文件
		m.RenderDownload(path.Join(p, m.Option(kit.MDB_NAME)))

	case http.MethodPost: // 上传文件
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option(kit.MDB_DATA), p)
		m.RenderResult(m.Option(kit.MDB_PATH))
	}
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	prefix := kit.Path(m.Conf(SERVE, kit.Keym(ice.REQUIRE)))
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil { // 克隆代码
		m.Cmd("web.code.git.repos", mdb.CREATE, kit.SSH_REPOS, "https://"+repos, kit.MDB_PATH, path.Join(prefix, repos))
	}
	m.RenderDownload(path.Join(prefix, repos, path.Join(arg...)))
}

const (
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
	FIELD = "field"
)
const SHARE = "share"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SHARE: {Name: SHARE, Help: "共享链", Value: kit.Data(
				kit.MDB_EXPIRE, "72h", kit.MDB_FIELD, "time,hash,userrole,username,river,storm,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				ice.AddRender(ice.RENDER_DOWNLOAD, func(m *ice.Message, cmd string, args ...interface{}) string {
					list := []string{}
					if m.Option(ice.MSG_USERPOD) != "" {
						list = append(list, cli.POD, m.Option(ice.MSG_USERPOD))
					}

					arg := kit.Simple(args...)
					if len(arg) > 1 {
						list = append(list, "filename", arg[0])
					}
					return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`,
						_share_link(m, kit.Select(arg[0], arg, 1), list), path.Base(arg[0]), arg[0])
				})
			}},
			SHARE: {Name: "share hash auto prunes", Help: "共享链", Action: ice.MergeAction(map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, kit.MDB_TIME, m.Time(m.Config(kit.MDB_EXPIRE)),
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						RIVER, m.Option(ice.MSG_RIVER), STORM, m.Option(ice.MSG_STORM), arg)
					// m.Option(kit.MDB_LINK, _share_link(m, "/share/"+m.Result()))
					m.Option(kit.MDB_LINK, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/"+m.Result()))
				}},
				LOGIN: {Name: "login userrole=void,tech username", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(SHARE, mdb.CREATE, kit.MDB_TYPE, LOGIN, m.OptionSimple(aaa.USERROLE, aaa.USERNAME))
					m.EchoQRCode(msg.Option(kit.MDB_LINK))
					m.ProcessInner()
				}},
			}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.PodCmd(SHARE, arg) {
					return
				}
				m.Fields(len(arg), m.Conf(SHARE, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, kit.MDB_HASH, arg); len(arg) > 0 {
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
				if msg := m.Cmd(SHARE, m.Option(SHARE)); kit.Int(msg.Append(kit.MDB_TIME)) < kit.Int(msg.FormatTime()) {
					m.RenderResult("共享超时")
					return
				}
				m.RenderIndex(SERVE, ice.VOLCANOS)
			}},

			"/share/cache/": {Name: "/share/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_cache(m, arg...)
			}},
			"/share/proxy/": {Name: "/share/proxy/", Help: "文件流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_proxy(m, arg...)
			}},
			"/share/repos/": {Name: "/share/repos/", Help: "代码库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
			}},
			"/share/local/": {Name: "/share/local/", Help: "文件夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_local(m, arg...)
			}},
			"/share/local/avatar": {Name: "avatar", Help: "头像", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.AVATAR), "/share/local/"))
			}},
			"/share/local/background": {Name: "background", Help: "壁纸", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.RenderDownload(strings.TrimPrefix(m.Cmd(aaa.USER, m.Option(ice.MSG_USERNAME)).Append(aaa.BACKGROUND), "/share/local/"))
			}},
		},
	})
}
