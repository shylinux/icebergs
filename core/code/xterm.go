package code

import (
	"encoding/base64"
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

type _xterm struct {
	*exec.Cmd
	*os.File
}

func (s _xterm) Setsize(rows, cols string) error {
	return pty.Setsize(s.File, &pty.Winsize{Rows: uint16(kit.Int(rows)), Cols: uint16(kit.Int(cols))})
}
func (s _xterm) Write(data string) (int, error) {
	return s.File.Write([]byte(data))
}
func (s _xterm) Close() error {
	return s.Cmd.Process.Kill()
}

func _xterm_get(m *ice.Message, h string) _xterm {
	h = kit.Select(m.Option(mdb.HASH), h)
	m.Assert(h != "")
	t := mdb.HashSelectField(m, m.Option(mdb.HASH, h), mdb.TYPE)
	mdb.HashModify(m, "view", m.Option(ice.MSG_DAEMON))
	return mdb.HashSelectTarget(m, h, func() ice.Any {
		ls := kit.Split(kit.Select(nfs.SH, t))
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "TERM=xterm")
		tty, err := pty.Start(cmd)
		m.Assert(err)
		m.Go(func() {
			// defer mdb.HashSelectUpdate(m, h, func(value ice.Map) { delete(value, mdb.TARGET) })
			defer mdb.HashRemove(m, mdb.HASH, h)
			defer tty.Close()
			// m.Option("log.disable", ice.TRUE)
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) && e == nil {
					m.Option(ice.MSG_DAEMON, mdb.HashSelectField(m, h, "view"))
					web.PushNoticeGrow(m, string(buf[:n]))
				} else {
					break
				}
			}
		})
		return _xterm{cmd, tty}
	}).(_xterm)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
			ctx.PROCESS: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{m.Cmdx("", mdb.CREATE, arg)}
				}
				ctx.ProcessField(m, m.PrefixKey(), arg, arg...)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg).Cmdy(FAVOR, "_system_term", ice.OptionFields(arg[0])).Cut(arg[0]); arg[0] {
				case mdb.TYPE:
					if m.Option(nfs.LINE) != "" && m.Option(nfs.FILE) != "" {
						m.Push(arg[0], "vim +"+m.Option(nfs.LINE)+ice.SP+m.Option(nfs.PATH)+m.Option(nfs.FILE))
					}
					m.Push(arg[0], "bash", "sh")
				case mdb.NAME:
					m.Push(arg[0], ice.Info.Hostname, path.Base(m.Option(mdb.TYPE)))
				}
			}},
			mdb.CREATE: {Name: "create type=sh name text theme:textarea", Hand: func(m *ice.Message, arg ...string) {
				m.Debug("what %v", m.FormatChain())
				mdb.HashCreate(m, mdb.NAME, m.OptionDefault(mdb.NAME, kit.Split(m.Option(mdb.TYPE))[0]), m.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT, web.THEME))
			}},
			"resize": {Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.OptionDefault("rows", "24"), m.OptionDefault("cols", "80"))
			}},
			"input": {Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, "").Write(string(b))
				}
			}},
			"debug": {Help: "日志", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write("cd ~/contexts; tail -f var/log/bench.log" + ice.NL)
				ctx.ProcessHold(m)
			}},
			"proxy": {Help: "代理", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write(kit.Format(`git config --global url."%s".insteadOf "https://shylinux.com"`, m.Option(ice.MSG_USERHOST)) + ice.NL)
				ctx.ProcessHold(m)
			}},
			INSTALL: {Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write(m.Cmdx(PUBLISH, ice.CONTEXTS, INSTALL) + ice.NL)
				ctx.ProcessHold(m)
			}},
			web.WEBSITE: {Help: "网页", Hand: func(m *ice.Message, arg ...string) { web.ProcessWebsite(m, "", "", m.OptionSimple(mdb.HASH)) }},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,view,theme", mdb.TOOLS, FAVOR), ctx.CmdAction(), ctx.ProcessAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.PushAction(web.WEBSITE, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
			} else {
				if m.Length() == 0 {
					arg[0] = m.Cmdx("", mdb.CREATE, mdb.TYPE, arg[0])
					mdb.HashSelect(m, arg[0])
					m.Push(mdb.HASH, arg[0])
				}
				m.Action(INSTALL, "debug", "proxy")
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func _xterm_show(m *ice.Message, cmds, text string, arg ...string) {
	m.Cmdy(ctx.COMMAND, XTERM).Push(ctx.ARGS, kit.Format([]string{m.Cmdx(XTERM, mdb.CREATE, mdb.TYPE, cmds, mdb.NAME, kit.Select("", arg, 0), mdb.TEXT, text)})).ProcessField(XTERM)
}
