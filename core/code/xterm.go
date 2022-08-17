package code

import (
	"encoding/base64"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	pty "shylinux.com/x/creackpty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _xterm_socket(m *ice.Message, h, t string) {
	defer mdb.RLock(m)()
	m.Option(ice.MSG_DAEMON, m.Conf("", kit.Keys(mdb.HASH, h, mdb.META, mdb.TEXT)))
	m.Option(mdb.TEXT, t)
}
func _xterm_get(m *ice.Message, h string, must bool) *os.File {
	if f, ok := mdb.HashTarget(m, h, func() ice.Any {
		if !must {
			return nil
		}

		ls := kit.Split(kit.Select("sh", m.Option(mdb.TYPE)))
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)
		cmd.Env = append(os.Environ(), "TERM=xterm")
		m.Option(mdb.HASH, h)

		tty, err := pty.Start(cmd)
		m.Assert(err)

		m.Go(func() {
			mdb.HashSelectUpdate(m, h, func(value ice.Map) {
				value["_cmd"] = nfs.NewCloser(func() error { return cmd.Process.Kill() })
			})
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) {
					_xterm_socket(m, h, base64.StdEncoding.EncodeToString(buf[:n]))
					web.PushNoticeGrow(m, "data")
				} else {
					break
				}
			}
			web.PushNoticeGrow(m, "exit")
		})
		return tty
	}).(*os.File); m.Warn(!ok, ice.ErrNotValid, f) {
		mdb.HashSelectUpdate(m, h, func(value ice.Map) { delete(value, mdb.TARGET) })
		return nil
	} else {
		return f
	}
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash refresh", Help: "终端", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectValue(m, func(value ice.Map) {
					if c, ok := value["_cmd"].(io.Closer); ok {
						c.Close()
					}
				})
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], "python", "node", "bash", "sh")
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)))
				}
			}},
			mdb.CREATE: {Name: "create type name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, arg, mdb.TEXT, m.Option(ice.MSG_DAEMON))
				_xterm_get(m, m.Result(), true)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if f := _xterm_get(m, m.Option(mdb.HASH), false); f != nil {
					f.Close()
				}
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(mdb.HASH), arg)
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
			}},
			"resize": {Name: "resize", Help: "大小", Hand: func(m *ice.Message, arg ...string) {
				pty.Setsize(_xterm_get(m, m.Option(mdb.HASH), true), &pty.Winsize{Rows: uint16(kit.Int(m.Option("rows"))), Cols: uint16(kit.Int(m.Option("cols")))})
			}},
			"rename": {Name: "rename", Help: "重命名", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, arg)
			}},
			"select": {Name: "select", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, mdb.TEXT, m.Option(ice.MSG_DAEMON))
				m.Cmd("", "input", arg)
			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, mdb.TIME, m.Time())
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); m.Assert(e) {
					_xterm_get(m, m.Option(mdb.HASH), true).Write(b)
				}
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,extra"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
			ctx.DisplayLocal(m, "")
		}},
	})
}
