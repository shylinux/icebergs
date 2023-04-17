package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _vimer_make(m *ice.Message, dir string, msg *ice.Message) {
	defer m.StatusTimeCount()
	for _, line := range strings.Split(msg.Append(cli.CMD_ERR), ice.NL) {
		if !strings.Contains(line, ice.DF) {
			continue
		}
		if ls := strings.SplitN(line, ice.DF, 4); len(ls) > 3 {
			for i, p := range kit.Split(dir) {
				if strings.HasPrefix(ls[0], p) {
					m.Push(nfs.PATH, p)
					m.Push(nfs.FILE, strings.TrimPrefix(ls[0], p))
					m.Push(nfs.LINE, ls[1])
					m.Push(mdb.TEXT, ls[3])
					break
				} else if i == strings.Count(dir, ice.FS) {
					ps := nfs.SplitPath(m, ls[0])
					m.Push(nfs.PATH, ps[0])
					m.Push(nfs.FILE, ps[1])
					m.Push(nfs.LINE, ls[1])
					m.Push(mdb.TEXT, ls[3])
				}
			}
		}
	}
	kit.If(m.Length() == 0, func() { m.Echo(msg.Append(cli.CMD_OUT)).Echo(msg.Append(cli.CMD_ERR)) })
}

const VIMER = "vimer"

