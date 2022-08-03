package web

import (
	"net"
	"net/http"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/task"
)

type Frame struct {
	*ice.Message
	*http.Client
	*http.Server
	*http.ServeMux

	send ice.Messages
	lock task.Lock
}

func (frame *Frame) getSend(key string) (*ice.Message, bool) {
	defer frame.lock.RLock()()
	msg, ok := frame.send[key]
	return msg, ok
}
func (frame *Frame) addSend(key string, msg *ice.Message) string {
	defer frame.lock.Lock()()
	frame.send[key] = msg
	return key
}
func (frame *Frame) delSend(key string) {
	defer frame.lock.Lock()()
	delete(frame.send, key)
}

func (frame *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _serve_main(frame.Message, w, r) {
		frame.ServeMux.ServeHTTP(w, r)
	}
}
func (frame *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (frame *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	frame.send = ice.Messages{}
	return frame
}
func (frame *Frame) Start(m *ice.Message, arg ...string) bool {
	list := map[*ice.Context]string{}
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if frame, ok := s.Server().(*Frame); ok {
			if frame.ServeMux != nil {
				return
			}
			frame.ServeMux = http.NewServeMux()
			meta := logs.FileLineMeta("")

			// 级联路由
			msg := m.Spawn(s)
			if pframe, ok := p.Server().(*Frame); ok && pframe.ServeMux != nil {
				route := ice.PS + s.Name + ice.PS
				msg.Log(ROUTE, "%s <= %s", p.Name, route, meta)
				pframe.Handle(route, http.StripPrefix(path.Dir(route), frame))
				list[s] = path.Join(list[p], route)
			}

			// 静态路由
			m.Confm(SERVE, kit.Keym(nfs.PATH), func(key string, value string) {
				m.Log(ROUTE, "%s <- %s <- %s", s.Name, key, value, meta)
				frame.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s != sub || k[0] != '/' {
					return
				}
				msg.Log(ROUTE, "%s <- %s", s.Name, k, meta)
				ice.Info.Route[path.Join(list[s], k)] = ctx.FileCmd(kit.FileLine(x.Hand, 300))
				frame.HandleFunc(k, func(frame http.ResponseWriter, r *http.Request) {
					m.TryCatch(msg.Spawn(), true, func(msg *ice.Message) {
						_serve_handle(k, x, msg, frame, r)
					})
				})
			})
		}
	})

	gdb.Event(m, SERVE_START)
	defer gdb.Event(m, SERVE_STOP)

	frame.Message, frame.Server = m, &http.Server{Handler: frame}
	switch cb := m.OptionCB("").(type) {
	case func(http.Handler):
		cb(frame) // 启动框架
	default:
		m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, WEB, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(l net.Listener) {
			mdb.HashCreate(m, mdb.NAME, WEB, arg, m.OptionSimple(tcp.PROTO, ice.DEV), cli.STATUS, tcp.START, kit.Dict(mdb.TARGET, l))
			defer mdb.HashModify(m, m.OptionSimple(mdb.NAME), cli.STATUS, tcp.STOP)
			m.Warn(frame.Server.Serve(l)) // 启动服务
		})
	}
	return true
}
func (frame *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	SERVE_START = "serve.start"
	SERVE_STOP  = "serve.stop"
	WEBSITE     = "website"
)
const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{},
		BROAD, SERVE, SPACE, DREAM,
		SHARE, CACHE, SPIDE, ROUTE,
	)
}

func P(arg ...string) string  { return path.Join(ice.PS, path.Join(arg...)) }
func PP(arg ...string) string { return P(arg...) + ice.PS }
