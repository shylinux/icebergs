package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _vimer_make(m *ice.Message, dir string, msg *ice.Message) {
	defer m.StatusTime()
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
				} else if n := 2; i == strings.Count(dir, ice.FS) {
					if strings.HasPrefix(ls[0], "src/") {
						n = 1
					}
					m.Push(nfs.PATH, kit.Join(kit.Slice(kit.Split(ls[0], ice.PS, ice.PS), 0, n), ice.PS)+ice.PS)
					m.Push(nfs.FILE, kit.Join(kit.Slice(kit.Split(ls[0], ice.PS, ice.PS), n), ice.PS))
					m.Push(nfs.LINE, ls[1])
					m.Push(mdb.TEXT, ls[3])
				}
			}
		}
	}
	if m.Length() == 0 {
		m.Echo(msg.Append(cli.CMD_OUT)).Echo(msg.Append(cli.CMD_ERR))
	}
}

const VIMER = "vimer"

func init() {
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/@key file=main.go line=1 list", Help: "编辑器", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case web.DREAM, AUTOGEN, XTERM:
					m.Cmdy(m.Option(ctx.ACTION), mdb.INPUTS, arg)
				case web.WEBSITE, nfs.SCRIPT:
					m.Cmdy(COMPLETE, mdb.FOREACH, kit.Select("", arg, 1), m.Option(ctx.ACTION))
				case "extension":
					nfs.DirDeepAll(m, "usr/volcanos/plugin/local/code/", "inner/", nil, nfs.PATH)
				default:
					switch arg[0] {
					case ctx.INDEX:
						m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX))
					case nfs.PATH:
						m.Cmdy(INNER, mdb.INPUTS, arg).Cut("path,size,time")
					case nfs.FILE:
						list, p := ice.Map{}, m.Option(nfs.PATH)
						m.Cmd(FAVOR, "_recent_file").Tables(func(value ice.Maps) {
							kit.IfNoKey(list, value[nfs.PATH]+value[nfs.FILE], func(p string) { m.Push(nfs.PATH, p) })
						})
						for _, p := range kit.Split(kit.Select(m.Option(nfs.PATH), m.Option("paths"))) {
							nfs.DirDeepAll(m, nfs.PWD, p, func(value ice.Maps) { kit.IfNoKey(list, value[nfs.PATH], func(p string) { m.Push(nfs.PATH, p) }) }, nfs.PATH)
						}
						m.Cmd(mdb.RENDER, kit.Ext(m.Option(nfs.FILE)), m.Option(nfs.FILE), p).Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, kit.Format("line:%s:%s:%s", value[nfs.LINE], value["kind"], value[mdb.NAME]))
						})
						m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX)).Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, "index:"+value[ctx.INDEX])
						})
						m.Cmd(FAVOR, "_system_app").Tables(func(value ice.Maps) {
							m.Push(nfs.PATH, "_open:"+strings.ToLower(kit.Select(value[mdb.TEXT], value[mdb.NAME])))
						})
					}
				}
			}},
			nfs.SAVE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.CONTENT) == "" {
					m.Option(nfs.CONTENT, m.Cmdx("", TEMPLATE))
				}
				switch m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE))); m.Option(nfs.FILE) {
				case "proto.js", "page/index.css":
					m.Cmd("", DEVPACK)
				}
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, arg[0]) }},
			nfs.SCRIPT: {Name: "script file*=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			web.WEBSITE: {Name: "website file*=hi.zml", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx("", TEMPLATE))
			}},
			web.DREAM: {Name: "dream name*=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			"_open": {Help: "打开", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.DAEMON, cli.OPEN, "-a", kit.Split(arg[0], ice.PT, ice.PT)[0])
			}},
			XTERM: {Name: "xterm type=sh name text", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(XTERM, mdb.CREATE, arg)
			}}, FAVOR: {Help: "收藏"},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			DEVPACK: {Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			BINPACK: {Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.CREATE)
				m.Cmdy(AUTOGEN, BINPACK)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			AUTOGEN: {Name: "create name*=h2 help=示例 type*=Zone,Hash,Data,Code main*=main.go zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, nfs.MODULE, arg)
			}},
			COMPILE: {Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				cmds := []string{COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN}
				if strings.HasSuffix(os.Args[0], "contexts.app/Contents/MacOS/contexts") {
					m.Option(cli.ENV, "CGO_ENABLED", "1", cli.HOME, kit.Env(cli.HOME), cli.PATH, kit.Path("usr/local/go/bin")+ice.DF+kit.Env(cli.PATH))
					cmds = []string{COMPILE, "src/webview.go", "usr/publish/contexts.app/Contents/MacOS/contexts"}
				}
				if msg := m.Cmd(cmds); cli.IsSuccess(msg) {
					m.Cmd(UPGRADE, cli.RESTART)
				} else {
					_vimer_make(m, nfs.PWD, msg)
				}
			}},
			PUBLISH: {Help: "发布", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(PUBLISH, ice.CONTEXTS) }},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "源码")) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[1] == m.CommandKey(), func() { web.ProcessWebsite(m, m.Option(mdb.NAME), m.PrefixKey()) })
			}},
		}, web.DreamAction(), aaa.RoleAction(ctx.COMMAND)), Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] != ctx.ACTION {
				m.Action(AUTOGEN, nfs.SCRIPT, web.DREAM, web.WEBSITE, nfs.SAVE, COMPILE)
				m.Options("tabs", m.Config("show.tabs"), "plug", m.Config("show.plug"), "exts", m.Config("show.exts"))
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
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		switch sub {
		case TEMPLATE, COMPLETE, NAVIGATE:
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, mdb.CREATE, key, m.PrefixKey()) }, nil
		}
		return nil, nil
	})
}
func Complete(m *ice.Message, text string, data ice.Map) {
	if strings.HasSuffix(text, ice.PT) {
		m.Push(mdb.TEXT, kit.Simple(data[kit.Slice(kit.Split(text, " ."), -1)[0]]))
	} else {
		m.Push(mdb.TEXT, data[""])
		kit.Fetch(data, func(k string, v ice.Any) { m.Push(mdb.TEXT, k) })
	}
}
