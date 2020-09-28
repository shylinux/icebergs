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

func _ssh_sess(m *ice.Message, h string, client *ssh.Client) (*ssh.Session, error) {
	session, e := client.NewSession()
	m.Assert(e)

	out, e := session.StdoutPipe()
	m.Assert(e)

	in, e := session.StdinPipe()
	m.Assert(e)

	m.Go(func() {
		for {
			buf := make([]byte, 1024)
			n, e := out.Read(buf)
			if e != nil {
				break
			}

			m.Debug(string(buf[:n]))
			m.Grow(SESSION, kit.Keys(kit.MDB_HASH, h), kit.Dict(
				kit.MDB_TYPE, RES, kit.MDB_TEXT, string(buf[:n]),
			))
		}
	})

	m.Richs(SESSION, "", h, func(key string, value map[string]interface{}) {
		kit.Value(value, "meta.output", out)
		kit.Value(value, "meta.input", in)
	})

	return session, nil
}

const (
	CMD = "cmd"
	ARG = "arg"
	ENV = "env"
	RES = "res"
)

const SESSION = "session"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESSION: {Name: SESSION, Help: "会话", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			SESSION: {Name: "session hash id auto 命令 清理", Help: "会话", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(SESSION, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						if w, ok := kit.Value(value, "meta.input").(io.Writer); ok {
							m.Grow(SESSION, kit.Keys(kit.MDB_HASH, key), kit.Dict(kit.MDB_TYPE, RES, kit.MDB_TEXT, m.Option(CMD)))
							n, e := w.Write([]byte(m.Option(CMD) + "\n"))
							m.Debug("%v %v", n, e)
						}
					})
					m.Sleep("300ms")
					m.Cmdy(SESSION, m.Option(kit.MDB_HASH))
				}},

				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, SESSION, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,status,count,connect")
					m.Cmdy(mdb.SELECT, SESSION, "", mdb.HASH, kit.MDB_HASH, arg)
					return
				}

				m.Option(mdb.FIELDS, "time,id,type,text")
				m.Cmdy(mdb.SELECT, SESSION, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.Sort(kit.MDB_ID)
			}},
		},
	}, nil)
}
