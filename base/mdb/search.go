package mdb

import ice "shylinux.com/x/icebergs"

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{SEARCH: {Help: "搜索", Actions: RenderAction()}})
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		if sub == SEARCH {
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, CREATE, m.CommandKey(), m.PrefixKey()) }, nil
		}
		return nil, nil
	})
}
