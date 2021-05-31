package web

import (
	"net"
	"net/http"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

type Frame struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m *ice.Message

	send map[string]*ice.Message
}

func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _serve_main(web.m, w, r) {
		web.ServeMux.ServeHTTP(w, r)
	}
}
func (web *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (web *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	web.send = map[string]*ice.Message{}
	return web
}
func (web *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if w, ok := s.Server().(*Frame); ok {
			if w.ServeMux != nil {
				return
			}
			w.ServeMux = http.NewServeMux()

			// 静态路由
			msg := m.Spawn(s)
			m.Confm(SERVE, kit.Keym("static"), func(key string, value string) {
				m.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 级联路由
			route := "/" + s.Name + "/"
			if f, ok := p.Server().(*Frame); ok && f.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				f.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleFunc(k, func(w http.ResponseWriter, r *http.Request) {
						m.TryCatch(msg.Spawn(), true, func(msg *ice.Message) {
							_serve_handle(k, x, msg, w, r)
						})
					})
				}
			})
		}
	})

	web.m, web.Server = m, &http.Server{Handler: web}
	m.Option(kit.Keycb(tcp.LISTEN), func(l net.Listener) {
		m.Cmdy(mdb.INSERT, SERVE, "", mdb.HASH, arg, kit.MDB_STATUS, tcp.START, kit.MDB_PROTO, m.Option(kit.MDB_PROTO), SPIDE_DEV, m.Option(SPIDE_DEV))
		defer m.Cmd(mdb.MODIFY, SERVE, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_STATUS, tcp.STOP)

		// 启动服务
		m.Warn(true, SERVE, ": ", web.Server.Serve(l))
	})

	m.Cmd(tcp.SERVER, tcp.LISTEN, kit.MDB_TYPE, WEB, kit.MDB_NAME, m.Option(kit.MDB_NAME),
		tcp.HOST, m.Option(tcp.HOST), tcp.PORT, m.Option(tcp.PORT))
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Conf(SPACE, kit.MDB_HASH, "")
			m.Cmd(mdb.SEARCH, mdb.CREATE, SPACE, m.Prefix(SPACE))

			m.Cmd(SPIDE, mdb.CREATE, SPIDE_DEV, kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_dev")))
			m.Cmd(SPIDE, mdb.CREATE, SPIDE_SELF, kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_self")))
			m.Cmd(SPIDE, mdb.CREATE, SPIDE_SHY, kit.Select("https://shylinux.com:443", m.Conf(cli.RUNTIME, "conf.ctx_shy")))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()

			m.Done(true)
			m.Cmd(SERVE).Table(func(index int, value map[string]string, head []string) {
				m.Done(value[kit.MDB_STATUS] == tcp.START)
			})
		}},
	},
}

func init() {
	ice.Index.Register(Index, &Frame{},
		SPIDE, CACHE, STORY, ROUTE,
		SHARE, SERVE, SPACE, DREAM,
	)
}
