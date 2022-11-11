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
	if aaa.Right(m, dir, file) {
		if nfs.IsSourceFile(m, ext) {
			m.Cmdy(nfs.CAT, path.Join(dir, file))
		}
		if m.IsErrNotFound() {
			_inner_show(m.SetResult(), ext, file, dir)
		}
	}
}
func _inner_show(m *ice.Message, ext, file, dir string) {
	if aaa.Right(m, dir, file) {
		m.Cmdy(mdb.RENDER, ext, file, dir)
	}
}
func _inner_exec(m *ice.Message, ext, file, dir string) {
	if aaa.Right(m, dir, file) {
		m.Cmdy(mdb.ENGINE, ext, file, dir)
	}
}
func _inner_tags(m *ice.Message, dir string, value string) {
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, nfs.GREP, "^"+value+"\\>", nfs.TAGS, kit.Dict(cli.CMD_DIR, dir)), ice.NL) {
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
	Index.MergeCommands(ice.Commands{
		INNER: {Name: "inner path=src/@key file=main.go@key line=1 auto", Help: "源代码", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.PATH:
					p := kit.Select(nfs.PWD, arg, 1)
					m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS).Sort(nfs.PATH).ProcessAgain()
					if strings.HasPrefix(p, "usr/icebergs/core/") && len(kit.Split(p, ice.PS)) > 3 {
						p = strings.Replace(p, "usr/icebergs/core/", "usr/volcanos/plugin/local/", 1)
						m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS).Sort(nfs.PATH)
					} else if strings.HasPrefix(p, "usr/volcanos/plugin/local/") && len(kit.Split(p, ice.PS)) > 4 {
						p = strings.Replace(p, "usr/volcanos/plugin/local/", "usr/icebergs/core/", 1)
						m.Cmdy(nfs.DIR, p, nfs.DIR_CLI_FIELDS).SortStrR(nfs.PATH)
					}
				case nfs.FILE:
					p := kit.Select(nfs.PWD, arg, 1)
					m.Option(nfs.DIR_ROOT, m.Option(nfs.PATH))
					m.Cmdy(nfs.DIR, kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.PS))+ice.PS, nfs.DIR_CLI_FIELDS).ProcessAgain()
				default:
					m.Cmdy(FAVOR, mdb.INPUTS, arg)
				}
			}},
			mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(mdb.PLUGIN, arg) }},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _inner_show(m, arg[0], arg[1], arg[2]) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _inner_exec(m, arg[0], arg[1], arg[2]) }},

			nfs.GREP: {Name: "grep", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.GREP, arg[0], m.Option(nfs.PATH)).StatusTimeCount(mdb.INDEX, 0)
			}},
			nfs.TAGS: {Name: "tags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				if _inner_tags(m, m.Option(nfs.PATH), arg[0]); m.Length() == 0 {
					_inner_tags(m, "", arg[0])
				}
			}},
			NAVIGATE: {Name: "navigate", Help: "跳转", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(NAVIGATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			FAVOR: {Name: "favor", Help: "收藏"},
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				if !ctx.PodCmd(m, ctx.COMMAND, arg) {
					m.Cmdy(ctx.COMMAND, arg)
				}
				if len(arg) == 2 && arg[0] == mdb.SEARCH && arg[1] == ctx.COMMAND {
					return
				}
				m.Cmd(FAVOR, mdb.INSERT, mdb.ZONE, "_recent_cmd", nfs.FILE, arg[0])
			}},
		}, ctx.CmdAction(), aaa.RoleAction(ice.SRC_MAIN_GO)), Hand: func(m *ice.Message, arg ...string) {
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

			arg[1] = strings.Split(arg[1], ice.FS)[0]
			_inner_list(m, kit.Ext(arg[1]), arg[1], arg[0])
			defer m.Cmd(FAVOR, mdb.INSERT, mdb.ZONE, "_recent_file", nfs.PATH, arg[0], nfs.FILE, arg[1])
			m.Option("tabs", m.Config("show.tabs"))
			m.Option("plug", m.Config("show.plug"))
			m.Option("exts", m.Config("show.exts"))
			ctx.DisplayLocal(m, "")
		}},
	})
}
func PlugAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			for _, cmd := range []string{mdb.PLUGIN, mdb.RENDER, mdb.ENGINE} {
				m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey())
			}
			LoadPlug(m, m.CommandKey())
		}},
		mdb.PLUGIN: {Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Config(PLUG)) }},
		mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.CAT, path.Join(arg[2], arg[1])) }},
		mdb.SELECT: {Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == m.CommandKey() {
				m.Cmdy("", mdb.ENGINE, m.CommandKey(), arg[0], ice.SRC)
				return
			}
			m.Option(nfs.DIR_ROOT, ice.SRC)
			m.Option(nfs.DIR_DEEP, ice.TRUE)
			m.Option(nfs.DIR_REG, kit.Format(`.*\.(%s)$`, m.CommandKey()))
			m.Cmdy(nfs.DIR, arg)
		}},
	}
}
func LoadPlug(m *ice.Message, lang ...string) {
	for _, lang := range lang {
		m.Conf(nfs.CAT, kit.Keym(nfs.SOURCE, kit.Ext(lang)), ice.TRUE)
		m.Confm(lang, kit.Keym(PLUG, PREPARE), func(key string, value ice.Any) {
			for _, v := range kit.Simple(value) {
				m.Conf(lang, kit.Keym(PLUG, KEYWORD, v), key)
			}
		})
	}
}
func TagsList(m *ice.Message, cmds ...string) {
	if len(cmds) == 0 {
		cmds = []string{"ctags", "--excmd=number", "--sort=no", "-f", "-", path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE))}
	}
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, cmds), ice.NL) {
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
