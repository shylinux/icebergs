package tcp

import (
	ice "shylinux.com/x/icebergs"
)

const TCP = "tcp"

var Index = &ice.Context{Name: TCP, Help: "通信模块"}

func init() { ice.Index.Register(Index, nil, HOST, PORT, CLIENT, SERVER) }
