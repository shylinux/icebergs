package cli

import (
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	CMD     = "cmd"
	OSID    = "osid"
	UBUNTU  = "ubuntu"
	CENTOS  = "centos"
	ALPINE  = "alpine"
	BUSYBOX = "busybox"
	RELEASE = "release"
)

const MIRRORS = "mirrors"

func init() {
	Index.MergeCommands(ice.Commands{
		MIRRORS: {Name: "mirrors cli auto", Help: "软件镜像", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert cli* osid cmd*"},
			CMD: {Name: "cmd cli osid", Hand: func(m *ice.Message, arg ...string) {
				osid := kit.Select(mdb.Conf(m, RUNTIME, kit.Keys(HOST, OSID)), m.Option(OSID))
				mdb.ZoneSelectCB(m, m.Option(CLI), func(value ice.Map) {
					kit.If(strings.Contains(osid, kit.Format(value[OSID])), func() { m.Cmdy(kit.Split(kit.Format(value[CMD]))) })
				})
			}},
			ALPINE: {Name: "alpine cli cmd", Hand: func(m *ice.Message, arg ...string) { IsAlpine(m, arg...) }},
		}, mdb.ZoneAction(mdb.SHORT, CLI, mdb.FIELD, "time,id,osid,cmd"), mdb.ClearOnExitHashAction())},
	})
}

var _release = ""

func release(m *ice.Message) string {
	list := []string{runtime.GOOS}
	if list[0] != LINUX || !nfs.Exists(m, "/etc/os-release") {
		return list[0]
	}
	m.Option(nfs.CAT_CONTENT, _release)
	_release = m.Cmdx(nfs.CAT, "/etc/os-release", kit.Dict(ice.MSG_USERROLE, aaa.ROOT), func(text string, _ int) string {
		if ls := kit.Split(text, mdb.EQ); len(ls) > 1 {
			kit.Switch(ls[0], []string{"ID", "ID_LIKE"}, func() { list = append(list, strings.TrimSpace(ls[1])) })
		}
		return text
	})
	return strings.Join(list, lex.SP)
}
func insert(m *ice.Message, sys, cmd string, arg ...string) bool {
	if !strings.Contains(release(m), sys) {
		return false
	}
	if len(arg) > 0 {
		m.GoSleep("300ms", mdb.INSERT, kit.Keys(CLI, MIRRORS), "", mdb.ZONE, arg[0], OSID, sys, CMD, cmd+lex.SP+kit.Select(arg[0], arg, 1))
	}
	return true
}
func IsAlpine(m *ice.Message, arg ...string) bool {
	return insert(m, ALPINE, "system apk add", arg...)
}
func IsCentos(m *ice.Message, arg ...string) bool {
	return insert(m, CENTOS, "system yum install -y", arg...)
}
func IsUbuntu(m *ice.Message, arg ...string) bool {
	return insert(m, UBUNTU, "system apt get -y", arg...)
}
func IsSystem(m *ice.Message, arg ...string) bool {
	return IsAlpine(m, arg...) || IsCentos(m, arg...) || IsUbuntu(m, arg...)
}
