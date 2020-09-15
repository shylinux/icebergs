package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"path"
)

func _vimer_save(m *ice.Message, ext, file, dir string, text string) {
	if f, p, e := kit.Create(path.Join(dir, file)); e == nil {
		defer f.Close()
		if n, e := f.WriteString(text); m.Assert(e) {
			m.Log_EXPORT("file", path.Join(dir, file), "size", n)
		}
		m.Echo(p)
	}
}

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			VIMER: {Name: "vimer path=usr/demo file=hi.sh line=1 刷新:button=auto 保存:button 运行:button 项目:button", Help: "编辑器", Meta: kit.Dict(
				"display", "/plugin/local/code/vimer.js", "style", "editor",
			), Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload path name", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.CACHE, web.UPLOAD)
					m.Cmdy(web.CACHE, web.WATCH, m.Option(web.DATA), path.Join(m.Option("path"), m.Option("name")))
				}},
				nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_vimer_save(m, arg[0], arg[1], arg[2], m.Option("content"))
				}},

				"cmd": {Name: "cmd type file path", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(INNER, arg)
			}},
		},
	}, nil)
}
