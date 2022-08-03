package nfs

import ice "shylinux.com/x/icebergs"

var Index = &ice.Context{Name: "nfs", Help: "存储模块"}

func init() {
	ice.Index.Register(Index, nil, TAR, CAT, DIR, PACK, DEFS, SAVE, PUSH, COPY, LINK, TAIL, TRASH, GREP)
}
