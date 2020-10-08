package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func _share_repos(m *ice.Message, repos string, arg ...string) {
	prefix := m.Conf(SERVE, "meta.volcanos.require")
	if _, e := os.Stat(path.Join(prefix, repos)); e != nil {
		m.Cmd(cli.SYSTEM, "git", "clone", "https://"+repos, path.Join(prefix, repos))
	}
	m.Render(ice.RENDER_DOWNLOAD, path.Join(prefix, repos, path.Join(arg...)))
}
func _share_remote(m *ice.Message, pod string, arg ...string) {
	m.Cmdy(SPACE, pod, "web./publish/", arg)
	m.Render(ice.RENDER_RESULT)
}
func _share_local(m *ice.Message, arg ...string) {
	p := path.Join(arg...)
	switch ls := strings.Split(p, "/"); ls[0] {
	case "etc", "var":
		// 私有文件
		m.Render(STATUS, http.StatusUnauthorized, "not auth")
		return
	default:
		if m.Warn(!m.Right(ls), ice.ErrNotAuth, m.Option(ice.MSG_USERROLE), " of ", p) {
			m.Render(STATUS, http.StatusUnauthorized, "not auth")
			return
		}
	}

	if m.Option("pod") != "" {
		// 远程文件
		pp := path.Join("var/proxy", m.Option("pod"), p)
		cache := time.Now().Add(-time.Hour * 240000)
		if s, e := os.Stat(pp); e == nil {
			cache = s.ModTime()
		}
		m.Cmdy(SPACE, m.Option("pod"), SPIDE, "dev", "raw", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/proxy/"),
			"part", "pod", m.Option("pod"), "path", p, "cache", cache.Format(ice.MOD_TIME), "upload", "@"+p)

		m.Render(ice.RENDER_DOWNLOAD, path.Join("var/proxy", m.Option("pod"), p))
		return
	}

	// 本地文件
	m.Render(ice.RENDER_DOWNLOAD, p)
}
func _share_proxy(m *ice.Message, arg ...string) {
	switch m.Option(ice.MSG_METHOD) {
	case http.MethodGet:
		m.Render(ice.RENDER_DOWNLOAD, path.Join("var/proxy", path.Join(m.Option("pod"), m.Option("path"), m.Option("name"))))
	case http.MethodPost:
		m.Cmdy(CACHE, UPLOAD)
		m.Cmdy(CACHE, WATCH, m.Option("data"), path.Join("var/proxy", m.Option("pod"), m.Option("path")))
		m.Render(ice.RENDER_RESULT, m.Option("path"))
	}

}
func _share_cache(m *ice.Message, arg ...string) {
	msg := m.Cmd(CACHE, arg[0])
	m.Render(ice.RENDER_DOWNLOAD, msg.Append(kit.MDB_FILE), msg.Append(kit.MDB_TYPE), msg.Append(kit.MDB_NAME))
}

const SHARE = "share"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SHARE: {Name: SHARE, Help: "共享链", Value: kit.Data("expire", "72h")},
		},
		Commands: map[string]*ice.Command{
			SHARE: {Name: "share hash auto", Help: "共享链", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, SHARE, "", mdb.HASH,
						aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
						"river", m.Option(ice.MSG_RIVER), "storm", m.Option(ice.MSG_STORM),
						kit.MDB_TIME, m.Time(m.Conf(SHARE, "meta.expire")), arg)
				}},
				mdb.SELECT: {Name: "select hash", Help: "查询", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,userrole,username,river,storm,type,name,text")
					m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SHARE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,userrole,username,river,storm,type,name,text", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("删除")
			}},
			"/share/": {Name: "/share/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,userrole,username,river,storm,type,name,text"))
				switch msg := m.Cmd(mdb.SELECT, SHARE, "", mdb.HASH, kit.MDB_HASH, arg[0]); msg.Append(kit.MDB_TYPE) {
				case "login":
					switch kit.Select("", arg, 1) {
					case "share":
						list := []string{}
						for _, k := range []string{"river", "storm"} {
							if msg.Append(k) != "" {
								list = append(list, k, msg.Append(k))
							}
						}
						m.Render(ice.RENDER_QRCODE, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/", SHARE, arg[0], list))
					}
				}
			}},

			"/share/local/": {Name: "/share/local/", Help: "共享链", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_local(m, arg...)
			}},
			"/share/proxy/": {Name: "/share/proxy/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_proxy(m, arg...)
			}},
			"/share/cache/": {Name: "/share/cache/", Help: "缓存池", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_cache(m, arg...)
			}},
			"/plugin/github.com/": {Name: "/space/", Help: "空间站", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_share_repos(m, path.Join(strings.Split(cmd, "/")[2:5]...), arg[6:]...)
			}},
			"/publish/": {Name: "/publish/", Help: "发布", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "order.js" && len(ice.BinPack) > 0 {
					m.Render(ice.RENDER_RESULT, "{}")
					return
				}
				if p := m.Option("pod"); p != "" {
					m.Option("pod", "")
					_share_remote(m, p, arg...)
					return
				}

				p := path.Join(kit.Simple(m.Conf(SERVE, "meta.publish"), arg)...)
				if m.W == nil {
					m.Cmdy("nfs.cat", p)
					return
				}
				_share_local(m, p)
			}},
		}}, nil)
}
