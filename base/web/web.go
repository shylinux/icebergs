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
	meta := logs.FileLineMeta("")
	list := map[*ice.Context]string{}
	m.Travel(func(p *ice.Context, c *ice.Context) {
		if f, ok := c.Server().(*Frame); ok {
			if f.ServeMux != nil {
				return
			}
			f.ServeMux = http.NewServeMux()
			msg := m.Spawn(c)
			if pf, ok := p.Server().(*Frame); ok && pf.ServeMux != nil {
				route := ice.PS + c.Name + ice.PS
				msg.Log(ROUTE, "%s <= %s", p.Name, route, meta)
				pf.Handle(route, http.StripPrefix(path.Dir(route), f))
				list[c] = path.Join(list[p], route)
			}
			m.Confm(SERVE, kit.Keym(nfs.PATH), func(key string, value string) {
				m.Log(ROUTE, "%s <- %s <- %s", c.Name, key, value, meta)
				f.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})
			m.Travel(func(p *ice.Context, _c *ice.Context, key string, cmd *ice.Command) {
				if c != _c || key[0] != '/' {
					return
				}
				msg.Log(ROUTE, "%s <- %s", c.Name, key, meta)
				ice.Info.Route[path.Join(list[c], key)] = ctx.FileURI(cmd.GetFileLine())
				f.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
					m.TryCatch(msg.Spawn(), true, func(msg *ice.Message) { _serve_handle(key, cmd, msg, w, r) })
				})
			})
		}
	})

	gdb.Event(m, SERVE_START, arg)
	defer gdb.Event(m, SERVE_STOP)

	frame.Message, frame.Server = m, &http.Server{Handler: frame}
	switch cb := m.OptionCB("").(type) {
	case func(http.Handler):
		cb(frame) // 启动框架
	default:
		mdb.HashCreate(m, mdb.NAME, WEB, arg, m.OptionSimple(tcp.PROTO, ice.DEV), cli.STATUS, tcp.START)
		m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, WEB, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(l net.Listener) {
			defer mdb.HashModify(m, m.OptionSimple(mdb.NAME), cli.STATUS, tcp.STOP)
			mdb.HashTarget(m, m.Option(mdb.NAME), func() ice.Any { return l })
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

	CODE_INNER = "web.code.inner"
	WIKI_WORD  = "web.wiki.word"
)
const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{},
		BROAD, SERVE, SPACE, DREAM,
		SHARE, CACHE, SPIDE, ROUTE,
	)
}

func ApiAction(arg ...string) ice.Actions {
	return ice.Actions{kit.Select(ice.PS, arg, 0): {}}
}
func P(arg ...string) string  { return path.Join(ice.PS, path.Join(arg...)) }
func PP(arg ...string) string { return P(arg...) + ice.PS }
