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
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _inner_list(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() {
		kit.If(nfs.IsSourceFile(m, ext), func() { m.Cmdy(nfs.CAT, path.Join(dir, file)) })
		// kit.If(m.IsErrNotFound(), func() { _inner_show(m.RenderResult().SetResult(), ext, file, dir) })
	})
}
func _inner_show(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.RENDER, ext, file, dir) })
}
func _inner_exec(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.ENGINE, ext, file, dir) })
}
func _inner_tags(m *ice.Message, dir string, value string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, cli.GREP, "^"+value+"\\>", nfs.TAGS, kit.Dict(cli.CMD_DIR, dir)), ice.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.SplitN(l, ice.TB, 3)
		if len(ls) < 3 {
			continue
		}
		file, ls := ls[1], strings.SplitN(ls[2], ";\"", 2)
		text := strings.TrimSuffix(strings.TrimPrefix(ls[0], "/^"), "$/")
		if text, line := _inner_line(m, kit.Path(dir, file), text); dir == "" {
			m.PushRecord(kit.Dict(nfs.PATH, path.Dir(file)+ice.PS, nfs.FILE, path.Base(file), nfs.LINE, kit.Format(line), mdb.TEXT, text))
		} else {
			m.PushRecord(kit.Dict(nfs.PATH, dir, nfs.FILE, strings.TrimPrefix(file, nfs.PWD), nfs.LINE, kit.Format(line), mdb.TEXT, text))
		}
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
				case nfs.PATH:
					m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS).ProcessAgain()
					kit.If(strings.HasPrefix(p, bind[0]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[0], bind[1], 1), nfs.DIR_CLI_FIELDS) })
					kit.If(strings.HasPrefix(p, bind[1]), func() { m.Cmdy(nfs.DIR, strings.Replace(p, bind[1], bind[0], 1), nfs.DIR_CLI_FIELDS) })
				case nfs.FILE:
					m.Option(nfs.DIR_ROOT, m.Option(nfs.PATH))
					m.Cmdy(nfs.DIR, kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.PS))+ice.PS, nfs.DIR_CLI_FIELDS).ProcessAgain()
				default:
					m.Cmdy(FAVOR, mdb.INPUTS, arg)
				}
			}},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(mdb.PLUGIN, arg) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _inner_show(m, arg[0], arg[1], arg[2]) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _inner_exec(m, arg[0], arg[1], arg[2]) }},
			nfs.GREP: {Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.GREP, arg[0], kit.Select(m.Option(nfs.PATH), arg, 1))
			}},
			nfs.TAGS: {Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				if _inner_tags(m, m.Option(nfs.PATH), arg[0]); m.Length() == 0 {
					_inner_tags(m, "", arg[0])
				}
			}}, FAVOR: {},
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
				ctx.DisplayLocal(m, "").Option(nfs.REPOS, kit.Join(m.Cmd("web.code.git.repos", ice.OptionFields(nfs.PATH)).Sort(nfs.PATH).Appendv(nfs.PATH)))
			}
		}},
	})
	ctx.AddRunChecker(func(m *ice.Message, cmd, check string, arg ...string) bool {
		process := func(m *ice.Message, file string) bool {
			ls, n := kit.Split(file, ice.PS), kit.Int(kit.Select("2", "1", strings.HasPrefix(file, ice.SRC+ice.PS)))
			ctx.ProcessFloat(m, web.CODE_INNER, kit.Join(kit.Slice(ls, 0, n), ice.PS)+ice.PS, kit.Join(kit.Slice(ls, n), ice.PS))
			return true
		}
		switch check {
		case nfs.SCRIPT:
			if file := kit.ExtChange(ctx.GetCmdFile(m, cmd), nfs.JS); nfs.ExistsFile(m, file) {
				return process(m, file)
			} else if strings.HasPrefix(file, bind[0]) {
				if file := strings.Replace(file, bind[0], bind[1], 1); nfs.ExistsFile(m, file) {
					return process(m, file)
				}
			}
		case nfs.SOURCE:
			if file := ctx.GetCmdFile(m, cmd); nfs.ExistsFile(m, file) {
				return process(m, file)
			}
		}
		return false
	})
}
func InnerPath(arg ...string) (dir, file string) {
	p := strings.TrimPrefix(path.Join(arg...), kit.Path("")+ice.PS)
	if list := strings.Split(p, ice.PS); strings.HasPrefix(p, "usr/") {
		return path.Join(list[:2]...) + ice.PS, path.Join(list[2:]...)
	} else if strings.HasPrefix(p, ".ish/pluged/") {
		return path.Join(list[:5]...) + ice.PS, path.Join(list[5:]...)
	} else {
		return list[0] + ice.PS, path.Join(list[1:]...)
	}
}
func PlugAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			kit.Fetch([]string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE}, func(cmd string) { m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey()) })
			LoadPlug(m, m.CommandKey())
		}},
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(PLUG)) }},
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
		m.Confm(lang, kit.Keym(PLUG, PREPARE), func(k string, v ice.Any) {
			kit.Fetch(kit.Simple(v), func(v string) { m.Conf(lang, kit.Keym(PLUG, KEYWORD, v), k) })
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
