package ssh

import (
	"io"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	psh "shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _ssh_watch(m *ice.Message, h string, output io.Writer, input io.Reader) io.Closer {
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
				m.Cmdy(mdb.INSERT, m.Prefix(CHANNEL), kit.Keys(mdb.HASH, h), mdb.LIST, mdb.TYPE, CMD, mdb.TEXT, cmd)
				i = 0
			default:
				if i += n; i >= ice.MOD_BUFS {
					i = 0
				}
			}
		}
	})
	return r
}

const CHANNEL = "channel"

func init() {
	psh.Index.MergeCommands(ice.Commands{
		CHANNEL: {Name: "channel hash id auto", Help: "通道", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, mdb.FOREACH, func(value ice.Map) { kit.Value(value, mdb.STATUS, tcp.CLOSE) })
			}},
			ctx.COMMAND: {Name: "command cmd=pwd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneInsert(m, m.OptionSimple(mdb.HASH), mdb.TYPE, CMD, mdb.TEXT, m.Option(CMD))
				if w, ok := mdb.HashSelectTarget(m, m.Option(mdb.HASH), nil).(io.Writer); ok {
					w.Write([]byte(m.Option(CMD) + lex.NL))
					m.Sleep300ms()
				}
			}},
			mdb.REPEAT: {Help: "执行", Hand: func(m *ice.Message, arg ...string) { m.Cmdy("", ctx.COMMAND, CMD, m.Option(mdb.TEXT)) }},
		}, mdb.HashAction(mdb.FIELDS, "time,hash,status,tty,count,username,hostport", mdb.FIELD, "time,id,type,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Table(func(value ice.Maps) {
					m.PushButton(kit.Select("", ctx.COMMAND, value[mdb.STATUS] == tcp.OPEN), mdb.REMOVE)
				}).Action(mdb.PRUNES)
			} else {
				m.PushAction(mdb.REPEAT).Action(ctx.COMMAND)
			}
		}},
	})
}
