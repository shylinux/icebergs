package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _wiki_path(m *ice.Message, cmd string, arg ...string) string {
	return path.Join(m.Option(ice.MSG_LOCAL), m.Conf(cmd, kit.META_PATH), path.Join(arg...))
}

func _wiki_list(m *ice.Message, cmd string, arg ...string) bool {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
		if m.Option(nfs.DIR_DEEP) != "true" {
			// 目录列表
			m.Option(nfs.DIR_TYPE, nfs.DIR)
			m.Cmdy(nfs.DIR, kit.Select("./", arg, 0), "time,size,path")
		}

		// 文件列表
		m.Option(nfs.DIR_TYPE, nfs.CAT)
		m.Cmdy(nfs.DIR, kit.Select("./", arg, 0), "time,size,path")
		return true
	}
	return false
}
func _wiki_show(m *ice.Message, cmd, name string, arg ...string) {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	m.Cmdy(nfs.CAT, name)
}
func _wiki_save(m *ice.Message, cmd, name, text string, arg ...string) {
	m.Option(nfs.DIR_ROOT, _wiki_path(m, cmd))
	m.Cmd(nfs.SAVE, name, text)
}
func _wiki_upload(m *ice.Message, cmd string, dir string) {
	up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
	if p := _wiki_path(m, cmd, dir, up[1]); m.Option(ice.MSG_USERPOD) == "" {
		// 本机文件
		m.Cmdy(web.CACHE, web.WATCH, up[0], p)
	} else {
		// 下发文件
		m.Cmdy(web.SPIDE, web.SPIDE_DEV, web.SPIDE_SAVE, p, web.SPIDE_GET,
			kit.MergeURL2(m.Option(ice.MSG_USERWEB), path.Join("/share/cache", up[0])))
	}
}

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.RENDER, mdb.CREATE, PNG, m.Prefix(IMAGE))
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		FEEL, DRAW, DATA, WORD,
		SPARK, IMAGE,
	)
}
