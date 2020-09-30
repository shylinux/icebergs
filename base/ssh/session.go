package ssh

import (
	"os"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"io"

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
			buf := make([]byte, 4096)
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

func _watch(m *ice.Message, from io.Reader, to io.Writer, cb func([]byte)) {
	m.Go(func() {
		buf := make([]byte, 1024)
		for {
			n, e := from.Read(buf)
			if e != nil {
				cb(nil)
				break
			}
			cb(buf[:n])
			to.Write(buf[:n])
		}
	})
}

const (
	TTY = "tty"
	ENV = "env"
	ARG = "arg"
	CMD = "cmd"
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
							m.Grow(SESSION, kit.Keys(kit.MDB_HASH, key), kit.Dict(kit.MDB_TYPE, CMD, kit.MDB_TEXT, m.Option(CMD)))
							n, e := w.Write([]byte(m.Option(CMD) + "\n"))
							m.Debug("%v %v", n, e)
						}
					})
					m.Sleep("300ms")
					m.Cmdy(SESSION, m.Option(kit.MDB_HASH))
				}},
				"bind": {Name: "bind", Help: "绑定", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(SESSION, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)

						input := value["input"].(io.Writer)
						stdin.read(func(buf []byte) {
							m.Debug("input %v", string(buf))
							input.Write(buf)
						})

						output := value["output"].(io.Reader)
						_watch(m, output, os.Stdout, func(buf []byte) {
							m.Debug("output %v", string(buf))
						})
					})
				}},

				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SESSION, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, SESSION, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,status,count,connect")
					if m.Cmdy(mdb.SELECT, SESSION, "", mdb.HASH, kit.MDB_HASH, arg); len(arg) == 0 {
						m.Table(func(index int, value map[string]string, head []string) {
							m.PushButton(kit.Select("绑定", "删除", value[kit.MDB_STATUS] == tcp.CLOSE))
						})
					}
					return
				}

				m.Option(mdb.FIELDS, kit.Select("time,id,type,text", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, SESSION, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.Sort(kit.MDB_ID)
			}},
		},
	}, nil)
}
