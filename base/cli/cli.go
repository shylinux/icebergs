package cli

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const CLI = "cli"

var Index = &ice.Context{Name: CLI, Help: "命令模块"}

func Prefix(arg ...string) string { return kit.Keys(CLI, arg) }

func init() { ice.Index.Register(Index, nil, RUNTIME, SYSTEM, DAEMON, FOREVER, MIRRORS, QRCODE) }
