package team

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const TEAM = "team"

var Index = &ice.Context{Name: TEAM, Help: "团队中心"}

func init() { web.Index.Register(Index, nil, TODO, EPIC, TASK, PLAN) }
