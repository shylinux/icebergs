package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _wiki_list(m *ice.Message, cmd, name string, arg ...string) bool {
	m.Debug(name)
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
	m.Cmdy(web.CACHE, web.UPLOAD)
	m.Cmdy(web.CACHE, web.WATCH, m.Option(web.DATA), path.Join(m.Conf(cmd, "meta.path"), m.Option("path"), m.Option("name")))
}

func reply(m *ice.Message, cmd string, arg ...string) bool {
	// 文件列表
	m.Option(nfs.DIR_ROOT, m.Conf(cmd, "meta.path"))
	if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
		m.Option("_display", "table")
		// if m.Option(nfs.DIR_DEEP) == "true" {
		// 	return true
		// }

		// 目录列表
		m.Option(nfs.DIR_TYPE, nfs.DIR)
		m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))

		// 文件列表
		m.Option(nfs.DIR_TYPE, nfs.FILE)
		m.Option(nfs.DIR_REG, m.Conf(cmd, "meta.regs"))
		m.Cmdy(nfs.DIR, kit.Select("./", arg, 0))
		return true
	}
	return false
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			"walk": {Name: "walk", Help: "走遍世界", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.csv")},
		},
		Commands: map[string]*ice.Command{
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
	}, nil)
}
