package cli

import ice "shylinux.com/x/icebergs"

const CLI = "cli"

var Index = &ice.Context{Name: CLI, Help: "命令模块"}

func init() { ice.Index.Register(Index, nil, MIRRORS, RUNTIME, QRCODE, SYSTEM, DAEMON, FOREVER) }
