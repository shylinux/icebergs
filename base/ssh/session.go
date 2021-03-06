package ssh

import (
	"io"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
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

			m.Grow(SESSION, kit.Keys(kit.MDB_HASH, h), kit.Dict(
				kit.MDB_TYPE, RES, kit.MDB_TEXT, string(buf[:n]),
			))
		}
	})

	m.Richs(SESSION, "", h, func(key string, value map[string]interface{}) {
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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESSION: {Name: SESSION, Help: "会话", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SESSION: {Name: "session hash id auto command prunes", Help: "会话", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SESSION, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,count,connect")
					m.Cmdy(mdb.PRUNES, SESSION, "", mdb.HASH, kit.MDB_STATUS, tcp.ERROR)
					m.Cmdy(mdb.PRUNES, SESSION, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
				ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(SESSION, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						if w, ok := kit.Value(value, kit.Keym(INPUT)).(io.Writer); ok {
							m.Grow(SESSION, kit.Keys(kit.MDB_HASH, key), kit.Dict(kit.MDB_TYPE, CMD, kit.MDB_TEXT, m.Option(CMD)))
							w.Write([]byte(m.Option(CMD) + ice.MOD_NL))
						}
					})
					m.ProcessRefresh("300ms")
				}},
				mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SESSION, kit.MDB_ACTION, ctx.COMMAND, CMD, m.Option(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Fields(len(arg) == 0, "time,hash,status,count,connect")
					if m.Cmdy(mdb.SELECT, SESSION, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
						m.Table(func(index int, value map[string]string, head []string) {
							m.PushButton(kit.Select("", ctx.COMMAND, value[kit.MDB_STATUS] == tcp.OPEN), mdb.REMOVE)
						})
					}
					return
				}

				m.Fields(len(arg) == 1, "time,id,type,text")
				m.Cmdy(mdb.SELECT, SESSION, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushButton(kit.Select("", mdb.REPEAT, value[kit.MDB_TYPE] == CMD))
				})
			}},
		},
	})
}
