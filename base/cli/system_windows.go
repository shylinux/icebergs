package cli

import (
	"os/exec"

	ice "shylinux.com/x/icebergs"
)

func _system_cmds(m *ice.Message, cmd *exec.Cmd, arg ...string) *exec.Cmd {
	return cmd
}
