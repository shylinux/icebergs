package wiki

import (
	ice "github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

func _wiki_list(m *ice.Message, cmd, name string, arg ...string) bool {
	if strings.HasSuffix(name, "/") {
		m.Option(nfs.DIR_ROOT, m.Conf(cmd, "meta.path"))
		m.Option(nfs.DIR_TYPE, nfs.TYPE_DIR)
		m.Cmdy(nfs.DIR, name, "time size path")

		m.Option(nfs.DIR_TYPE, nfs.TYPE_FILE)
		m.Option(nfs.DIR_REG, m.Conf(cmd, "meta.regs"))
		m.Cmdy(nfs.DIR, name, "time size path")
		return true
	}
	return false
}
func _wiki_show(m *ice.Message, cmd, name string, arg ...string) {
	m.Cmdy(nfs.CAT, path.Join(m.Conf(cmd, "meta.path"), name))
}
func _wiki_save(m *ice.Message, cmd, name, text string, arg ...string) {
	m.Cmd(nfs.SAVE, path.Join(m.Conf(cmd, "meta.path"), name), text)
}
func _wiki_upload(m *ice.Message, cmd string) {
	m.Option("request", m.R)
	msg := m.Cmd(web.CACHE, "catch", "", "")
	m.Cmd(web.CACHE, "watch", msg.Append("data"), path.Join(msg.Conf(cmd, "meta.path"),
		path.Dir(m.Option("path")), msg.Append("name")))
}

func reply(m *ice.Message, cmd string, arg ...string) bool {
	// 文件列表
	m.Option("dir_root", m.Conf(cmd, "meta.path"))
	m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
	m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
	m.Sort("time", "time_r")

	if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
		// 目录列表
		m.Option("dir_reg", "")
		m.Option("dir_type", "dir")
		m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
		m.Option("_display", "table")
		return true
	}
	return false
}

var Index = &ice.Context{Name: "wiki", Help: "文档中心",
	Configs: map[string]*ice.Config{
		"walk": {Name: "walk", Help: "走遍世界", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.csv")},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"walk": {Name: "walk path=@province auto", Help: "走遍世界", Meta: kit.Dict("display", "local/wiki/walk"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf(cmd, "meta.path"), arg[2]), arg[3])
				}
				return
			}

			// 文件列表
			m.Option("dir_root", m.Conf(cmd, "meta.path"))
			m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
			m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			m.Sort("time", "time_r")
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
				return
			}
			m.Option("title", "我走过的世界")
			m.CSV(m.Result())
		}},

		"mind": {Name: "mind zone type name text", Help: "思考", List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "action", "auto", "figure", "key",
			kit.MDB_INPUT, "text", "name", "type", "figure", "key",
			kit.MDB_INPUT, "text", "name", "name", "figure", "key",
			kit.MDB_INPUT, "button", "name", "添加",
			kit.MDB_INPUT, "textarea", "name", "text",
			kit.MDB_INPUT, "text", "name", "location", "figure", "key", "cb", "location",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "input":
					// 输入补全
					switch arg[2] {
					case "type":
						m.Push("type", []string{"spark", "order", "table", "label", "chain", "refer", "brief", "chapter", "section", "title"})
					case "path":
						m.Option("_refresh", "true")
						reply(m, "word", arg[3:]...)
					}
					return
				}
			}

			if len(arg) < 2 {
				m.Cmdy("word", arg)
				return
			}
			m.Cmd("word", "action", "追加", arg)

			m.Option("scan_mode", "scan")
			m.Cmdy(ssh.SOURCE, path.Join(m.Conf("word", "meta.path"), arg[0]))
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
