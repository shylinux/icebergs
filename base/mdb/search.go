package mdb

import (
	ice "shylinux.com/x/icebergs"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{SEARCH: {Name: "search type name text auto", Help: "搜索", Actions: RenderAction()}})
}
