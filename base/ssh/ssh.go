package ssh

import (
	ice "shylinux.com/x/icebergs"
)

const SSH = "ssh"

var Index = &ice.Context{Name: SSH, Help: "终端模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if f, ok := m.Target().Server().(*Frame); ok {
			f.close()
		}
		m.Save()
	}},
}}

func init() {
	ice.Index.Register(Index, &Frame{},
		SOURCE, TARGET, PROMPT, PRINTF, SCREEN, RETURN,
	)
}
