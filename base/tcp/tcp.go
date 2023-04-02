package tcp

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const TCP = "tcp"

var Index = &ice.Context{Name: TCP, Help: "通信模块"}

func init() { ice.Index.Register(Index, nil, HOST, PORT, CLIENT, SERVER) }

func Prefix(arg ...string) string { return kit.Keys(TCP, kit.Keys(arg)) }
