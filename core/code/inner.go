package code

import (
	"bufio"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if aaa.Right(m, dir, file) {
		if nfs.IsSourceFile(m, ext) {
			m.Cmdy(nfs.CAT, path.Join(dir, file))
		} else {
			_inner_show(m, ext, file, dir, arg...)
		}
	}
}
func _inner_show(m *ice.Message, ext, file, dir string, arg ...string) {
	if aaa.Right(m, dir, file) {
		m.Cmdy(mdb.RENDER, ext, file, dir, arg)
	}
}
func _inner_exec(m *ice.Message, ext, file, dir string, arg ...string) {
	if aaa.Right(m, dir, file) {
		m.Cmdy(mdb.ENGINE, ext, file, dir, arg)
	}
}
func _inner_make(m *ice.Message, dir string, msg *ice.Message) {
	for _, line := range strings.Split(msg.Append(cli.CMD_ERR), ice.NL) {
		if strings.Contains(line, ice.DF) {
			if ls := strings.SplitN(line, ice.DF, 4); len(ls) > 3 {
				m.Push(nfs.PATH, dir)
				m.Push(nfs.FILE, strings.TrimPrefix(ls[0], dir))
				m.Push(nfs.LINE, ls[1])
				m.Push(mdb.TEXT, ls[3])
			}
		}
	}
	if m.Length() == 0 {
		m.Echo(msg.Append(cli.CMD_OUT))
		m.Echo(msg.Append(cli.CMD_ERR))
	}
	m.StatusTime()
}
func _inner_tags(m *ice.Message, dir string, value string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, nfs.GREP, "^"+value+"\\>", nfs.TAGS, kit.Dict(cli.CMD_DIR, dir)), ice.NL) {
		ls := strings.SplitN(l, ice.TB, 3)
		if len(ls) < 3 {
			continue
		}

		file, ls := ls[1], strings.SplitN(ls[2], ";\"", 2)
		text := strings.TrimSuffix(strings.TrimPrefix(ls[0], "/^"), "$/")
		line := kit.Int(text)

		f, e := nfs.OpenFile(m, kit.Path(dir, file))
		m.Assert(e)
		defer f.Close()

		bio := bufio.NewScanner(f)
		for i := 1; bio.Scan(); i++ {
			if i == line || bio.Text() == text {
				if dir == "" {
					m.PushRecord(kit.Dict(nfs.PATH, path.Dir(file)+ice.PS, nfs.FILE, path.Base(file), nfs.LINE, kit.Format(i), mdb.TEXT, bio.Text()))
				} else {
					m.PushRecord(kit.Dict(nfs.PATH, dir, nfs.FILE, strings.TrimPrefix(file, nfs.PWD), nfs.LINE, kit.Format(i), mdb.TEXT, bio.Text()))
				}
				return
			}
		}
	}
}

const (
	COMMENT  = "comment"
	KEYWORD  = "keyword"
	CONSTANT = "constant"
	DATATYPE = "datatype"
	FUNCTION = "function"
)
const (
	SPLIT   = "split"
	SPACE   = "space"
	OPERATE = "operate"
	PREFIX  = "prefix"
	SUFFIX  = "suffix"
)
const (
	PLUG = "plug"
	SHOW = "show"
	EXEC = "exec"
)
const INNER = "inner"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		INNER: {Name: "inner path=src/@key file=main.go@key line=1 auto", Help: "源代码", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.SRC_MAIN_GO)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.PATH:
					m.Cmdy(nfs.DIR, arg[1:], nfs.DIR_CLI_FIELDS).ProcessAgain()
				case nfs.FILE:
					p := kit.Select(nfs.PWD, arg, 1)
					m.Option(nfs.DIR_ROOT, m.Option(nfs.PATH))
					m.Cmdy(nfs.DIR, kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.FS))+ice.PS, nfs.DIR_CLI_FIELDS).ProcessAgain()
				default:
					m.Cmdy(FAVOR, mdb.INPUTS, arg)
				}
			}},
			mdb.PLUGIN: {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
					m.Echo(kit.Select("{}", m.Config(kit.Keys(PLUG, arg[0]))))
				}
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_inner_show(m, arg[0], arg[1], arg[2])
			}},
			mdb.ENGINE: {Name: "engine", Help: "引擎", Hand: func(m *ice.Message, arg ...string) {
				_inner_exec(m, arg[0], arg[1], arg[2])
			}},
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if _inner_tags(m, m.Option(nfs.PATH), arg[1]); m.Length() == 0 {
					_inner_tags(m, "", arg[1])
				}
			}},

			nfs.GREP: {Name: "grep", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.GREP, m.Option(nfs.PATH), arg[0]).StatusTimeCount(mdb.INDEX, 0)
			}},
			nfs.TAGS: {Name: "tags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				if _inner_tags(m, m.Option(nfs.PATH), arg[0]); m.Length() == 0 {
					_inner_tags(m, "", arg[0])
				}
			}},
			cli.MAKE: {Name: "make", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_inner_make(m, m.Option(nfs.PATH), m.Cmd(cli.SYSTEM, cli.MAKE, arg))
			}},

			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.code.vim.tags", "listTags", arg)
			}},
			NAVIGATE: {Name: "navigate", Help: "跳转", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(NAVIGATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			FAVOR: {Name: "favor", Help: "收藏"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if arg[0] = strings.Split(arg[0], ice.FS)[0]; !strings.HasSuffix(arg[0], ice.PS) && len(arg) == 1 {
				arg[1] = kit.Slice(strings.Split(arg[0], ice.PS), -1)[0]
				arg[0] = strings.TrimSuffix(arg[0], arg[1])
				ctx.ProcessRewrite(m, nfs.PATH, arg[0], nfs.FILE, arg[1])
				return
			}
			if len(arg) < 2 {
				nfs.Dir(m, nfs.PATH)
				return
			}

			list := kit.Simple()
			for k, v := range ice.Info.File {
				if strings.HasPrefix(k, path.Dir(path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))) {
					list = append(list, v)
				}
			}
			m.Option("keys", list)
			m.Option("module", ice.Info.Make.Module)

			m.Option("tabs", m.Config("show.tabs"))
			m.Option("plug", m.Config("show.plug"))
			m.Option("exts", m.Config("show.exts"))

			arg[1] = strings.Split(arg[1], ice.FS)[0]
			_inner_list(m, kit.Ext(arg[1]), arg[1], arg[0])
			ctx.DisplayLocal(m, "")
		}},
	}})
}
func PlugAction() ice.Actions {
	return ice.Actions{
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(PLUG)) }},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
	}
}
func LoadPlug(m *ice.Message, language ...string) {
	for _, language := range language {
		m.Conf(nfs.CAT, kit.Keym(nfs.SOURCE, kit.Ext(language)), ice.TRUE)
		m.Confm(language, kit.Keym(PLUG, PREPARE), func(key string, value interface{}) {
			for _, v := range kit.Simple(value) {
				m.Conf(language, kit.Keym(PLUG, KEYWORD, v), key)
			}
		})
	}
}
