package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func _share_cache(m *ice.Message, arg ...string) {
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

	if m.Option(kit.SSH_POD) != "" { // 远程文件
		pp := path.Join("var/proxy", m.Option(kit.SSH_POD), p)
		cache := time.Now().Add(-time.Hour * 240000)
		if s, e := os.Stat(pp); e == nil {
			cache = s.ModTime()
		}

		m.Cmdy(SPACE, m.Option(kit.SSH_POD), SPIDE, SPIDE_DEV, SPIDE_RAW, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/proxy/"),
			SPIDE_PART, kit.SSH_POD, m.Option(kit.SSH_POD), kit.MDB_PATH, p, CACHE, cache.Format(ice.MOD_TIME), UPLOAD, "@"+p)

		if s, e := os.Stat(pp); e == nil && !s.IsDir() {
			p = pp
		}
	}

	if p == "usr/publish/order.js" {
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
		m.Render(ice.RENDER_DOWNLOAD, path.Join("var/proxy", path.Join(m.Option(kit.SSH_POD), m.Option(kit.MDB_PATH), m.Option(kit.MDB_NAME))))

	case http.MethodPost: // 上传文件
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option("data"), path.Join("var/proxy", m.Option(kit.SSH_POD), m.Option(kit.MDB_PATH)))
		m.Render(ice.RENDER_RESULT, m.Option(kit.MDB_PATH))
	}
}
func _share_repos(m *ice.Message, repos string, arg ...string) {
	prefix := kit.Path(m.Conf(SERVE, "meta.require"))
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil {
		m.Cmd("web.code.git.repos", mdb.CREATE, kit.SSH_REPOS, "https://"+repos, kit.MDB_PATH, path.Join(prefix, repos))
	}
	m.Render(ice.RENDER_DOWNLOAD, path.Join(prefix, repos, path.Join(arg...)))
}

const (
	LOGIN = "login"
	RIVER = "river"
	STORM = "storm"
)
const SHARE = "share"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SHARE: {Name: SHARE, Help: "共享链", Value: kit.Data(kit.MDB_EXPIRE, "72h")},
		},
		Commands: map[string]*ice.Command{
			SHARE: {Name: "share hash auto", Help: "共享链", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, SHARE, "", mdb.HASH, kit.MDB_TIME, m.Time(m.Conf(SHARE, "meta.expire")),
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						RIVER, m.Option(ice.MSG_RIVER), STORM, m.Option(ice.MSG_STORM), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.SELECT: {Name: "select hash", Help: "查询", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,userrole,username,river,storm,type,name,text")
					m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,userrole,username,river,storm,type,name,text", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)

				if len(arg) > 0 {
					m.PushAnchor(kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/"+arg[0], SHARE, arg[0]))
					m.PushScript("shell", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/"+arg[0], SHARE, arg[0]))
					m.PushQRCode("share", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/"+arg[0], SHARE, arg[0]))
				}
			}},
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,userrole,username,river,storm,type,name,text"))
				msg := m.Cmd(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, kit.Select(m.Option(SHARE), arg, 0))

				list := []string{SHARE, kit.Select(m.Option(SHARE), arg, 0)}
				for _, k := range []string{RIVER, STORM} {
					if msg.Append(k) != "" {
						list = append(list, k, msg.Append(k))
					}
				}

				switch msg.Append(kit.MDB_TYPE) {
				case LOGIN, RIVER:
					switch kit.Select("", arg, 1) {
					case SHARE:
						m.Render(ice.RENDER_QRCODE, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/", list))
					default:
						m.Render(REDIRECT, "/", list)
					}

				case STORM:
					switch kit.Select("", arg, 1) {
					case SHARE:
						m.Render(ice.RENDER_QRCODE, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/page/share.html", list))
					default:
						m.Render(REDIRECT, "/page/share.html", SHARE, m.Option(SHARE))
					}
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
