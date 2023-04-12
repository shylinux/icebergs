package code

import (
	"bufio"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _inner_list(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() {
		kit.If(nfs.IsSourceFile(m, ext), func() { m.Cmdy(nfs.CAT, path.Join(dir, file)) }, func() { _inner_show(m.RenderResult().SetResult(), ext, file, dir) })
	})
}
func _inner_show(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.RENDER, ext, file, dir) })
}
func _inner_exec(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.ENGINE, ext, file, dir) })
}
func _inner_tags(m *ice.Message, dir string, value string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, cli.GREP, "^"+value+"\\>", nfs.TAGS, dir), ice.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.SplitN(l, ice.TB, 3)
		if len(ls) < 3 {
			continue
		}
		file, ls := ls[1], strings.SplitN(ls[2], ";\"", 2)
		text := strings.TrimSuffix(strings.TrimPrefix(ls[0], "/^"), "$/")
		text, line := _inner_line(m, kit.Path(dir, file), text)
		_ls := nfs.SplitPath(m, path.Join(dir, file))
		m.PushRecord(kit.Dict(nfs.PATH, _ls[0], nfs.FILE, _ls[1], nfs.LINE, kit.Format(line), mdb.TEXT, text))
	}
}
func _inner_line(m *ice.Message, file, text string) (string, int) {
	line := kit.Int(text)
	f, e := nfs.OpenFile(m, file)
	m.Assert(e)
	defer f.Close()
	bio := bufio.NewScanner(f)
	for i := 1; bio.Scan(); i++ {
		if i == line || bio.Text() == text {
			return bio.Text(), i
		}
	}
	return "", 0
}

const (
	SPLIT    = lex.SPLIT
	SPACE    = lex.SPACE
	OPERATOR = lex.OPERATOR
	PREFIX   = lex.PREFIX
	SUFFIX   = lex.SUFFIX
)
const (
	COMMENT  = "comment"
	KEYWORD  = "keyword"
	CONSTANT = "constant"
	DATATYPE = "datatype"
	FUNCTION = "function"
)
const (
	PLUG = "plug"
	SHOW = "show"
	EXEC = "exec"
)
const INNER = "inner"

func init() {
	var bind = []string{"usr/icebergs/core/", "usr/volcanos/plugin/local/"}
	Index.MergeCommands(ice.Commands{
		INNER: {Name: "inner path=src/@key file=main.go@key line=1 auto", Help: "源代码", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch p := kit.Select(nfs.PWD, arg, 1); arg[0] {
				case ice.CMD:
					m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX), func(value ice.Maps) {
						if strings.HasPrefix(value[ctx.INDEX], kit.Select("", arg, 1)) {
							ls := kit.Split(strings.TrimPrefix(value[ctx.INDEX], kit.Select("", arg, 1)), ice.PT)
							m.Push(arg[0], ls[0]+kit.Select("", ice.PT, len(ls) > 1))
						}
					})
				case ctx.INDEX:
					m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX))
				case ctx.ARGS:
					kit.If(m.Option(ctx.INDEX) != "", func() {
						m.Cmdy(m.Option(ctx.INDEX)).Search(m.Option(ctx.INDEX), func(key string, cmd *ice.Command) { m.Cut(kit.Format(kit.Value(cmd.List, "0.name"))) })
					})
				case nfs.PATH:
					m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS)
					kit.If(strings.HasPrefix(p, bind[0]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[0], bind[1], 1), nfs.DIR_CLI_FIELDS) })
					kit.If(strings.HasPrefix(p, bind[1]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[1], bind[0], 1), nfs.DIR_CLI_FIELDS) })
				case nfs.FILE:
					m.Option(nfs.DIR_DEEP, ice.TRUE)
					m.Cmdy(nfs.DIR, path.Join(m.Option(nfs.PATH), kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.PS))+ice.PS), nfs.PATH)
				default:
					m.Cmdy(FAVOR, mdb.INPUTS, arg)
				}
			}},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(mdb.PLUGIN, arg) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _inner_show(m, arg[0], arg[1], arg[2]) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _inner_exec(m, arg[0], arg[1], arg[2]) }},
			nfs.GREP:   {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.GREP, arg) }},
			nfs.TAGS:   {Hand: func(m *ice.Message, arg ...string) { _inner_tags(m, m.Option(nfs.PATH), arg[0]) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(NAVIGATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
		}, ctx.CmdAction(), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if arg[0] = strings.Split(arg[0], ice.FS)[0]; !strings.HasSuffix(arg[0], ice.PS) && len(arg) == 1 {
				arg[1] = kit.Slice(strings.Split(arg[0], ice.PS), -1)[0]
				arg[0] = strings.TrimSuffix(arg[0], arg[1])
				ctx.ProcessRewrite(m, nfs.PATH, arg[0], nfs.FILE, arg[1])
			} else if len(arg) < 2 {
				nfs.Dir(m, nfs.PATH)
			} else {
				arg[1] = strings.Split(arg[1], ice.FS)[0]
				_inner_list(m, kit.Ext(arg[1]), arg[1], arg[0])
				ctx.DisplayLocal(m, "").Option(REPOS, kit.Join(m.Cmd(REPOS, ice.OptionFields(nfs.PATH)).Sort(nfs.PATH).Appendv(nfs.PATH)))
			}
		}},
	})
}
func PlugAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			kit.For([]string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, TEMPLATE, COMPLETE}, func(cmd string) { m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey()) })
			LoadPlug(m, m.CommandKey())
		}},
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(mdb.Config(m, PLUG)) }},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.SELECT: {Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy("", mdb.ENGINE, m.CommandKey(), arg[0], ice.SRC)
			} else {
				m.Cmdy(nfs.DIR, arg, kit.Dict(nfs.DIR_ROOT, ice.SRC, nfs.DIR_DEEP, ice.TRUE, nfs.DIR_REG, kit.ExtReg(m.CommandKey())))
			}
		}},
	}
}
func LoadPlug(m *ice.Message, lang ...string) {
	for _, lang := range lang {
		m.Conf(nfs.CAT, kit.Keym(nfs.SOURCE, kit.Ext(lang)), ice.TRUE)
		mdb.Confm(m, lang, kit.Keym(PLUG, PREPARE), func(k string, v ice.Any) {
			kit.For(kit.Simple(v), func(v string) { m.Conf(lang, kit.Keym(PLUG, KEYWORD, v), k) })
		})
	}
}
func TagsList(m *ice.Message, cmds ...string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, kit.Default(cmds, "ctags", "--excmd=number", "--sort=no", "-f", "-", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))), ice.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.Split(l, ice.TB)
		if len(ls) < 3 {
			continue
		}
		switch ls[3] {
		case "w", "m":
			continue
		}
		m.PushRecord(kit.Dict(mdb.TYPE, ls[3], mdb.NAME, ls[0], nfs.LINE, strings.TrimSuffix(ls[2], ";\"")))
	}
	m.Sort(nfs.LINE).StatusTimeCount()
}
