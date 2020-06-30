package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"net/http"
	"path"
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
			msg := m.Spawns(s)
			m.Confm(SERVE, "meta.static", func(key string, value string) {
				m.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*Frame); ok && n.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					w.HandleFunc(k, func(w http.ResponseWriter, r *http.Request) {
						m.TryCatch(msg.Spawns(), true, func(msg *ice.Message) {
							_serve_handle(k, x, msg, w, r)
						})
					})
				}
			})
		}
	})

	// TODO simple
	m.Richs(SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
		client := value["client"].(map[string]interface{})

		// 服务地址
		port := m.Cap(ice.CTX_STREAM, client["hostname"])
		m.Log("serve", "listen %s %s %v", arg[0], port, m.Conf(cli.RUNTIME, "node"))

		// 启动服务
		web.m, web.Server = m, &http.Server{Addr: port, Handler: web}
		m.Event(gdb.SERVE_START, arg[0])
		m.Warn(true, "listen %s", web.Server.ListenAndServe())
		m.Event(gdb.SERVE_CLOSE, arg[0])
	})
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网络模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			m.Cmd(SPIDE, mdb.CREATE, "dev", kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_dev")))
			m.Cmd(SPIDE, mdb.CREATE, "self", kit.Select("http://:9020", m.Conf(cli.RUNTIME, "conf.ctx_self")))
			m.Cmd(SPIDE, mdb.CREATE, "shy", kit.Select("https://shylinux.com:443", m.Conf(cli.RUNTIME, "conf.ctx_shy")))

			m.Cmd(aaa.ROLE, aaa.White, aaa.VOID, "web", "/publish/")
			m.Cmd(aaa.ROLE, aaa.White, aaa.VOID, ctx.COMMAND)

			m.Cmd(mdb.SEARCH, mdb.CREATE, FAVOR)
			m.Cmd(mdb.SEARCH, mdb.CREATE, SPIDE)
			m.Cmd(mdb.RENDER, mdb.CREATE, SPIDE)

			for k := range c.Commands[mdb.RENDER].Action {
				m.Cmdy(mdb.RENDER, mdb.CREATE, k, mdb.RENDER, c.Cap(ice.CTX_FOLLOW))
			}
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(SPIDE, SERVE, GROUP, LABEL,
				FAVOR, CACHE, STORY, SHARE)

			m.Done()
			m.Richs(SPACE, nil, "*", func(key string, value map[string]interface{}) {
				if kit.Format(value["type"]) == "master" {
					m.Done()
				}
			})
		}},
	},
}

func init() {
	ice.Index.Register(Index, &Frame{},
		SPIDE, SERVE, SPACE, DREAM,
		CACHE, FAVOR, STORY, SHARE,
	)
}
