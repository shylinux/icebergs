package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心"}

func init() { web.Index.Register(Index, &web.Frame{}) }

func MergePod(m *ice.Message, pod string, arg ...ice.Any) string {
	return kit.MergePOD(kit.Select(ice.Info.Domain, m.Option(ice.MSG_USERWEB)), pod, arg...)
}
func MergeCmd(m *ice.Message, cmd string, arg ...ice.Any) string {
	return mergeurl(m, path.Join(ice.CMD, kit.Select(m.PrefixKey(), cmd)), arg...)
}
func MergeWebsite(m *ice.Message, web string, arg ...ice.Any) string {
	if m.Option(ice.MSG_USERPOD) == "" {
		return mergeurl(m, path.Join(ice.PS, "chat", WEBSITE, web), arg...)
	} else {
		return mergeurl(m, path.Join(WEBSITE, web), arg...)
	}
}
func mergeurl(m *ice.Message, p string, arg ...ice.Any) string {
	if m.Option(ice.MSG_USERPOD) == "" {
		p = path.Join(ice.PS, p)
	} else {
		p = path.Join(ice.PS, "chat", ice.POD, m.Option(ice.MSG_USERPOD), p)
	}
	return kit.MergeURL2(kit.Select(ice.Info.Domain, m.Option(ice.MSG_USERWEB)), path.Join(p), arg...)
}
