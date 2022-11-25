package mdb

import (
	ice "shylinux.com/x/icebergs"
)

const ENGINE = "engine"

func init() { Index.MergeCommands(ice.Commands{ENGINE: {Help: "引擎", Actions: RenderAction()}}) }
