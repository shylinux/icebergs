package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _wiki_list(m *ice.Message, cmd string, arg ...string) bool {
	m.Option(nfs.DIR_ROOT, m.Conf(cmd, "meta.path"))
	if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
		m.Option("_display", "table")
		if m.Option(nfs.DIR_DEEP) != "true" {
			// 目录列表
			m.Option(nfs.DIR_TYPE, nfs.DIR)
			m.Cmdy(nfs.DIR, kit.Select("./", arg, 0), "time size path")

		}

		// 文件列表
		m.Option(nfs.DIR_TYPE, nfs.FILE)
		m.Option(nfs.DIR_REG, m.Conf(cmd, "meta.regs"))
		m.Cmdy(nfs.DIR, kit.Select("./", arg, 0), "time size path")
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
	if up := kit.Simple(m.Optionv("_upload")); m.Option(ice.MSG_USERPOD) == "" {
		m.Cmdy(web.CACHE, web.WATCH, up[0], path.Join(m.Conf(cmd, "meta.path"), m.Option("path"), up[1]))
	} else {
		m.Cmdy(web.SPIDE, web.SPIDE_DEV, web.SPIDE_SAVE, path.Join(m.Conf(cmd, "meta.path"), m.Option("path"), up[1]), web.SPIDE_GET, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/cache/"+up[0]))
	}
}

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		FEEL, WORD, DATA, DRAW,
		SPARK, IMAGE,
	)
}
