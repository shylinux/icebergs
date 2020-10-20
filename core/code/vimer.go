package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
)

func _vimer_path(m *ice.Message, arg ...string) string {
	return path.Join(m.Option(ice.MSG_LOCAL), path.Join(arg...))
}
func _vimer_upload(m *ice.Message, dir string) {
	up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
	if p := _vimer_path(m, dir, up[1]); m.Option(ice.MSG_USERPOD) == "" {
		m.Cmdy(web.CACHE, web.WATCH, up[0], p)
	} else {
		m.Cmdy(web.SPIDE, web.SPIDE_DEV, web.SPIDE_SAVE, p, web.SPIDE_GET, kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/share/cache/"+up[0]))
	}
}

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			VIMER: {Name: "vimer path=usr/demo file=hi.sh line=1 刷新:button=auto save project search", Help: "编辑器", Meta: kit.Dict(
				"display", "/plugin/local/code/vimer.js", "style", "editor",
				"trans", kit.Dict("project", "项目", "search", "搜索"),
			), Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					_vimer_upload(m, m.Option(kit.MDB_PATH))
				}},
				nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.SAVE, path.Join(m.Option(kit.MDB_PATH), m.Option(kit.MDB_FILE)))
				}},
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if !m.Warn(!m.Right(arg)) {
						m.Cmdy(arg)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(INNER, arg)
			}},
		},
	}, nil)
}