func init() {
	web.Index.MergeCommands(ice.Commands{
		ice.REQUIRE_SRC: {Actions: ice.MergeActions(ctx.CmdAction(), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			web.ShareLocalFile(m, ice.SRC, path.Join(arg...))
		}},
		ice.REQUIRE_USR: {Hand: func(m *ice.Message, arg ...string) { web.ShareLocalFile(m, ice.USR, path.Join(arg...)) }},
	})
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/@key file=main.go line=1 list", Help: "编辑器", Meta: kit.Dict(ctx.STYLE, INNER), Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.PushSearch(mdb.TYPE, nfs.FILE, mdb.NAME, "main", mdb.TEXT, ice.SRC_MAIN_GO)
					m.PushSearch(mdb.TYPE, nfs.FILE, mdb.NAME, "main", mdb.TEXT, ice.SRC_MAIN_SH)
					m.PushSearch(mdb.TYPE, nfs.FILE, mdb.NAME, "main", mdb.TEXT, ice.SRC_MAIN_JS)
					m.PushSearch(mdb.TYPE, web.LINK, mdb.NAME, "admin", mdb.TEXT, web.MergeURL2(m, ice.PS, log.DEBUG, ice.TRUE))
					m.PushSearch(mdb.TYPE, web.LINK, mdb.NAME, m.CommandKey(), mdb.TEXT, m.MergePodCmd("", "", log.DEBUG, ice.TRUE))
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case nfs.MODULE:
					m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
				case nfs.SCRIPT:
					for _, ext := range []string{SH, SHY, PY, JS, CSS, HTML} {
						m.Push(nfs.PATH, kit.ExtChange(m.Option(nfs.FILE), ext))
					}
					m.Option(nfs.DIR_REG, kit.ExtReg(SH, SHY, PY, JS, CSS, HTML))
					nfs.DirDeepAll(m, nfs.SRC, nfs.PWD, nil, nfs.PATH)
				case web.DREAM, XTERM, AUTOGEN:
					m.Cmdy(m.Option(ctx.ACTION), mdb.INPUTS, arg)
				default:
					switch arg[0] {
					case nfs.FILE:
						list := ice.Map{}
						push := func(k, p string) {
							kit.IfNoKey(list, kit.Select(k, k+ice.DF, k != "")+p, func(p string) { m.Push(nfs.PATH, p) })
						}
						mdb.HashSelect(m.Spawn()).TablesLimit(10, func(value ice.Maps) { push("", value[nfs.PATH]) })
						m.Cmd(mdb.SEARCH, mdb.FOREACH, "", ice.OptionFields("type,name,text")).Sort("type,name,text").Table(func(value ice.Maps) {
							switch value[mdb.TYPE] {
							case nfs.FILE:
								push("", value[mdb.TEXT])
							case tcp.GATEWAY:
								push(web.SPACE, value[mdb.TEXT])
							case web.LINK:
								push(web.SPACE, value[mdb.TEXT])
							case web.SERVER:
								push(web.SPACE, value[mdb.TEXT])
							case web.WORKER:
								push(web.SPACE, value[mdb.NAME])
							case ctx.INDEX:
								push(ctx.INDEX, value[mdb.TEXT])
							case ssh.SHELL:
								push(ssh.SHELL, value[mdb.TEXT])
							case cli.OPENS:
								push(cli.OPENS, value[mdb.TEXT])
							}
						})
						for _, p := range kit.Split(kit.Select(m.Option(nfs.PATH), m.Option("paths"))) {
							nfs.DirDeepAll(m.Spawn(), nfs.PWD, p, func(value ice.Maps) { push("", value[nfs.PATH]) }, nfs.PATH)
						}
						m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX)).Table(func(value ice.Maps) { push(ctx.INDEX, value[ctx.INDEX]) })
						m.Cmd(mdb.SEARCH, cli.SYSTEM, cli.OPENS, ice.OptionFields("type,name,text")).Sort("type,name,text").Table(func(value ice.Maps) { push(cli.OPENS, value[nfs.NAME]) })
					default:
						m.Cmdy(INNER, mdb.INPUTS, arg)
					}
				}
			}},
			nfs.SAVE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.CONTENT) == "" {
					m.Option(nfs.CONTENT, m.Cmdx("", TEMPLATE))
				}
				p := path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE))
				switch m.Cmd(nfs.SAVE, p); m.Option(nfs.FILE) {
				case ice.INDEX_CSS, ice.PROTO_JS:
					m.Cmd("", DEVPACK)
				}
				switch arg[0] {
				case nfs.GO:
					defer m.Cmdy(nfs.CAT, p)
					m.Cmd(cli.SYSTEM, "gofmt", "-w", p)
					m.Cmd(cli.SYSTEM, "goimports", "-w", p)
				}
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, arg[0]) }},
			nfs.MODULE: {Name: "create name*=h2 help=示例 type*=Hash,Zone,Data,Code main*=main.go zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, nfs.MODULE, arg)
			}},
			nfs.SCRIPT: {Name: "script file*=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			nfs.REPOS: {Help: "仓库"}, web.SPACE: {Help: "空间"}, web.DREAM: {Help: "空间"}, FAVOR: {Help: "收藏"},
			cli.OPENS:  {Hand: func(m *ice.Message, arg ...string) { cli.Opens(m, arg...) }},
			"listTags": {Help: "生成索引", Hand: func(m *ice.Message, arg ...string) { m.Cmd("web.code.vim.tags", nfs.LOAD) }},

			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPILE: {Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				const app, _app = "usr/publish/contexts.app", "Contents/MacOS/contexts"
				isWebview := func() bool { return strings.HasSuffix(os.Args[0], _app) }
				cmds := []string{COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN}
				if isWebview() {
					m.Option(cli.ENV, "CGO_ENABLED", "1", cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Path(ice.USR_LOCAL_GO_BIN)+ice.DF+kit.Env(cli.PATH))
					cmds = []string{COMPILE, ice.SRC_WEBVIEW_GO, path.Join(app, _app)}
				}
				if msg := m.Cmd(cmds); cli.IsSuccess(msg) {
					if isWebview() {
						m.Go(func() { m.Cmd(cli.SYSTEM, "./bin/ice.bin", cli.FOREVER, cli.DELAY, "300ms", cli.SYSTEM, cli.OPEN, app) })
					}
					m.Go(func() { m.Sleep30ms().Cmd(UPGRADE, cli.RESTART) })
				} else {
					_vimer_make(m, nfs.PWD, msg)
				}
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "源码")) })
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path"), web.DreamAction(), aaa.RoleAction(ctx.COMMAND)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] != ctx.ACTION {
				kit.If(len(arg) > 1, func() { mdb.HashCreate(m.Spawn(), nfs.PATH, path.Join(kit.Slice(arg, 0, 2)...)) })
				m.Action(nfs.MODULE, nfs.SCRIPT, nfs.SAVE, COMPILE, "show", "exec")
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

const TEMPLATE = "template"
const COMPLETE = "complete"
const NAVIGATE = "navigate"

func init() {
	Index.MergeCommands(ice.Commands{TEMPLATE: {Name: "template type name text auto", Help: "模板", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{COMPLETE: {Name: "complete type name text auto", Help: "补全", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{NAVIGATE: {Name: "navigate type name text auto", Help: "跳转", Actions: mdb.RenderAction()}})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) ice.Handler {
		switch sub {
		case TEMPLATE, COMPLETE, NAVIGATE:
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, mdb.CREATE, key, m.PrefixKey()) }
		}
		return nil
	})
}
func Complete(m *ice.Message, text string, data ice.Map) {
	if strings.HasSuffix(text, ice.PT) {
		m.Push(mdb.TEXT, kit.Simple(data[kit.Slice(kit.Split(text, " ."), -1)[0]]))
	} else {
		m.Push(mdb.TEXT, data[""])
		kit.For(data, func(k string, v ice.Any) { m.Push(mdb.TEXT, k) })
	}
}
