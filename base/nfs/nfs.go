package nfs

import ice "shylinux.com/x/icebergs"

const NFS = "nfs"

var Index = &ice.Context{Name: NFS, Help: "存储模块"}

func init() {
	ice.Index.Register(Index, nil, TAR, CAT, DIR, PACK, DEFS, SAVE, PUSH, COPY, LINK, GREP, TRASH)
}
