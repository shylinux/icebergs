package ssh

import (
	"io"
	"net"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
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
				m.Cmdy(mdb.INSERT, CHANNEL, kit.Keys(mdb.HASH, h), mdb.LIST, mdb.TYPE, CMD, mdb.TEXT, cmd)
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
	psh.Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CHANNEL: {Name: "channel", Help: "通道", Value: kit.Data(
			mdb.FIELD, "time,hash,status,username,hostport,tty,count",
		)},
	}, Commands: map[string]*ice.Command{
		CHANNEL: {Name: "channel hash id auto", Help: "通道", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Richs(CHANNEL, "", mdb.FOREACH, func(key string, value ice.Map) {
					kit.Value(value, kit.Keym(mdb.STATUS), tcp.CLOSE)
				})
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.FIELD))
				m.Cmdy(mdb.PRUNES, SERVICE, "", mdb.HASH, mdb.STATUS, tcp.ERROR)
				m.Cmdy(mdb.PRUNES, CHANNEL, "", mdb.HASH, mdb.STATUS, tcp.CLOSE)
			}},
			mdb.REPEAT: {Name: "repeat", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(CHANNEL, ctx.ACTION, ctx.COMMAND, CMD, m.Option(mdb.TEXT))
			}},
			ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, CHANNEL, kit.Keys(mdb.HASH, m.Option(mdb.HASH)),
					mdb.LIST, mdb.TYPE, CMD, mdb.TEXT, m.Option(CMD))
				m.Richs(CHANNEL, "", m.Option(mdb.HASH), func(key string, value ice.Map) {
					if w, ok := kit.Value(value, kit.Keym(INPUT)).(io.Writer); ok {
						w.Write([]byte(m.Option(CMD) + ice.NL))
					}
				})
				m.ProcessRefresh300ms()
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 通道列表
				m.Action(mdb.PRUNES)
				mdb.HashSelect(m, arg...)
				m.Set(ice.MSG_APPEND, ctx.ACTION)
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushButton(kit.Select("", ctx.COMMAND, value[mdb.STATUS] == tcp.OPEN), mdb.REMOVE)
				})
				return
			}

			// 通道命令
			m.Action(ctx.COMMAND)
			m.Fields(len(arg[1:]), "time,id,type,text")
			mdb.ZoneSelect(m, arg...)
			m.PushAction(mdb.REPEAT)
		}},
	}})
}
