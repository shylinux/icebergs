package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _git_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, GIT, arg) }
func _git_cmds(m *ice.Message, arg ...string) string      { return _git_cmd(m, arg...).Results() }

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库"}

func init() { code.Index.Register(Index, &web.Frame{}, STATUS, REPOS) }

func Prefix(arg ...string) string { return code.Prefix(GIT, kit.Keys(arg)) }
