package relay

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/ssh"
	kit "shylinux.com/x/toolkits"
)

const (
	SSH_AUTHS     = "ssh.auths"
	SSH_AUTH_KEYS = ".ssh/authorized_keys"
)

type auths struct {
	relay
	insert string `name:"insert machine* server*"`
	list   string `name:"list auto"`
}

func (s auths) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case MACHINE:
		m.Cmdy(s.relay).Cut(arg[0], mdb.ICONS)
		m.DisplayInputKeyNameIcon()
	case web.SERVER:
		m.AdminCmd(web.DREAM, web.SERVER).Table(func(value ice.Maps) {
			if value[nfs.MODULE] == ice.Info.Make.Module {
				m.Push(arg[0], value[mdb.NAME])
				m.Push(mdb.ICONS, value[mdb.ICONS])
				m.Push(nfs.MODULE, value[nfs.MODULE])
			}
		})
		m.DisplayInputKeyNameIcon()
	}
}
func (s auths) Insert(m *ice.Message, arg ...string) {
	m.Options(m.Cmd(s.relay, m.Option(tcp.MACHINE)).AppendSimple())
	msg := m.AdminCmd(web.SPACE, m.Option(web.SERVER), ssh.RSA, ssh.PUBLIC)
	if msg.IsErr() {
		m.Copy(msg)
		return
	}
	key := msg.Result()
	list := strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, kit.Split(s.CmdArgs(m, cli.CD, ctx.CMDS, cli.CAT+lex.SP+SSH_AUTH_KEYS), lex.SP, lex.SP)).Result()), lex.NL)
	if kit.IndexOf(list, key) == -1 {
		m.Cmd(cli.SYSTEM, kit.Split(s.CmdArgs(m, cli.CD), lex.SP, lex.SP), ctx.CMDS, kit.Format(`echo -e %q >> `+SSH_AUTH_KEYS, key))
	}
}
func (s auths) Delete(m *ice.Message, arg ...string) {
	m.Cmd(s.relay).GoToastTable(tcp.MACHINE, func(val ice.Maps) {
		if m.Option(val[MACHINE]) != ice.TRUE {
			return
		}
		list := []string{}
		kit.For(strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, kit.Split(s.CmdArgs(m.Spawn(val), cli.CD, ctx.CMDS, cli.CAT+lex.SP+SSH_AUTH_KEYS), lex.SP, lex.SP)).Result()), lex.NL), func(text string) {
			if ls := kit.Split(text); len(ls) > 2 {
				kit.If(ls[2] != kit.Format("%s@%s", m.Option(aaa.USERNAME), m.Option(web.SERVER)), func() { list = append(list, text) })
			}
		})
		m.Push(MACHINE, val[MACHINE]).Push(mdb.TEXT, strings.Join(list, lex.NL))
		m.Cmd(cli.SYSTEM, kit.Split(s.CmdArgs(m.Spawn(val), cli.CD), lex.SP, lex.SP), ctx.CMDS, kit.Format(`echo -e %q > `+SSH_AUTH_KEYS, strings.Join(list, lex.NL)))
	})
}
func (s auths) List(m *ice.Message, arg ...string) {
	list := map[string]map[string]bool{}
	head := []string{}
	m.Cmd(s.relay).GoToastTable(tcp.MACHINE, func(val ice.Maps) {
		head = append(head, val[tcp.MACHINE])
		kit.For(strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, kit.Split(s.CmdArgs(m.Spawn(val), cli.CD, ctx.CMDS, cli.CAT+lex.SP+SSH_AUTH_KEYS), lex.SP, lex.SP)).Result()), lex.NL), func(text string) {
			if ls := kit.Split(text); len(ls) > 2 {
				if _, ok := list[ls[2]]; !ok {
					list[ls[2]] = map[string]bool{}
				}
				list[ls[2]][val[tcp.MACHINE]] = true
			}
		})
	})
	for _, server := range kit.SortedKey(list) {
		ls := kit.Split(server, "@")
		m.Push(aaa.USERNAME, ls[0]).Push(web.SERVER, ls[1])
		keys := list[server]
		for _, k := range head {
			if keys[k] {
				m.Push(k, keys[k])
			} else {
				m.Push(k, "")
			}
		}
	}
	m.PushAction(s.Delete).Action(s.Insert)
}

func init() { ice.Cmd(SSH_AUTHS, auths{}) }
