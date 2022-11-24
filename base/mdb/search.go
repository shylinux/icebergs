package mdb

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{SEARCH: {Help: "搜索", Actions: RenderAction()}})
	ice.AddMerges(func(c *ice.Context, key string, cmd *ice.Command, sub string, action *ice.Action) (ice.Handler, ice.Handler) {
		switch sub {
		case SEARCH:
			return func(m *ice.Message, arg ...string) { m.Cmd(sub, CREATE, m.CommandKey(), m.PrefixKey()) }, nil
		}
		return nil, nil
	})
}
func SearchAction() ice.Actions { return ice.Actions{SEARCH: {Hand: func(m *ice.Message, arg ...string) { HashSelectSearch(m, arg) }}} }
func HashSearchAction(arg ...Any) ice.Actions { return ice.MergeActions(HashAction(arg...), SearchAction()) }
func HashSelectSearch(m *ice.Message, args []string, keys ...string) *ice.Message {
	if args[0] != m.CommandKey() {
		return m
	}
	if len(keys) == 0 {
		keys = kit.Filters(kit.Split(m.Config(FIELD)), TIME, HASH)
	}
	HashSelectValue(m, func(value ice.Map) {
		if args[1] == "" || args[1] == value[keys[1]] {
			m.PushSearch(kit.SimpleKV("", value[keys[0]], value[keys[1]], value[keys[2]]), value)
		}
	})
	return m
}
