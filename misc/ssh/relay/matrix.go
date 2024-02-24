package relay

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

type matrix struct {
	list string `name:"list refresh" help:"矩阵"`
}

func (s matrix) List(m *ice.Message, arg ...string) *ice.Message {
	m.Cmdy(SSH_RELAY, web.DREAM).PushAction(s.Portal, s.Admin, s.Desktop, s.Xterm, s.Runtime).Action(html.FILTER).Display("")
	m.Sort("type,status,space,machine", []string{web.SERVER, web.WORKER, ""}, []string{cli.START, cli.STOP, ""}, "str_r", "str")
	return m
}
func (s matrix) Portal(m *ice.Message, arg ...string)  { s.iframe(m, arg...) }
func (s matrix) Admin(m *ice.Message, arg ...string)   { s.open(m, arg...) }
func (s matrix) Desktop(m *ice.Message, arg ...string) { s.open(m, arg...) }
func (s matrix) Runtime(m *ice.Message, arg ...string) { s.plug(m, arg...) }
func (s matrix) Xterm(m *ice.Message, arg ...string) {
	m.ProcessXterm(func() []string {
		if m.Option(MACHINE) == tcp.LOCALHOST {
			p := kit.Select(ice.CONTEXTS)
			kit.If(m.Option(web.SPACE) != ice.CONTEXTS, func() { p = path.Join(p, nfs.USR_LOCAL_WORK+m.Option(web.SPACE)) })
			return []string{cli.Shell(m.Message), "", kit.Format("cd ~/%s", p)}
		}
		msg := m.Cmd(SSH_RELAY, m.Option(MACHINE))
		p := kit.Select(ice.CONTEXTS, msg.Append(web.DREAM))
		kit.If(m.Option(web.SPACE) != ice.CONTEXTS, func() { p = path.Join(p, nfs.USR_LOCAL_WORK+m.Option(web.SPACE)) })
		return []string{m.Option(MACHINE), "", kit.Format("cd ~/%s", p)}
	}, arg...)
	kit.If(!kit.HasPrefixList(arg, ice.RUN), func() {
		m.Push(ctx.STYLE, html.FLOAT)
		m.Push(web.TITLE, s.title(m))
	})
}

func init() { ice.Cmd("ssh.matrix", matrix{}) }

func (s matrix) title(m *ice.Message) string {
	return kit.Keys(kit.Select("", m.Option(MACHINE), m.Option(MACHINE) != tcp.LOCALHOST), m.Option(web.SPACE), m.ActionKey())
}
func (s matrix) iframe(m *ice.Message, arg ...string) {
	m.ProcessIframe(s.title(m), s.link(m), arg...)
}
func (s matrix) plug(m *ice.Message, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) || m.Option(MACHINE) == tcp.LOCALHOST {
		m.ProcessFloat(m.ActionKey(), arg, arg...)
		if !kit.HasPrefixList(arg, ctx.RUN) {
			m.Push(web.TITLE, s.title(m))
		}
	} else {
		m.ProcessOpen(s.link(m))
	}
}
func (s matrix) open(m *ice.Message, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) || m.Option(MACHINE) == tcp.LOCALHOST {
		m.ProcessIframe(s.title(m), s.link(m), arg...)
	} else {
		m.ProcessOpen(s.link(m))
	}
}
func (s matrix) link(m *ice.Message, arg ...string) (res string) {
	if m.Option(MACHINE) == tcp.LOCALHOST {
		res = m.UserHost()
	} else {
		res = m.Cmd(SSH_RELAY, m.Option(MACHINE)).Append(mdb.LINK)
	}
	if m.Option(web.SPACE) != ice.CONTEXTS {
		res += web.S(m.Option(web.SPACE))
	}
	if m.ActionKey() != web.OPEN {
		res += web.C(m.ActionKey())
	}
	return
}
