package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
)

const BASH = "bash"

var Index = &ice.Context{Name: BASH, Help: "命令行", Commands: ice.Commands{
	BASH: {Name: "bash path auto order build download", Help: "命令行", Actions: ice.MergeActions(ice.Actions{
		cli.ORDER: {Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.ORDER, m.Config(nfs.SOURCE), "_install/bin")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/bash/5.1_1/bash-5.1.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, m.ConfigSimple(nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
