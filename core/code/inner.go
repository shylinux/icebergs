package code

import (
	"bufio"
	"net/http"
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
	file = kit.Split(file, "?")[0]
	kit.If(aaa.Right(m, dir, file), func() {
		kit.If(nfs.IsSourceFile(m, ext), func() {
			m.Cmdy(nfs.CAT, path.Join(dir, file))
		}, func() {
			_inner_show(m.RenderResult().SetResult(), ext, file, dir)
			kit.If(m.Result() == "", func() { m.Cmdy(nfs.CAT, path.Join(dir, file)) })
		})
	})
}
func _inner_show(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.RENDER, ext, file, dir) })
}
func _inner_exec(m *ice.Message, ext, file, dir string) {
	kit.If(aaa.Right(m, dir, file), func() { m.Cmdy(mdb.ENGINE, ext, file, dir) })
}
func _inner_tags(m *ice.Message, dir string, value string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, cli.GREP, "^"+value+"\\>", nfs.TAGS, dir), lex.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.SplitN(l, lex.TB, 3)
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
	INCLUDE  = "include"
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
	Index.MergeCommands(ice.Commands{
		INNER: {Name: "inner path=src/ file=main.go line=1 auto", Help: "源代码", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(FAVOR, mdb.INPUTS, arg) }},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.PLUGIN, arg); m.Result() == "" {
					m.Cmdy(mdb.PLUGIN, m.Option(lex.PARSE, strings.ToLower(kit.Split(path.Base(arg[1]), nfs.PT)[0])), arg[1:])
				}
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _inner_show(m, arg[0], arg[1], arg[2]) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _inner_exec(m, arg[0], arg[1], arg[2]) }},
			nfs.TAGS:   {Hand: func(m *ice.Message, arg ...string) { _inner_tags(m, m.Option(nfs.PATH), arg[0]) }},
			nfs.GREP:   {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.GREP, arg) }},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(NAVIGATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefix(arg[0], "/volcanos/", "/require/", ice.HTTP) {
				if kit.HasPrefix(arg[0], nfs.REQUIRE_SRC) {
					m.Option(nfs.FILE, strings.Split(strings.TrimPrefix(arg[0], nfs.REQUIRE_SRC), "?")[0])
					m.Option(nfs.PATH, nfs.SRC)
				}
				if kit.HasPrefix(arg[0], nfs.REQUIRE_USR) {
					ls := kit.Split(arg[0], nfs.PS)
					m.Option(nfs.FILE, strings.Split(strings.TrimPrefix(arg[0], nfs.REQUIRE_USR+ls[2]+nfs.PS), "?")[0])
					m.Option(nfs.PATH, nfs.USR+ls[2]+nfs.PS)
				}
				m.Echo(m.Cmdx(web.SPIDE, ice.OPS, web.SPIDE_RAW, http.MethodGet, arg[0]))
				m.Options("mode", "simple", lex.PARSE, kit.Ext(kit.ParseURL(arg[0]).Path))
				ctx.DisplayLocal(m, "")
			} else if arg[0] = strings.Split(arg[0], mdb.FS)[0]; !strings.HasSuffix(arg[0], nfs.PS) && len(arg) == 1 {
				arg[1] = kit.Slice(strings.Split(arg[0], nfs.PS), -1)[0]
				arg[0] = strings.TrimSuffix(arg[0], arg[1])
				m.ProcessRewrite(nfs.PATH, arg[0], nfs.FILE, arg[1])
			} else if len(arg) < 2 {
				nfs.Dir(m, nfs.PATH)
			} else if strings.Contains(arg[1], nfs.DF) {
				ls := strings.Split(arg[1], nfs.DF)
				m.ProcessRewrite(nfs.LINE, ls[0], nfs.FILE, ls[1])
			} else {
				arg[1] = strings.Split(arg[1], mdb.FS)[0]
				_inner_list(m, kit.Ext(arg[1]), arg[1], arg[0])
				if ctx.DisplayLocal(m, ""); !strings.HasPrefix(arg[0], ice.USR_INSTALL) {
					m.Option(REPOS, kit.Join(m.Cmd(REPOS, ice.OptionFields(nfs.PATH)).Sort(nfs.PATH).Appendv(nfs.PATH)))
				}
				m.StatusTime(mdb.TIME, ice.Info.Make.Time, nfs.FILE, arg[1], nfs.LINE, kit.Select("1", arg, 2))
			}
		}},
	})
}
func PlugAction(arg ...ice.Any) ice.Actions {
	return ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if cmd := m.Target().Commands[m.CommandKey()]; cmd != nil {
				if cmd.Name == "" {
					cmd.Name = kit.Format("%s path auto", m.CommandKey())
					cmd.List = ice.SplitCmd(cmd.Name, cmd.Actions)
				}
			}
			kit.For([]string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE}, func(cmd string) {
				m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			})
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
	}, ctx.ConfAction(arg...))
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
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, kit.Default(cmds, CTAGS, "--excmd=number", "--sort=no", "-f", "-", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))), lex.NL) {
		if strings.HasPrefix(l, "!_") {
			continue
		}
		ls := strings.Split(l, lex.TB)
		if len(ls) < 3 {
			continue
		}
		switch ls[3] {
		case "w", "m":
			continue
		}
		m.PushRecord(kit.Dict(mdb.TYPE, ls[3], mdb.NAME, ls[0], nfs.LINE, strings.TrimSuffix(ls[2], ";\"")))
	}
	m.Sort(nfs.LINE)
}
