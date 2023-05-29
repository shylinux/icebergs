package nfs

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const NFS = "nfs"

var Index = &ice.Context{Name: NFS, Help: "存储模块"}

func init() {
	ice.Index.Register(Index, nil, ZIP, TAR, CAT, DIR, PACK, DEFS, SAVE, PUSH, COPY, LINK, GREP, TRASH)
}
func Prefix(arg ...string) string { return kit.Keys(NFS, arg) }
