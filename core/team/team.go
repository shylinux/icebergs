package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TEAM = "team"

var Index = &ice.Context{Name: TEAM, Help: "团队中心"}

func init() { web.Index.Register(Index, nil, TODO, EPIC, TASK, PLAN) }

func Prefix(arg ...string) string { return web.Prefix(TEAM, kit.Keys(arg)) }
