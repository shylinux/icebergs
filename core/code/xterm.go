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
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
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
func (s _xterm) Writeln(data string, arg ...ice.Any) {
	s.Write(kit.Format(data, arg...) + ice.NL)
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
	mdb.HashModify(m, mdb.TIME, m.Time(), web.VIEW, m.Option(ice.MSG_DAEMON))
	return mdb.HashSelectTarget(m, h, func(value ice.Maps) ice.Any {
		text := strings.Split(value[mdb.TEXT], ice.NL)
		ls := kit.Split(strings.Split(kit.Select(nfs.SH, value[mdb.TYPE]), " # ")[0])
		cmd := exec.Command(cli.SystemFind(m, ls[0]), ls[1:]...)
		cmd.Dir = nfs.MkdirAll(m, kit.Path(value[nfs.PATH]))
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, "TERM=xterm")
		tty, err := pty.Start(cmd)
		m.Assert(err)
		m.Go(func() {
			defer tty.Close()
			defer mdb.HashRemove(m, mdb.HASH, h)
			m.Log(cli.START, strings.Join(cmd.Args, ice.SP))
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := tty.Read(buf); !m.Warn(e) && e == nil {
					if _xterm_echo(m, h, string(buf[:n])); len(text) > 0 {
						if cmd := text[0]; text[0] != "" {
							m.Go(func() {
								m.Sleep("10ms")
								tty.Write([]byte(cmd + ice.NL))
							})
						}
						text = text[1:]
					}
				} else {
					_xterm_echo(m, h, "~~~end~~~")
					break
				}
			}
		})
		return _xterm{cmd, tty}
	}).(_xterm)
}
func _xterm_echo(m *ice.Message, h string, str string) {
	m.Options(ice.MSG_DAEMON, mdb.HashSelectField(m, h, web.VIEW))
	web.PushNoticeGrow(m, h, str)
}
func _xterm_cmds(m *ice.Message, h string, cmd string, arg ...ice.Any) {
	kit.If(cmd != "", func() { _xterm_get(m, h).Writeln(cmd, arg...) })
	ctx.ProcessHold(m)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.PushSearch(mdb.TYPE, web.LINK, mdb.NAME, m.CommandKey(), mdb.TEXT, web.MergePodCmds(m, "", "", log.DEBUG, ice.TRUE))
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Cmd(mdb.SEARCH, mdb.FOREACH, ssh.SHELL, ice.OptionFields("type,name,text"), func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == ssh.SHELL, func() { m.Push(arg[0], value[mdb.TEXT]) })
					})
					m.Push(arg[0], BASH, SH)
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)), ice.Info.Hostname)
				case nfs.PATH:
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_REPOS, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_DAEMON, nfs.PATH)
				case nfs.FILE:
					push := func(arg ...string) { m.Push(nfs.FILE, strings.Join(arg, ice.DF)) }
					m.Cmd("", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == web.LAYOUT, func() { push(web.LAYOUT, value[mdb.HASH], value[mdb.NAME]) })
					})
					m.Cmd("", mdb.INPUTS, mdb.TYPE, func(value ice.Maps) { push(ssh.SHELL, value[mdb.TYPE]) })
					m.Cmd(nfs.CAT, kit.HomePath(".bash_history"), func(text string) { push(text) })
					m.Cmd(nfs.CAT, kit.HomePath(".zsh_history"), func(text string) { push(text) })
					m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields(ctx.INDEX), func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
				}
			}},
			web.RESIZE: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.OptionDefault("rows", "24"), m.OptionDefault("cols", "80"))
			}},
			web.INPUT: {Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, "").Write(string(b))
				}
			}},
			web.OUTPUT: {Help: "全屏", Hand: func(m *ice.Message, arg ...string) {
				web.ProcessWebsite(m, "", "", m.OptionSimple(mdb.HASH), ctx.STYLE, web.OUTPUT)
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.CREATE, mdb.TYPE, BASH, m.OptionSimple(mdb.NAME), nfs.PATH, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "命令")) })
			}},
			ctx.PROCESS: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() string { return m.Cmdx("", mdb.CREATE, arg) }, arg...)
			}},
		}, ctx.CmdAction(), ctx.ProcessAction(), web.DreamAction(), mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,path,view,theme")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.PushAction(web.OUTPUT, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
			} else {
				if m.Length() == 0 {
					arg[0] = m.Cmdx("", mdb.CREATE, arg)
					mdb.HashSelect(m, arg[0])
				}
				m.Push(mdb.HASH, arg[0])
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func ProcessXterm(m *ice.Message, cmds, text string, arg ...string) {
	ctx.Process(m, XTERM, []string{mdb.TYPE, cmds, mdb.NAME, kit.Select("", arg, 0), mdb.TEXT, text}, arg...)
}
