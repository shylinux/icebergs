package nfs

import ice "shylinux.com/x/icebergs"

const NFS = "nfs"

var Index = &ice.Context{Name: "nfs", Help: "存储模块"}

func init() { ice.Index.Register(Index, nil, CAT, DIR, PACK, DEFS, SAVE, PUSH, COPY, LINK, GREP, TAIL, TRASH) }
