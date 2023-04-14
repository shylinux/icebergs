package git

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _git_url(m *ice.Message, repos string) string {
	return kit.MergeURL2(web.UserHost(m), "/x/"+path.Join(repos)+".git")
}
func _git_dir(arg ...string) string                       { return path.Join(path.Join(arg...), ".git") }
func _git_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, GIT, arg) }
func _git_cmds(m *ice.Message, arg ...string) string      { return _git_cmd(m, arg...).Results() }
func _git_tags(m *ice.Message) string                     { return _git_cmds(m, "describe", "--tags") }
func _git_diff(m *ice.Message) string                     { return _git_cmds(m, DIFF, "--shortstat") }
func _git_status(m *ice.Message) string                   { return _git_cmds(m, STATUS, "-sb") }
func _git_remote(m *ice.Message) string {
	return _git_cmds(m, nfs.REMOTE, "get-url", nfs.ORIGIN)
}

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库", Commands: ice.Commands{
	GIT: {Name: "git path auto order build download", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
		cli.ORDER: {Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(code.INSTALL, cli.ORDER, mdb.Config(m, nfs.SOURCE), "_install/libexec/git-core")
			m.Cmdy(code.INSTALL, cli.ORDER, mdb.Config(m, nfs.SOURCE), "_install/bin")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/git-cinnabar/git-2.31.1.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, ctx.ConfigSimple(m, nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}, REPOS) }
