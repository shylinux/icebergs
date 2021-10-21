package ssh

import (
	ice "shylinux.com/x/icebergs"
)

const SSH = "ssh"

var Index = &ice.Context{Name: SSH, Help: "终端模块"}

func init() {
	ice.Index.Register(Index, &Frame{}, SOURCE, TARGET, PROMPT, PRINTF, SCREEN, RETURN)
}
