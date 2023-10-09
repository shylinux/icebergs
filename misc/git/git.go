package git

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _git_dir(arg ...string) string                       { return path.Join(path.Join(arg...), ".git") }
func _git_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, GIT, arg) }
func _git_cmds(m *ice.Message, arg ...string) string      { return _git_cmd(m, arg...).Results() }
func _git_tags(m *ice.Message) string                     { return _git_cmds(m, "describe", "--tags") }
func _git_diff(m *ice.Message) string                     { return _git_cmds(m, DIFF, "--shortstat") }
func _git_status(m *ice.Message) string                   { return _git_cmds(m, STATUS, "-sb") }
func _git_remote(m *ice.Message) string                   { return _git_cmds(m, nfs.REMOTE, "get-url", nfs.ORIGIN) }

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库"}

func init() { code.Index.Register(Index, &web.Frame{}) }

func Prefix(arg ...string) string { return code.Prefix(GIT, kit.Keys(arg)) }
