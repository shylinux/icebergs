package git

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const TREND = "trend"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			TREND: {Name: "trend name begin_time@date auto", Help: "趋势图", Meta: kit.Dict(
				"display", "/plugin/story/trend.js",
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(ice.MSG_DISPLAY, "table")
				}
				m.Cmdy(TOTAL, arg)
			}},
		},
	}, nil)
}
