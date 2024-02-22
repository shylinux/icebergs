package log

import (
	ice "shylinux.com/x/icebergs"
)

const BENCH = "bench"

func init() {
	Index.MergeCommands(ice.Commands{
		BENCH: {Help: "记录", Hand: func(m *ice.Message, arg ...string) {}},
	})
}
