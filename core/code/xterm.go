package code

import (
	"encoding/base64"
	"os"
	"os/exec"
	"sync"
	"time"

	pty "shylinux.com/x/creackpty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const XTERM = "xterm"

func init() {
	cache := sync.Map{}
	add := func(m *ice.Message, key string) string {
		cmd := exec.Command("/bin/sh")
		cmd.Env = append(os.Environ(), "TERM=xterm")
		tty, err := pty.Start(cmd)
		m.Assert(err)
		cache.Store(key, tty)

		m.Go(func() {
			defer m.Cmd(m.PrefixKey(), mdb.PRUNES)
			defer cache.Delete(key)
			buf := make([]byte, 1024)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) {
					m.Optionv(ice.MSG_OPTS, kit.Simple("hash"))
					m.Option(mdb.HASH, key)
					m.Option(ice.MSG_DAEMON, m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, key, mdb.META, mdb.TEXT)))
					m.PushNoticeGrow(kit.Format(kit.Dict(mdb.TYPE, "data", mdb.TEXT, base64.StdEncoding.EncodeToString(buf[:n]))))
				} else {
					break
				}
			}
			m.PushNoticeGrow(kit.Format(kit.Dict(mdb.TYPE, "exit")))
		})
		return key
	}
	get := func(m *ice.Message, key string) *os.File {
		if w, ok := cache.Load(key); ok {
			if w, ok := w.(*os.File); ok {
				return w
			}
		}
		add(m, key)
		if w, ok := cache.Load(key); ok {
			if w, ok := w.(*os.File); ok {
				return w
			}
		}
		return nil
	}

	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash id auto prunes", Help: "终端", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create type name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TEXT, m.Option(ice.MSG_DAEMON)) != "" {
					m.Echo(add(m, m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple("type,name,text"))))
				}
			}},
			"resize": {Name: "resize", Help: "大小", Hand: func(m *ice.Message, arg ...string) {
				pty.Setsize(get(m, m.Option(mdb.HASH)), &pty.Winsize{Rows: uint16(kit.Int(m.Option("rows"))), Cols: uint16(kit.Int(m.Option("cols")))})
			}},
			"select": {Name: "select", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), mdb.TEXT, m.Option(ice.MSG_DAEMON))
			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.MODIFY, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH), mdb.TIME, m.Time())
				get(m, m.Option(mdb.HASH)).Write([]byte(arg[0]))
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(m.CommandKey()).Tables(func(value ice.Maps) {
					if _, ok := cache.Load(value[mdb.HASH]); !ok || kit.Time(m.Time())-kit.Time(value[mdb.TIME]) > int64(time.Hour) {
						m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, mdb.HASH, value[mdb.HASH])
					}
				})
			}},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.OptionFields("time,hash,type,name,text")
			}
			mdb.ZoneSelect(m, kit.Slice(arg, 0, 2)...)
			m.DisplayLocal("")
		}},
	})
}
