package git

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _spide_for(text string, cb func([]string)) {
	for _, line := range strings.Split(text, ice.NL) {
		if len(line) == 0 || strings.HasPrefix(line, "!_") {
			continue
		}

		cb(kit.Split(line, "\t ", "\t ", "\t "))
	}
}
func _spide_go(m *ice.Message, file string) {
	_spide_for(m.Cmdx(cli.SYSTEM, "gotags", file), func(ls []string) {
		switch ls[3] {
		case "i":
			return
		case "w", "e":
			return
			ls[0] = "-" + ls[0] + ":" + strings.TrimPrefix(ls[len(ls)-1], "type:")
		case "m":
			if strings.HasPrefix(ls[6], "ntype") {
				return
			} else if strings.HasPrefix(ls[5], "ctype") {
				ls[0] = strings.TrimPrefix(ls[5], "ctype:") + ":" + ls[0]
			} else {
				ls[0] = ls[3] + ":" + ls[0]
			}
		default:
			ls[0] = ls[3] + ":" + ls[0]
		}

		m.Push(kit.MDB_NAME, ls[0])
		m.Push(kit.MDB_FILE, ls[1])
		m.Push(kit.MDB_LINE, strings.TrimSuffix(ls[2], ";\""))
		m.Push(kit.MDB_TYPE, ls[3])
		m.Push(kit.MDB_EXTRA, strings.Join(ls[4:], ice.SP))
	})
}
func _spide_c(m *ice.Message, file string) {
	_spide_for(m.Cmdx(cli.SYSTEM, "ctags", "-f", "-", file), func(ls []string) {
		m.Push(kit.MDB_NAME, ls[0])
		m.Push(kit.MDB_FILE, ls[1])
		m.Push(kit.MDB_LINE, "1")
	})
}

const SPIDE = "spide"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		SPIDE: {Name: "spide name auto", Help: "构架图", Meta: kit.Dict(
			ice.Display("/plugin/story/spide.js"),
		), Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(REPOS, ice.OptionFields("name,time"))
			}}, code.INNER: {Name: "web.code.inner"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 { // 仓库列表
				m.Cmdy(REPOS)
				return
			}

			if arg[0] == path.Base(kit.Pwd()) {
				m.Option(nfs.DIR_ROOT, path.Join(ice.SRC)+ice.PS)
			} else {
				m.Option(nfs.DIR_ROOT, path.Join(ice.USR, arg[0])+ice.PS)
			}

			if len(arg) == 1 { // 目录列表
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				color := []string{cli.YELLOW, cli.BLUE, cli.CYAN, cli.RED}
				nfs.Dir(m, kit.MDB_PATH).Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_COLOR, color[strings.Count(value[kit.MDB_PATH], ice.PS)%len(color)])
				})
				return
			}
			if !strings.HasSuffix(arg[1], arg[2]) {
				return
			}

			// 语法解析
			switch m.Option(cli.CMD_DIR, m.Option(nfs.DIR_ROOT)); kit.Ext(arg[1]) {
			case code.GO:
				_spide_go(m, arg[1])
			default:
				_spide_c(m, arg[1])
			}
			m.SortInt(kit.MDB_LINE)
		}},
	}})
}
