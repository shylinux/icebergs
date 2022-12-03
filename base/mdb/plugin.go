package mdb

import ice "shylinux.com/x/icebergs"

const PLUGIN = "plugin"

func init() { Index.MergeCommands(ice.Commands{PLUGIN: {Help: "插件", Actions: RenderAction()}}) }
