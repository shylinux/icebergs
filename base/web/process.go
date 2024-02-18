package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func ProcessIframe(m *ice.Message, title, link string, arg ...string) *ice.Message {
	if m.IsMetaKey() {
		m.ProcessOpen(link)
		return m
	} else if !kit.HasPrefixList(arg, ctx.RUN) {
		defer m.Push(TITLE, title)
	}
	return ctx.ProcessFloat(m, CHAT_IFRAME, link, arg...)
}
func ProcessPodCmd(m *ice.Message, pod, cmd string, args ice.Any, arg ...string) *ice.Message {
	if kit.HasPrefixList(arg, ctx.RUN) {
		pod, cmd, arg = arg[1], arg[2], kit.Simple(arg[0], arg[3:])
	} else {
		cmd = kit.Select(m.ActionKey(), cmd)
		defer processSpace(m, pod, pod, cmd)
	}
	return ctx.ProcessFloat(m.Options(ice.POD, pod), cmd, args, arg...)
}
func ProcessHashPodCmd(m *ice.Message, arg ...string) (msg *ice.Message) {
	if kit.HasPrefixList(arg, ctx.RUN) {
		msg = mdb.HashSelects(m.Spawn(), arg[1])
		arg = kit.Simple(arg[0], arg[2:])
	} else {
		msg = mdb.HashSelects(m.Spawn(), m.Option(mdb.HASH))
		defer processSpace(m, msg.Append(SPACE), m.Option(mdb.HASH))
	}
	return ctx.ProcessFloat(m.Options(ice.POD, msg.Append(SPACE)), msg.Append(ctx.INDEX), kit.Split(msg.Append(ctx.ARGS)), arg...)
}
func processSpace(m *ice.Message, pod string, arg ...string) {
	m.ProcessField(ctx.ACTION, m.ActionKey(), ctx.RUN, arg)
	m.RewriteAppend(func(value, key string, index int) string { return kit.Select("", value, key != SPACE) })
	m.Push(ice.MSG_SPACE, pod)
}
