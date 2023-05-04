package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{SEARCH: {Help: "搜索", Actions: RenderAction()}})
	ice.AddMergeAction(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) ice.Handler {
		if sub == SEARCH {
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, CREATE, m.CommandKey(), m.PrefixKey()) }
		}
		return nil
	})
}
func IsSearchForEach(m *ice.Message, arg []string, cb func() []string) bool {
	if arg[0] == FOREACH && arg[1] == "" {
		if cb != nil {
			args := cb()
			m.PushSearch(TYPE, kit.Select("", args, 0), NAME, kit.Select("", args, 1), TEXT, kit.Select("", args, 2))
		}
		return true
	}
	return false
}
