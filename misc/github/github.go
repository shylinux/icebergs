package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/code"
)

var Index = &ice.Context{Name: "github", Help: "github"}

func init() { code.Index.Merge(Index) }
