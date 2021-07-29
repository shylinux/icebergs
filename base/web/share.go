package web

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

func _share_domain(m *ice.Message) string {
	link := m.Conf(SHARE, kit.Keym(kit.MDB_DOMAIN))
	if link == "" {
		link = m.Cmd(SPACE, SPIDE_DEV, cli.PWD).Append(kit.MDB_LINK)
	}
	if link == "" {
		link = m.Cmd(SPACE, SPIDE_SHY, cli.PWD).Append(kit.MDB_LINK)
	}
	if link == "" {
		link = kit.Format("http://%s:%s", m.Cmd(tcp.HOST).Append(tcp.IP), m.Cmd(SERVE).Append(tcp.PORT))
	}
	return link
}
func _share_cache(m *ice.Message, arg ...string) {
	if pod := m.Option(cli.POD); pod != "" {
		m.Option(cli.POD, "")
		msg := m.Cmd(SPACE, pod, CACHE, arg[0])
		if msg.Append(kit.MDB_FILE) == "" {
			m.Render(ice.RENDER_RESULT, msg.Append(kit.MDB_TEXT))
		} else {
			m.Option(cli.POD, pod)
			_share_local(m, msg.Append(kit.MDB_FILE))
		}
		return
	}
	msg := m.Cmd(CACHE, arg[0])
	m.Render(ice.RENDER_DOWNLOAD, msg.Append(kit.MDB_FILE), msg.Append(kit.MDB_TYPE), msg.Append(kit.MDB_NAME))
}
func _share_local(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, "/"); ls[0] {
	case kit.SSH_ETC, kit.SSH_VAR: // 私有文件
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

		m.Cmdy(SPACE, m.Option(cli.POD), SPIDE, SPIDE_DEV, SPIDE_RAW, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/proxy/"),
			SPIDE_PART, cli.POD, m.Option(cli.POD), kit.MDB_PATH, p, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)

		if s, e := os.Stat(pp); e == nil && !s.IsDir() {
			p = pp
		}
	}

	if p == path.Join(ice.USR_PUBLISH, ice.ORDER_JS) {
		if _, e := os.Stat(p); os.IsNotExist(e) {
			m.Render(ice.RENDER_RESULT, "")
			return
		}
	}
	m.Render(ice.RENDER_DOWNLOAD, p)
}
func _share_proxy(m *ice.Message, arg ...string) {
	switch m.R.Method {
	case http.MethodGet: // 下发文件
		m.Render(ice.RENDER_DOWNLOAD, path.Join(ice.VAR_PROXY, path.Join(m.Option(cli.POD), m.Option(kit.MDB_PATH), m.Option(kit.MDB_NAME))))

	case http.MethodPost: // 上传文件
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option(kit.MDB_DATA), path.Join(ice.VAR_PROXY, m.Option(cli.POD), m.Option(kit.MDB_PATH)))
		m.Render(ice.RENDER_RESULT, m.Option(kit.MDB_PATH))
	}
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	prefix := kit.Path(m.Conf(SERVE, kit.Keym(ice.REQUIRE)))
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil {
		m.Cmd("web.code.git.repos", mdb.CREATE, kit.SSH_REPOS, "https://"+repos, kit.MDB_PATH, path.Join(prefix, repos))
	}
	m.Render(ice.RENDER_DOWNLOAD, path.Join(prefix, repos, path.Join(arg...)))
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
			SHARE: {Name: SHARE, Help: "共享链", Value: kit.Data(kit.MDB_EXPIRE, "72h")},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				ice.AddRender(ice.RENDER_DOWNLOAD, func(m *ice.Message, cmd string, args ...interface{}) string {
					arg := kit.Simple(args...)
					if arg[0] == "" {
						return ""
					}
					list := []string{}
					if m.Option(ice.MSG_USERPOD) != "" {
						list = append(list, "pod", m.Option(ice.MSG_USERPOD))
					}
					if len(arg) == 1 {
						arg[0] = kit.MergeURL2(m.Option(ice.MSG_USERWEB), path.Join(kit.Select("", "/share/local",
							!strings.HasPrefix(arg[0], "/")), arg[0]), list)
					} else {
						arg[1] = kit.MergeURL2(m.Option(ice.MSG_USERWEB), path.Join(kit.Select("", "/share/local",
							!strings.HasPrefix(arg[1], "/")), arg[1]), list, "filename", arg[0])
					}
					arg[0] = m.ReplaceLocalhost(arg[0])
					return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, m.ReplaceLocalhost(kit.Select(arg[0], arg, 1)), path.Base(arg[0]), arg[0])
				})
			}},
			SHARE: {Name: "share hash auto prunes", Help: "共享链", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, SHARE, "", mdb.HASH, kit.MDB_TIME, m.Time(m.Conf(SHARE, kit.Keym(kit.MDB_EXPIRE))),
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						RIVER, m.Option(ice.MSG_RIVER), STORM, m.Option(ice.MSG_STORM), arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.SELECT: {Name: "select hash", Help: "查询", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,userrole,username,river,storm,type,name,text")
					m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {}},
				mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					list := []string{}
					m.Richs(SHARE, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if value = kit.GetMeta(value); kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Option("before")) {
							list = append(list, key)
						}
					})
					m.Option(mdb.FIELDS, "time,userrole,username,river,storm,type,name,text")
					for _, v := range list {
						m.Cmdy(mdb.DELETE, SHARE, "", mdb.HASH, kit.MDB_HASH, v)
					}
				}},

				LOGIN: {Name: "login userrole=void,tech username", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					m.EchoQRCode(kit.MergeURL(_share_domain(m),
						SHARE, m.Cmdx(SHARE, mdb.CREATE, kit.MDB_TYPE, LOGIN,
							aaa.USERNAME, kit.Select(m.Option(ice.MSG_USERNAME), m.Option(aaa.USERNAME)),
							aaa.USERROLE, m.Option(aaa.USERROLE),
						)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,hash,type,name,text,userrole,username,river,storm")
				m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)

				if len(arg) > 0 {
					link := kit.MergeURL(m.Option(ice.MSG_USERWEB), SHARE, arg[0])
					if strings.Contains(link, tcp.LOCALHOST) {
						link = strings.Replace(link, tcp.LOCALHOST, m.Cmd(tcp.HOST, ice.OptionFields(tcp.IP)).Append(tcp.IP), 1)
					}

					m.PushAnchor(link)
					m.PushScript("shell", link)
					m.PushQRCode("scan", link)
				} else {
					m.Action(LOGIN)
				}
			}},
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(0, "time,hash,userrole,username,river,storm,type,name,text")
				msg := m.Cmd(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, kit.Select(m.Option(SHARE), arg, 0))

				list := []string{SHARE, kit.Select(m.Option(SHARE), arg, 0)}
				for _, k := range []string{RIVER, STORM} {
					if msg.Append(k) != "" {
						list = append(list, k, msg.Append(k))
					}
				}

				switch msg.Append(kit.MDB_TYPE) {
				case LOGIN, RIVER:
					m.RenderRedirect("/", list)

				case STORM:
					m.RenderRedirect("/page/share.html", SHARE, m.Option(SHARE))
				}
			}},

			"/share/cache/": {Name: "/share/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_cache(m, arg...)
			}},
			"/share/local/": {Name: "/share/local/", Help: "文件夹", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_local(m, arg...)
			}},
			"/share/proxy/": {Name: "/share/proxy/", Help: "文件流", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_proxy(m, arg...)
			}},
			"/share/repos/": {Name: "/share/repos/", Help: "代码库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_repos(m, path.Join(arg[0], arg[1], arg[2]), arg[3:]...)
			}},
		}})
}
