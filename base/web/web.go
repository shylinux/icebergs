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
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	*ice.Message
	*http.Server
	*http.ServeMux
}

func (f *Frame) Begin(m *ice.Message, arg ...string) {}
func (f *Frame) Start(m *ice.Message, arg ...string) {
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
			msg.Log("route", "%s <= %s", p.Name, route)
			pf.Handle(route, http.StripPrefix(path.Dir(route), f))
			list[c] = path.Join(list[p], route)
		}
		for key, cmd := range c.Commands {
			if key[:1] != ice.PS {
				continue
			}
			func(key string, cmd *ice.Command) {
				msg.Log("route", "%s <- %s", c.Name, key)
				f.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
					m.TryCatch(m.Spawn(key, cmd, c, w, r), true, func(msg *ice.Message) { _serve_handle(key, cmd, msg, w, r) })
				})
				ice.Info.Route[path.Join(list[c], key)] = ctx.FileURI(cmd.FileLine())
			}(key, cmd)
		}
	})
	switch cb := m.OptionCB("").(type) {
	case func(http.Handler):
		cb(f)
	default:
		m.Cmd(tcp.SERVER, tcp.LISTEN, mdb.TYPE, HTTP, m.OptionSimple(mdb.NAME, tcp.HOST, tcp.PORT), func(l net.Listener) {
			defer mdb.HashCreateDeferRemove(m, m.OptionSimple(mdb.NAME, tcp.PROTO), arg, cli.STATUS, tcp.START)()
			gdb.EventDeferEvent(m, SERVE_START, arg)
			m.Warn(f.Server.Serve(l))
		})
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {}
func (f *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _serve_main(f.Message, w, r) {
		f.ServeMux.ServeHTTP(w, r)
	}
}

const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{}, BROAD, SERVE, SPACE, DREAM, SHARE, CACHE, SPIDE)
}
func ApiAction(arg ...string) ice.Actions { return ice.Actions{kit.Select(ice.PS, arg, 0): {}} }
func Prefix(arg ...string) string         { return kit.Keys(WEB, arg) }

func P(arg ...string) string  { return path.Join(ice.PS, path.Join(arg...)) }
func PP(arg ...string) string { return P(arg...) + ice.PS }
