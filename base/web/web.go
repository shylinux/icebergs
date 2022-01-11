package web

import (
	"net"
	"net/http"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	*ice.Message
	*http.Client
	*http.Server
	*http.ServeMux

	send map[string]*ice.Message
}

func (web *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _serve_main(web.Message, w, r) {
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
		if frame, ok := s.Server().(*Frame); ok {
			if frame.ServeMux != nil {
				return
			}
			frame.ServeMux = http.NewServeMux()

			// 级联路由
			msg := m.Spawn(s)
			if pframe, ok := p.Server().(*Frame); ok && pframe.ServeMux != nil {
				route := ice.PS + s.Name + ice.PS
				msg.Log(ROUTE, "%s <= %s", p.Name, route)
				pframe.Handle(route, http.StripPrefix(path.Dir(route), frame))
			}

			// 静态路由
			m.Confm(SERVE, kit.Keym(nfs.PATH), func(key string, value string) {
				m.Log(ROUTE, "%s <- %s <- %s", s.Name, key, value)
				frame.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s != sub || k[0] != '/' {
					return
				}
				msg.Log(ROUTE, "%s <- %s", s.Name, k)
				frame.HandleFunc(k, func(frame http.ResponseWriter, r *http.Request) {
					m.TryCatch(msg.Spawn(), true, func(msg *ice.Message) {
						_serve_handle(k, x, msg, frame, r)
					})
				})
			})
		}
	})

	m.Event(SERVE_START)
	defer m.Event(SERVE_STOP)

	web.Message, web.Server = m, &http.Server{Handler: web}
	switch cb := m.OptionCB(SERVE).(type) {
	case func(http.Handler):
		cb(web) // 启动框架
	default:
		m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, WEB, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(l net.Listener) {
			m.Cmdy(mdb.INSERT, SERVE, "", mdb.HASH, arg, m.OptionSimple(tcp.PROTO, ice.DEV), cli.STATUS, tcp.START)
			defer m.Cmd(mdb.MODIFY, SERVE, "", mdb.HASH, m.OptionSimple(mdb.NAME), cli.STATUS, tcp.STOP)
			m.Warn(web.Server.Serve(l)) // 启动服务
		})
	}
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return m.Done(true)
}

const (
	SERVE_START = "serve.start"
	SERVE_STOP  = "serve.stop"
)
const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{},
		SERVE, SPACE, DREAM, ROUTE,
		SHARE, SPIDE, CACHE, STORY,
	)
}
