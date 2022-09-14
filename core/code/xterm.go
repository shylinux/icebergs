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

func _xterm_get(m *ice.Message, h string) *os.File {
	t := mdb.HashSelectField(m, m.Option(mdb.HASH, h), mdb.TYPE)
	mdb.HashModify(m, mdb.TEXT, m.Option(ice.MSG_DAEMON))
	return mdb.HashTarget(m, h, func() ice.Any {
		ls := kit.Split(kit.Select(nfs.SH, t))
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)
		cmd.Env = append(os.Environ(), "TERM=xterm")

		tty, err := pty.Start(cmd)
		m.Assert(err)

		m.Go(func() {
			m.Logs(cli.DAEMON, ice.CMD, ls)
			defer web.PushNoticeGrow(m, ice.EXIT)
			mdb.HashSelectUpdate(m, h, func(value ice.Map) {
				value["_cmd"] = nfs.NewCloser(func() error {
					m.Option(ice.MSG_DAEMON, mdb.HashSelectField(m, h, mdb.TEXT))
					web.PushNoticeGrow(m, ice.EXIT)
					return cmd.Process.Kill()
				})
			})

			m.Option("log.disable", ice.TRUE)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) {
					m.Option(ice.MSG_DAEMON, mdb.HashSelectField(m, h, mdb.TEXT))
					m.Option(mdb.TEXT, string(buf[:n]))
					web.PushNoticeGrow(m)
				} else {
					break
				}
			}
		})
		return tty
	}).(*os.File)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "终端", Actions: ice.MergeActions(ice.Actions{
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
					m.Push(arg[0], "ice.bin source stdio", "tmux attach -t miss", "node", "python", "bash", "sh")
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)))
				}
			}},
			mdb.CREATE: {Name: "create type name background", Help: "创建"},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectDetail(m, m.Option(mdb.HASH), func(value ice.Map) {
					if c, ok := value["_cmd"].(io.Closer); ok {
						c.Close()
					}
				})
				mdb.HashRemove(m)
			}},
			"resize": {Name: "resize", Help: "大小", Hand: func(m *ice.Message, arg ...string) {
				pty.Setsize(_xterm_get(m, m.Option(mdb.HASH)), &pty.Winsize{Rows: uint16(kit.Int(m.Option("rows"))), Cols: uint16(kit.Int(m.Option("cols")))})
			}},
			"input": {Name: "input", Help: "输入", Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, m.Option(mdb.HASH)).Write(b)
				}
			}},
			INSTALL: {Name: "install", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select(m.Option(mdb.HASH), arg, 0)).Write([]byte(m.Cmdx(PUBLISH, ice.CONTEXTS, INSTALL) + ice.NL))
				m.ProcessHold()
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.PRUNES)
			} else {
				m.Action("full", INSTALL)
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func ProcessXterm(m *ice.Message, args []string, arg ...string) {
	if len(arg) == 0 || arg[0] != ice.RUN {
		args = []string{m.Cmdx("web.code.xterm", mdb.CREATE, args)}
	}
	ctx.ProcessField(m, "web.code.xterm", args, arg...)
}
