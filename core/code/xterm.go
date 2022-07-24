package code

import (
	"encoding/base64"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pty "shylinux.com/x/creackpty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const XTERM = "xterm"

func init() {
	cache := sync.Map{}
	add := func(m *ice.Message, key string) string {
		cmd := exec.Command(kit.Select("/bin/sh", m.Option(mdb.TYPE)))
		cmd.Env = append(os.Environ(), "TERM=xterm")

		tty, err := pty.Start(cmd)
		m.Assert(err)
		cache.Store(key, tty)

		m.Go(func() {
			defer m.Cmd(m.PrefixKey(), mdb.PRUNES)
			defer cache.Delete(key)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) {
					m.Option(mdb.HASH, key)
					m.Optionv(ice.MSG_OPTS, kit.Simple(mdb.HASH))
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
		XTERM: {Name: "xterm hash id auto", Help: "终端", Actions: ice.MergeAction(ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if mdb.HashInputs(m, arg); arg[0] == mdb.TYPE {
					m.Push(mdb.TYPE, "/usr/bin/python")
					m.Push(mdb.TYPE, "/usr/bin/node")
					m.Push(mdb.TYPE, "/bin/bash")
					m.Push(mdb.TYPE, "/bin/sh")
				}
			}},
			mdb.CREATE: {Name: "create type name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TEXT, m.Option(ice.MSG_DAEMON)) != "" {
					m.Echo(add(m, mdb.HashCreate(m, m.OptionSimple("type,name,text")).Result()))
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if w, ok := cache.Load(m.Option(mdb.HASH)); ok {
					if w, ok := w.(*os.File); ok {
						w.Close()
					}
					cache.Delete(m.Option(mdb.HASH))
				}
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), arg)
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m).Tables(func(value ice.Maps) {
					if _, ok := cache.Load(value[mdb.HASH]); !ok || kit.Time(m.Time())-kit.Time(value[mdb.TIME]) > int64(time.Hour) {
						m.Cmd(m.PrefixKey(), mdb.REMOVE, kit.Dict(value))
					}
				})
			}},
			"resize": {Name: "resize", Help: "大小", Hand: func(m *ice.Message, arg ...string) {
				pty.Setsize(get(m, m.Option(mdb.HASH)), &pty.Winsize{Rows: uint16(kit.Int(m.Option("rows"))), Cols: uint16(kit.Int(m.Option("cols")))})
			}},
			"rename": {Name: "rename", Help: "重命名", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), arg)
			}},
			"select": {Name: "select", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.TEXT, m.Option(ice.MSG_DAEMON))
				m.Cmd(m.PrefixKey(), "input", arg)
			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), mdb.TIME, m.Time())
				get(m, m.Option(mdb.HASH)).Write([]byte(strings.Join(arg, "")))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,extra"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
			m.DisplayLocal("")
		}},
	})
}
