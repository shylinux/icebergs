package ssh

import (
	"io"
	"net"
	"os/exec"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
)

func _ssh_exec(m *ice.Message, cmd string, arg []string, env []string, input io.Reader, output io.Writer, done func()) {
	m.Log_IMPORT(CMD, cmd, ARG, arg, ENV, env)
	c := exec.Command(cmd, arg...)
	// c.Env = env

	c.Stdin = input
	c.Stdout = output
	c.Stderr = output

	m.Assert(c.Start())

	m.Go(func() {
		defer done()
		c.Process.Wait()
	})
}
func _ssh_close(m *ice.Message, c net.Conn, channel ssh.Channel) {
	defer channel.Close()
	channel.Write([]byte(m.Conf(SERVICE, kit.Keym(GOODBYE))))
}
func _ssh_watch(m *ice.Message, meta map[string]string, h string, input io.Reader, output io.Writer) {
	r, w := io.Pipe()
	bio := io.TeeReader(input, w)
	m.Go(func() { io.Copy(output, r) })

	i, buf := 0, make([]byte, ice.MOD_BUFS)
	m.Go(func() {
		for {
			n, e := bio.Read(buf[i:])
			if e != nil {
				break
			}

			switch buf[i] {
			case '\r', '\n':
				cmd := strings.TrimSpace(string(buf[:i]))
				m.Log_IMPORT(tcp.HOSTNAME, meta[tcp.HOSTNAME], aaa.USERNAME, meta[aaa.USERNAME], CMD, cmd)
				m.Cmdy(mdb.INSERT, CHANNEL, kit.Keys(kit.MDB_HASH, h), mdb.LIST, kit.MDB_TYPE, CMD, kit.MDB_TEXT, cmd)
				i = 0
			default:
				if i += n; i >= ice.MOD_BUFS {
					i = 0
				}
			}
		}
	})
}

const CHANNEL = "channel"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CHANNEL: {Name: "channel", Help: "通道", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			CHANNEL: {Name: "channel hash id auto command prunes", Help: "通道", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, CHANNEL, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,username,hostport,tty,count")
					m.Cmdy(mdb.PRUNES, SERVICE, "", mdb.HASH, kit.MDB_STATUS, tcp.ERROR)
					m.Cmdy(mdb.PRUNES, CHANNEL, "", mdb.HASH, kit.MDB_STATUS, tcp.CLOSE)
				}},
				mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(CHANNEL, kit.MDB_ACTION, ctx.COMMAND, CMD, m.Option(kit.MDB_TEXT))
				}},
				ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, CHANNEL, kit.Keys(kit.MDB_HASH, m.Option(kit.MDB_HASH)), mdb.LIST, kit.MDB_TYPE, CMD, kit.MDB_TEXT, m.Option(CMD))
					m.Richs(CHANNEL, "", m.Option(kit.MDB_HASH), func(key string, value map[string]interface{}) {
						if w, ok := kit.Value(value, kit.Keym(INPUT)).(io.Writer); ok {
							w.Write([]byte(m.Option(CMD) + "\n"))
						}
					})
					m.ProcessRefresh("300ms")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 { // 通道列表
					m.Fields(len(arg), "time,hash,status,username,hostport,tty,count")
					if m.Cmdy(mdb.SELECT, CHANNEL, "", mdb.HASH); len(arg) == 0 {
						m.Table(func(index int, value map[string]string, head []string) {
							m.PushButton(kit.Select("", ctx.COMMAND, value[kit.MDB_STATUS] == tcp.OPEN), mdb.REMOVE)
						})
					}
					return
				}

				// 通道命令
				m.Fields(len(arg[1:]), "time,id,type,text")
				m.Cmdy(mdb.SELECT, CHANNEL, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.PushAction(mdb.REPEAT)
			}},
		},
	})
}
