package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func _inner_list(m *ice.Message, ext, file, dir string, arg ...string) {
	if !m.Right(dir, file) {
		return // 没有权限
	}
	if m.Conf(nfs.CAT, kit.Keym(ssh.SOURCE, ext)) == ice.TRUE {
		m.Cmdy(nfs.CAT, path.Join(dir, file))
	} else {
		_inner_show(m, ext, file, dir, arg...)
	}
}
func _inner_show(m *ice.Message, ext, file, dir string, arg ...string) {
	if !m.Right(dir, file) {
		return // 没有权限
	}
	m.Cmdy(mdb.RENDER, ext, file, dir, arg)
}
func _inner_exec(m *ice.Message, ext, file, dir string, arg ...string) {
	if !m.Right(dir, file) {
		return // 没有权限
	}
	defer m.StatusTime()
	if m.Cmdy(mdb.ENGINE, ext, file, dir, arg); m.Result() != "" {
		return // 执行成功
	}
	if ls := kit.Simple(m.Configv(kit.Keys(EXEC, ext))); len(ls) > 0 {
		m.Cmdy(cli.SYSTEM, ls, file, kit.Dict(cli.CMD_DIR, dir)).SetAppend()
	}
}
func _inner_make(m *ice.Message, msg *ice.Message) {
	for _, line := range strings.Split(msg.Append(cli.CMD_ERR), ice.NL) {
		if strings.Contains(line, ice.DF) {
			if ls := strings.SplitN(line, ice.DF, 4); len(ls) > 3 {
				m.Push(nfs.FILE, strings.TrimPrefix(ls[0], m.Option(nfs.PATH)))
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

func LoadPlug(m *ice.Message, language ...string) {
	for _, language := range language {
		m.Conf(nfs.CAT, kit.Keym(nfs.SOURCE, kit.Ext(language)), ice.TRUE)
		m.Confm(language, kit.Keym(PLUG, PREPARE), func(key string, value ice.Any) {
			for _, v := range kit.Simple(value) {
				m.Conf(language, kit.Keym(PLUG, KEYWORD, v), key)
			}
		})
	}
}
func PlugAction() map[string]*ice.Action {
	return map[string]*ice.Action{
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(PLUG)) }},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
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
	PREFIX  = "prefix"
	SUFFIX  = "suffix"
	SPACE   = "space"
	OPERATE = "operate"
)
const (
	PLUG = "plug"
	SHOW = "show"
	EXEC = "exec"
)
const INNER = "inner"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		INNER: {Name: "inner path=src/@key file=main.go line=1 auto", Help: "源代码", Meta: kit.Dict(ice.DisplayLocal("")), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.PLUGIN: {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
					m.Echo(kit.Select("{}", m.Config(kit.Keys(PLUG, arg[0]))))
				}
				m.Set(ice.MSG_STATUS)
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_inner_show(m, arg[0], arg[1], arg[2])
			}},
			mdb.ENGINE: {Name: "engine", Help: "引擎", Hand: func(m *ice.Message, arg ...string) {
				_inner_exec(m, arg[0], arg[1], arg[2])
			}},
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				return
				m.Option(nfs.DIR_ROOT, arg[2])
				m.Option(cli.CMD_DIR, kit.Path(arg[2]))
				m.Cmdy(mdb.SEARCH, arg[0], arg[1], arg[2])
				m.Cmd(FAVOR, arg[1], ice.OptionFields("")).Tables(func(value map[string]string) {
					if p := path.Join(value[nfs.PATH], value[nfs.FILE]); strings.HasPrefix(p, m.Option(nfs.PATH)) {
						m.Push(nfs.FILE, strings.TrimPrefix(p, m.Option(nfs.PATH)))
						m.Push(nfs.LINE, value[nfs.LINE])
						m.Push(mdb.TEXT, value[mdb.TEXT])
					}
				})
				if m.StatusTimeCount(mdb.INDEX, 0); m.Length() == 0 {
					m.Cmdy(INNER, nfs.GREP, arg[1])
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.ACTION) == "website" {
					switch arg[0] {
					case nfs.FILE:
						m.Cmdy(nfs.DIR, nfs.PWD, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_ROOT, "src/website/")).ProcessAgain()
					}
					return
				}

				switch arg[0] {
				case cli.MAIN:
					m.Cmdy(nfs.DIR, ice.SRC, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_REG, `.*\.go`)).ProcessAgain()
				case mdb.ZONE:
					m.Option(nfs.DIR_ROOT, ice.SRC)
					m.Option(nfs.DIR_TYPE, nfs.DIR)
					m.Cmdy(nfs.DIR, nfs.PWD, mdb.NAME).RenameAppend(mdb.NAME, mdb.ZONE)
				case nfs.PATH:
					m.Cmdy(nfs.DIR, arg[1:], nfs.DIR_CLI_FIELDS).ProcessAgain()
				case nfs.FILE:
					p := kit.Select(nfs.PWD, arg, 1)
					m.Option(nfs.DIR_ROOT, m.Option(nfs.PATH))
					m.Cmdy(nfs.DIR, kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.FS))+ice.PS, nfs.DIR_CLI_FIELDS).ProcessAgain()
				case "url":
					m.Option(nfs.DIR_ROOT, "usr/volcanos/plugin/local/code/")
					m.Cmdy(nfs.DIR, nfs.PWD, nfs.DIR_CLI_FIELDS, kit.Dict(nfs.DIR_DEEP, ice.TRUE)).ProcessAgain()
				default:
					m.Cmdy(FAVOR, mdb.INPUTS, arg)
				}
			}},

			nfs.TAGS: {Name: "tags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {}},
			nfs.GREP: {Name: "grep", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.GREP, m.Option(nfs.PATH), arg[0])
				m.StatusTimeCount(mdb.INDEX, 0)
			}},
			cli.MAKE: {Name: "make", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_inner_make(m, m.Cmd(cli.SYSTEM, cli.MAKE, arg))
			}},
			FAVOR: {Name: "favor", Help: "收藏"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if arg[0] = strings.Split(arg[0], ice.FS)[0]; !strings.HasSuffix(arg[0], ice.PS) {
				arg[1] = kit.Slice(strings.Split(arg[0], ice.PS), -1)[0]
				arg[0] = strings.TrimSuffix(arg[0], arg[1])
				m.ProcessRewrite(nfs.PATH, arg[0], nfs.FILE, arg[1])
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

			m.Option("plug", "web.chat.website,web.dream")
			m.Option("exts", "inner/search.js?a=1,inner/favor.js,inner/template.js")

			arg[1] = strings.Split(arg[1], ice.FS)[0]
			if _inner_list(m, kit.Ext(arg[1]), arg[1], arg[0]); m.IsErrNotFound() {
				m.SetResult("")
			}
			m.Set(ice.MSG_STATUS)
		}},
	}, Configs: map[string]*ice.Config{
		INNER: {Name: "inner", Help: "源代码", Value: kit.Data(
			EXEC, kit.Dict("py", []string{"python"}),
			PLUG, kit.Dict(
				"S", kit.Dict(PREFIX, kit.Dict("//", COMMENT)),
				"s", kit.Dict(PREFIX, kit.Dict("//", COMMENT), KEYWORD, kit.Dict("TEXT", KEYWORD, "RET", KEYWORD)),
				"py", kit.Dict(PREFIX, kit.Dict("#", COMMENT), KEYWORD, kit.Dict("print", KEYWORD)),
				nfs.HTML, kit.Dict(SPLIT, kit.Dict("space", " ", "operator", "<>"), KEYWORD, kit.Dict("head", KEYWORD, "body", KEYWORD)),
				nfs.CSS, kit.Dict(SUFFIX, kit.Dict("{", COMMENT)),
				"yaml", kit.Dict(PREFIX, kit.Dict("#", COMMENT)),
				"yml", kit.Dict(PREFIX, kit.Dict("#", COMMENT)),

				"makefile", kit.Dict(PREFIX, kit.Dict("#", COMMENT), SUFFIX, kit.Dict(":", COMMENT),
					KEYWORD, kit.Dict("ifeq", KEYWORD, "ifneq", KEYWORD, "else", KEYWORD, "endif", KEYWORD),
				),
			),
		)},
	}})
}
