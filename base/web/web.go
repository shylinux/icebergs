package web

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
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
			route := nfs.PS + c.Name + nfs.PS
			msg.Log("route", "%s <= %s", p.Name, route)
			pf.Handle(route, http.StripPrefix(path.Dir(route), f))
			list[c] = path.Join(list[p], route)
		}
		for key, cmd := range c.Commands {
			if key[:1] != nfs.PS {
				continue
			}
			func(key string, cmd *ice.Command) {
				msg.Log("route", "%s <- %s", c.Name, key)
				f.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
					m.TryCatch(m.Spawn(key, cmd, c, w, r), true, func(msg *ice.Message) { _serve_handle(key, cmd, msg, w, r) })
				})
				ice.Info.Route[path.Join(list[c], key)+kit.Select("", nfs.PS, strings.HasSuffix(key, nfs.PS))] = ctx.FileURI(cmd.FileLine())
			}(key, cmd)
		}
	})
	switch cb := m.OptionCB("").(type) {
	case func(http.Handler):
		cb(f)
	default:
		m.Cmdy(tcp.SERVER, tcp.LISTEN, mdb.TYPE, HTTP, mdb.NAME, logs.FileLine(1), m.OptionSimple(tcp.HOST, tcp.PORT), func(l net.Listener) {
			defer mdb.HashCreateDeferRemove(m, m.OptionSimple(mdb.NAME, tcp.PROTO), arg, cli.STATUS, tcp.START)()
			gdb.Event(m, SERVE_START, arg)
			m.Warn(f.Server.Serve(l))
		})
		if m.IsErr() {
			fmt.Println()
			fmt.Println(m.Result())
			m.Cmd(ice.QUIT)
		}
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {}
func (f *Frame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kit.If(_serve_main(f.Message, w, r), func() { f.ServeMux.ServeHTTP(w, r) })
}

const WEB = "web"

var Index = &ice.Context{Name: WEB, Help: "网络模块"}

func init() {
	ice.Index.Register(Index, &Frame{}, COUNT, BROAD, SERVE, SPACE, DREAM, CACHE, SPIDE, SHARE)
}

func ApiAction(arg ...string) ice.Actions { return ice.Actions{kit.Select(nfs.PS, arg, 0): {}} }
func Prefix(arg ...string) string {
	for i, k := range arg {
		switch k {
		case "Register":
			arg[i] = "Index.Register"
		}
	}
	return kit.Keys(WEB, arg)
}

func P(arg ...string) string  { return path.Join(nfs.PS, path.Join(arg...)) }
func PP(arg ...string) string { return P(arg...) + nfs.PS }
