package mdb

import (
	ice "shylinux.com/x/icebergs"
)

const ENGINE = "engine"

func init() {
	Index.MergeCommands(ice.Commands{ENGINE: {Name: "engine type name text auto", Help: "引擎", Actions: RenderAction()}})
}
