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

func _ssh_session(m *ice.Message, h string, client *ssh.Client) (*ssh.Session, error) {
	session, e := client.NewSession()
	m.Assert(e)

	out, e := session.StdoutPipe()
	m.Assert(e)

	in, e := session.StdinPipe()
	m.Assert(e)

	m.Go(func() {
		buf := make([]byte, ice.MOD_BUFS)
		for {
			n, e := out.Read(buf)
			if e != nil {
				break
			}

			mdb.Grow(m, SESSION, kit.Keys(mdb.HASH, h), kit.Dict(
				mdb.TYPE, RES, mdb.TEXT, string(buf[:n]),
			))
		}
	})

	mdb.Richs(m, SESSION, "", h, func(key string, value ice.Map) {
		kit.Value(value, kit.Keym(OUTPUT), out, kit.Keym(INPUT), in)
	})

	return session, nil
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
		SESSION: {Name: "session name id auto", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SESSION, ctx.ACTION, ctx.COMMAND, CMD, m.Option(mdb.TEXT))
			}},
			ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				mdb.Richs(m, SESSION, "", m.Option(mdb.NAME), func(key string, value ice.Map) {
					if w, ok := kit.Value(value, kit.Keym(INPUT)).(io.Writer); ok {
						mdb.Grow(m, SESSION, kit.Keys(mdb.HASH, key), kit.Dict(mdb.TYPE, CMD, mdb.TEXT, m.Option(CMD)))
						w.Write([]byte(m.Option(CMD) + ice.NL))
					}
				})
				m.ProcessRefresh300ms()
			}},
		}, mdb.ZoneAction(mdb.SHORT, "name", mdb.FIELD, "time,name,status,count,connect")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Tables(func(value ice.Maps) {
					m.PushButton(kit.Select("", ctx.COMMAND, value[mdb.STATUS] == tcp.OPEN), mdb.REMOVE)
				})
				return
			}

			m.Action(ctx.COMMAND, mdb.PAGE)
			mdb.OptionPage(m, kit.Slice(arg, 2)...)
			m.Fields(len(kit.Slice(arg, 1, 2)), "time,id,type,text")
			mdb.ZoneSelect(m, kit.Slice(arg, 0, 2)...).Tables(func(value ice.Maps) {
				m.PushButton(kit.Select("", mdb.REPEAT, value[mdb.TYPE] == CMD))
			})
		}},
	})
}
