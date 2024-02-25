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
	status string `name:"status" help:"源码" icon:"bi bi-git"`
	list   string `name:"list refresh" help:"矩阵"`
}

func (s matrix) List(m *ice.Message, arg ...string) {
	m.Cmdy(SSH_RELAY, web.DREAM).Table(func(value ice.Maps) {
		if value[mdb.STATUS] == cli.STOP {
			m.PushButton()
		} else if value[web.SPACE] == ice.CONTEXTS {
			m.PushButton(s.Portal, s.Desktop, s.Dream, s.Admin, s.Open, s.Word, s.Status, s.Vimer, s.Compile, s.Runtime, s.Xterm)
		} else if value[MACHINE] == tcp.LOCALHOST {
			m.PushButton(s.Portal, s.Word, s.Status, s.Vimer, s.Compile, s.Runtime, s.Xterm, s.Desktop, s.Admin, s.Open)
		} else {
			m.PushButton(s.Portal, s.Vimer, s.Runtime, s.Xterm, s.Desktop, s.Admin, s.Open)
		}
	}).Action(html.FILTER).Display("").Sort("type,status,space,machine", []string{web.SERVER, web.WORKER, ""}, []string{cli.START, cli.STOP, ""}, "str_r", "str")
}
func (s matrix) Portal(m *ice.Message, arg ...string)  { s.iframe(m, arg...) }
func (s matrix) Desktop(m *ice.Message, arg ...string) { s.plug(m, arg...) }
func (s matrix) Dream(m *ice.Message, arg ...string)   { s.plug(m, arg...) }
func (s matrix) Admin(m *ice.Message, arg ...string)   { s.open(m, arg...) }
func (s matrix) Open(m *ice.Message, arg ...string)    { s.open(m, arg...) }
func (s matrix) Word(m *ice.Message, arg ...string)    { s.plug(m, arg...) }
func (s matrix) Status(m *ice.Message, arg ...string)  { s.plug(m, arg...) }
func (s matrix) Vimer(m *ice.Message, arg ...string)   { s.plug(m, arg...) }
func (s matrix) Compile(m *ice.Message, arg ...string) { s.plug(m, arg...) }
func (s matrix) Runtime(m *ice.Message, arg ...string) { s.plug(m, arg...) }
func (s matrix) Xterm(m *ice.Message, arg ...string)   { s.xterm(m, arg...) }

func init() { ice.Cmd("ssh.matrix", matrix{}) }

func (s matrix) plug(m *ice.Message, arg ...string) {
	if !kit.HasPrefixList(arg, ctx.RUN) {
		defer m.Push(web.TITLE, s.title(m))
	}
	m.ProcessPodCmd(kit.Keys(kit.Select("", ice.OPS, ice.Info.NodeType == web.WORKER), m.Option(MACHINE), m.Option(web.SPACE)), m.ActionKey(), arg, arg...)
}
func (s matrix) xterm(m *ice.Message, arg ...string) {
	m.ProcessXterm(func() []string {
		cmd, dir := cli.SH, ice.CONTEXTS
		if m.Option(MACHINE) == "" {
			cmd = cli.Shell(m.Message)
		} else {
			cmd, dir = m.Option(MACHINE), kit.Select(dir, m.Cmd(SSH_RELAY, m.Option(MACHINE)).Append(web.DREAM))
		}
		kit.If(m.Option(web.SPACE), func() { dir = path.Join(dir, nfs.USR_LOCAL_WORK+m.Option(web.SPACE)) })
		return []string{cmd, "", kit.Format("cd ~/%s", dir)}
	}, arg...)
	kit.If(!kit.HasPrefixList(arg, ctx.RUN), func() { m.Push(web.STYLE, html.FLOAT).Push(web.TITLE, s.title(m)) })
}
func (s matrix) iframe(m *ice.Message, arg ...string) {
	m.ProcessIframe(s.title(m), s.link(m), arg...)
}
func (s matrix) open(m *ice.Message, arg ...string) {
	if kit.HasPrefixList(arg, ctx.RUN) || m.Option(MACHINE) == "" {
		m.ProcessIframe(s.title(m), s.link(m), arg...)
	} else {
		m.ProcessOpen(s.link(m))
	}
}
func (s matrix) link(m *ice.Message, arg ...string) (res string) {
	kit.If(m.Option(MACHINE), func(p string) { res = m.Cmd(SSH_RELAY, p).Append(mdb.LINK) })
	kit.If(m.Option(web.SPACE), func(p string) { res += web.S(p) })
	kit.If(m.ActionKey() != web.OPEN, func() { res += web.C(m.ActionKey()) })
	return kit.Select(nfs.PS, res)
}
func (s matrix) title(m *ice.Message) string {
	return kit.Select(ice.CONTEXTS, kit.Keys(m.Option(MACHINE), m.Option(web.SPACE), kit.Select("", m.ActionKey(), m.ActionKey() != web.OPEN)))
}
