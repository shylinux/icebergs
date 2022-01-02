package chrome

import (
	"shylinux.com/x/ice"
)

type video struct {
	operate

	play string `name:"play" help:"播放"`
	next string `name:"next" help:"下一集"`
	list string `name:"list tags='ul.stui-content__playlist.column10.clearfix li a' rate=1.5 skip=140 next=2520 auto play next" help:"操作"`
}

func (v video) List(m *ice.Message, arg ...string) {
	m.DisplayStory("video.js")
	m.Echo("hello world")
}

func init() { ice.CodeCtxCmd(video{}) }
