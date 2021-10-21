package yac

import (
	ice "shylinux.com/x/icebergs"
)

const YAC = "yac"

var Index = &ice.Context{Name: YAC, Help: "语法模块"}

func init() { ice.Index.Register(Index, nil) }
