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
	"shylinux.com/x/toolkits/task"
)

type Frame struct {
	*ice.Message
	*http.Server
	*http.ServeMux
	lock task.Lock
	send ice.Messages
}

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.send = ice.Messages{}
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	f.Message, f.Server = m, &http.Server{Handler: f}
	list := map[*ice.Context]string{}
	m.Travel(func(p *ice.Context, c *ice.Context) {
		f, ok := c.Server().(*Frame)
		if !ok || f.ServeMux != nil {
			return
		}
		f.ServeMux = http.NewServeMux()
		msg := m.Spawn(c)
		if pf, ok := p.Server().(*Frame); ok && pf.ServeMux != nil {
			route := ice.PS + c.Name + ice.PS
			msg.Log(ROUTE, "%s <= %s", p.Name, route)
			pf.Handle(route, http.StripPrefix(path.Dir(route), f))
			list[c] = path.Join(list[p], route)
		}
		m.Confm(SERVE, kit.Keym(nfs.PATH), func(key string, value string) {
			msg.Log(ROUTE, "%s <- %s <- %s", c.Name, key, value)
			f.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
		})
		for key, cmd := range c.Commands {
			if key[0] != '/' {
				continue
			}
			func(key string, cmd *ice.Command) {
				msg.Log(ROUTE, "%s <- %s", c.Name, key)
				f.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
					msg.TryCatch(msg.Spawn(), true, func(msg *ice.Message) { _serve_handle(key, cmd, msg, w, r) })
				})
				ice.Info.Route[path.Join(list[c], key)] = ctx.FileURI(cmd.GetFileLine())
			}(key, cmd)
		}
	})
	gdb.Event(m, SERVE_START, arg)
	defer gdb.Event(m, SERVE_STOP)
	switch cb := m.OptionCB("").(type) {
	case func(http.Handler):
		cb(f)
	default:
		m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, WEB, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(l net.Listener) {
			mdb.HashCreate(m, mdb.NAME, WEB, arg, m.OptionSimple(tcp.PROTO, ice.DEV), cli.STATUS, tcp.START)
			defer mdb.HashModify(m, m.OptionSimple(mdb.NAME), cli.STATUS, tcp.STOP)
			m.Warn(f.Server.Serve(l))
		})
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _serve_main(f.Message, w, r) {
		f.ServeMux.ServeHTTP(w, r)
	}
}
func (f *Frame) getSend(key string) (*ice.Message, bool) {
	defer f.lock.RLock()()
	msg, ok := f.send[key]
	return msg, ok
}
func (f *Frame) addSend(key string, msg *ice.Message) string {
	defer f.lock.Lock()()
	f.send[key] = msg
	return key
}
func (f *Frame) delSend(key string) string {
	defer f.lock.Lock()()
	delete(f.send, key)
	return key
}

const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{}, BROAD, SERVE, SPACE, DREAM, SHARE, CACHE, SPIDE, ROUTE)
}
func ApiAction(arg ...string) ice.Actions { return ice.Actions{kit.Select(ice.PS, arg, 0): {}} }

func P(arg ...string) string  { return path.Join(ice.PS, path.Join(arg...)) }
func PP(arg ...string) string { return P(arg...) + ice.PS }
