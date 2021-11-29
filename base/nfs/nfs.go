package nfs

import (
	ice "shylinux.com/x/icebergs"
)

var Index = &ice.Context{Name: "nfs", Help: "存储模块"}

func init() { ice.Index.Register(Index, nil, TAR, CAT, DIR, TAIL, TRASH, SAVE, PUSH, COPY, LINK, DEFS) }
