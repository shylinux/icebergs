package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"path"
)

const ZSH = "zsh"

var Index = &ice.Context{Name: ZSH, Help: "命令行",
	Configs: map[string]*ice.Config{
		ZSH: {Name: ZSH, Help: "命令行", Value: kit.Data(
			"source", "https://nchc.dl.sourceforge.net/project/zsh/zsh/5.8/zsh-5.8.tar.xz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		ZSH: {Name: "zsh port path auto start build download", Help: "命令行", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(ZSH, kit.META_SOURCE))
			}},
			gdb.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, gdb.BUILD, m.Conf(ZSH, kit.META_SOURCE))
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, gdb.START, m.Conf(ZSH, kit.META_SOURCE), "bin/zsh")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(ZSH, kit.META_SOURCE)), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
