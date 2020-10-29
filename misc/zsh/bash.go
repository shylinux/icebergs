package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"path"
)

const BASH = "bash"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BASH: {Name: BASH, Help: "命令行", Value: kit.Data(
				"source", "http://mirrors.aliyun.com/gnu/bash/bash-4.2.53.tar.gz",
			)},
		},
		Commands: map[string]*ice.Command{
			BASH: {Name: "bash port path auto start build download", Help: "命令行", Action: map[string]*ice.Action{
				web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(BASH, kit.META_SOURCE))
				}},
				gdb.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.INSTALL, gdb.BUILD, m.Conf(BASH, kit.META_SOURCE))
				}},
				gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.INSTALL, gdb.START, m.Conf(BASH, kit.META_SOURCE), "bin/bash")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(code.INSTALL, path.Base(m.Conf(BASH, kit.META_SOURCE)), arg)
			}},
		},
	})
}
