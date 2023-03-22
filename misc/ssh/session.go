package ssh

import (
	"io"

	"golang.org/x/crypto/ssh"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _ssh_session(m *ice.Message, client *ssh.Client) (*ssh.Session, error) {
	s, e := client.NewSession()
	m.Assert(e)
	out, e := s.StdoutPipe()
	m.Assert(e)
	in, e := s.StdinPipe()
	m.Assert(e)
	h := m.Cmdx(SESSION, mdb.CREATE, mdb.STATUS, tcp.OPEN, CONNECT, m.Option(mdb.NAME), kit.Dict(mdb.TARGET, in))
	m.Go(func() {
		buf := make([]byte, ice.MOD_BUFS)
		for {
			if n, e := out.Read(buf); e != nil {
				break
			} else {
				m.Cmd(SESSION, mdb.INSERT, mdb.HASH, h, mdb.TYPE, RES, mdb.TEXT, string(buf[:n]))
			}
		}
	})
	return s, nil
}

const (
	TTY = "tty"
	ENV = "env"
	CMD = "cmd"
	ARG = "arg"
	RES = "res"
)
const (
	INPUT  = "input"
	OUTPUT = "output"
)

const SESSION = "session"

func init() {
	psh.Index.MergeCommands(ice.Commands{
		SESSION: {Name: "session hash id auto", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.REPEAT: {Help: "执行", Hand: func(m *ice.Message, arg ...string) { m.Cmdy("", ctx.ACTION, ctx.COMMAND, CMD, m.Option(mdb.TEXT)) }},
			ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(mdb.HASH), mdb.TYPE, CMD, mdb.TEXT, m.Option(CMD))
				if w, ok := mdb.HashSelectTarget(m, m.Option(mdb.HASH), nil).(io.Writer); ok {
					w.Write([]byte(m.Option(CMD) + ice.NL))
					m.Sleep300ms()
				}
			}},
		}, mdb.PageZoneAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,count,status,connect", mdb.FIELDS, "time,id,type,text")), Hand: func(m *ice.Message, arg ...string) {
			m.Fields(len(kit.Slice(arg, 0, 2)), mdb.Config(m, mdb.FIELD), mdb.Config(m, mdb.FIELDS))
			if mdb.PageZoneSelect(m, arg...); len(arg) == 0 {
				m.Tables(func(value ice.Maps) {
					m.PushButton(kit.Select("", ctx.COMMAND, value[mdb.STATUS] == tcp.OPEN), mdb.REMOVE)
				})
			} else {
				m.Tables(func(value ice.Maps) { m.PushButton(kit.Select("", mdb.REPEAT, value[mdb.TYPE] == CMD)) }).Action(ctx.COMMAND, mdb.PAGE)
			}
		}},
	})
}
