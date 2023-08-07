package code

import (
	"encoding/base64"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
)

func _xterm_get(m *ice.Message, h string) xterm.XTerm {
	h = kit.Select(m.Option(mdb.HASH), h)
	m.Assert(h != "")
	m.Option("skip.important", ice.TRUE)
	if m.Option(ice.MSG_USERPOD) == "" {
		mdb.HashModify(m, mdb.TIME, m.Time(), cli.DAEMON, kit.Keys(m.Option(ice.MSG_DAEMON)))
	} else {
		mdb.HashModify(m, mdb.TIME, m.Time(), cli.DAEMON, kit.Keys(kit.Slice(kit.Simple(m.Optionv("__target")), 0, -1), m.Option(ice.MSG_DAEMON)))
	}
	return mdb.HashSelectTarget(m, h, func(value ice.Maps) ice.Any {
		text := strings.Split(value[mdb.TEXT], lex.NL)
		ls := kit.Split(strings.Split(kit.Select(ISH, value[mdb.TYPE]), " # ")[0])
		kit.If(value[nfs.PATH] != "" && !strings.HasSuffix(value[nfs.PATH], nfs.PS), func() { value[nfs.PATH] = path.Dir(value[nfs.PATH]) })
		term, e := xterm.Command(m, value[nfs.PATH], kit.Select(ls[0], cli.SystemFind(m, ls[0])), ls[1:]...)
		if m.Warn(e) {
			return nil
		}
		m.Go(func() {
			defer term.Close()
			defer mdb.HashRemove(m, mdb.HASH, h)
			m.Log(cli.START, strings.Join(ls, lex.SP))
			buf := make([]byte, ice.MOD_BUFS)
			for {
				if n, e := term.Read(buf); !m.Warn(e) && e == nil {
					if _xterm_echo(m, h, string(buf[:n])); len(text) > 0 {
						if cmd := text[0]; text[0] != "" {
							m.Go(func() {
								m.Sleep30ms()
								term.Writeln(cmd)
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
		return term
	}).(xterm.XTerm)
}
func _xterm_echo(m *ice.Message, h string, str string) {
	m.Options(ice.MSG_COUNT, "0", ice.LOG_DISABLE, ice.TRUE, "__target", "", ice.MSG_DAEMON, mdb.HashSelectField(m, h, cli.DAEMON))
	web.PushNoticeGrow(m, h, str)
}
func _xterm_cmds(m *ice.Message, h string, cmd string, arg ...ice.Any) {
	kit.If(cmd != "", func() { _xterm_get(m, h).Writeln(cmd, arg...) })
	ctx.ProcessHold(m)
}

const XTERM = "xterm"

func init() {
	Index.MergeCommands(ice.Commands{
		XTERM: {Name: "xterm hash auto install terminal", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Cmd("").Length() == 0, func() { m.Cmd("", mdb.CREATE, mdb.TYPE, ISH) })
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "shell" {
					m.PushSearch(mdb.TYPE, ssh.SHELL, mdb.NAME, SH, mdb.TEXT, "/bin/sh")
				}
				mdb.IsSearchPreview(m, arg, func() []string { return []string{ssh.SHELL, SH, kit.Select("/bin/sh", os.Getenv("SHELL"))} })
				mdb.IsSearchPreview(m, arg, func() []string { return []string{ssh.SHELL, ISH, "/bin/ish"} })
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], "/bin/ish", kit.Select("/bin/sh", os.Getenv("SHELL")))
					m.Cmd(mdb.SEARCH, mdb.FOREACH, ssh.SHELL, ice.OptionFields("type,name,text"), func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == ssh.SHELL, func() { m.Push(arg[0], value[mdb.TEXT]) })
					})
				case mdb.NAME:
					m.Push(arg[0], path.Base(m.Option(mdb.TYPE)), ice.Info.Hostname)
				case nfs.PATH:
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_REPOS, nfs.PATH)
					m.Cmdy(nfs.DIR, ice.USR_LOCAL_DAEMON, nfs.PATH)
				case nfs.FILE:
					push := func(arg ...string) { m.Push(nfs.FILE, strings.Join(arg, nfs.DF)) }
					m.Cmd("", func(value ice.Maps) {
						kit.If(value[mdb.TYPE] == web.LAYOUT, func() { push(web.LAYOUT, value[mdb.HASH], value[mdb.NAME]) })
					})
					m.Cmd("", mdb.INPUTS, mdb.TYPE, func(value ice.Maps) { push(ssh.SHELL, value[mdb.TYPE]) })
					m.Cmd(nfs.CAT, kit.HomePath(".bash_history"), func(text string) { push(text) })
					m.Cmd(nfs.CAT, kit.HomePath(".zsh_history"), func(text string) { push(text) })
					m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields(ctx.INDEX), func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				m.ProcessRewrite(mdb.HASH, mdb.HashCreate(m, arg))
			}},
			web.RESIZE: {Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, "").Setsize(m.OptionDefault("rows", "24"), m.OptionDefault("cols", "80"))
			}},
			web.INPUT: {Hand: func(m *ice.Message, arg ...string) {
				if b, e := base64.StdEncoding.DecodeString(strings.Join(arg, "")); !m.Warn(e) {
					_xterm_get(m, "").Write(b)
				}
			}},
			web.OUTPUT: {Help: "全屏", Hand: func(m *ice.Message, arg ...string) {
				web.ProcessPodCmd(m, "", "", m.OptionSimple(mdb.HASH), ctx.STYLE, web.OUTPUT)
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "终端")) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, []string{}, arg...) }},
			ctx.PROCESS: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 1 {
					ctx.ProcessField(m, m.PrefixKey(), arg, arg...)
				} else {
					ctx.ProcessField(m, m.PrefixKey(), func() string { return m.Cmdx("", mdb.CREATE, arg) }, arg...)
				}
			}},
			"install": {Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				_xterm_get(m, kit.Select("", arg, 0)).Write([]byte(m.Cmdx(PUBLISH, ice.CONTEXTS, ice.APP, kit.Dict("format", "raw")) + ice.NL))
				ctx.ProcessHold(m)
			}},
			"terminal": {Help: "本机", Hand: func(m *ice.Message, arg ...string) {
				if h := kit.Select(m.Option(mdb.HASH), arg, 0); h == "" {
					cli.Opens(m, "Terminal.app")
				} else {
					msg := m.Cmd("", h)
					cli.OpenCmds(m, msg.Append(mdb.TYPE))
				}
				m.ProcessHold()
			}},
		}, ctx.CmdAction(), ctx.ProcessAction(), mdb.ImportantHashAction(mdb.FIELD, "time,hash,type,name,text,path,theme,daemon")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				if m.Length() == 0 {
					m.Action(mdb.CREATE)
				} else {
					m.PushAction(web.OUTPUT, mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES)
				}
			} else {
				if m.Length() == 0 {
					arg[0] = m.Cmdx("", mdb.CREATE, kit.SimpleKV("type,name,text,path", arg))
					mdb.HashSelect(m, arg[0])
				}
				m.Push(mdb.HASH, arg[0])
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

func ProcessXterm(m *ice.Message, cmds, text string, arg ...string) {
	ctx.Process(m, XTERM, func() []string {
		if ls := kit.Simple(kit.UnMarshal(m.Option(ctx.ARGS))); len(ls) > 0 {
			return ls
		} else {
			return []string{mdb.TYPE, cmds, mdb.NAME, kit.Select("", arg, 0), mdb.TEXT, text}
		}
	}, arg...)
}
