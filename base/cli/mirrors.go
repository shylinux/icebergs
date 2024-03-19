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
	ADD     = "add"
	OSID    = "osid"
	REPOS   = "repos"
	ALPINE  = "alpine"
	BUSYBOX = "busybox"
	RELEASE = "release"

	ETC_OS_RELEASE = "/etc/os-release"
	ETC_APK_REPOS  = "/etc/apk/repositories"
)

const MIRRORS = "mirrors"

func init() {
	Index.MergeCommands(ice.Commands{
		MIRRORS: {Help: "软件镜像", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert cli* osid cmd*"},
			CMD: {Name: "cmd cli osid", Hand: func(m *ice.Message, arg ...string) {
				osid := kit.Select(mdb.Conf(m, RUNTIME, kit.Keys(HOST, OSID)), m.Option(OSID))
				mdb.ZoneSelectCB(m, m.Option(CLI), func(value ice.Map) {
					kit.If(strings.Contains(osid, kit.Format(value[OSID])), func() {
						m.Cmdy(kit.Split(kit.Format(value[CMD])))
					})
				})
			}},
			ADD: {Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				ice.Info.PushStream(m)
				mdb.ZoneSelect(m, m.Option(CLI)).Table(func(value ice.Maps) {
					m.ToastProcess()
					if msg := m.Cmd(kit.Split(value[CMD])); IsSuccess(msg) {
						m.ToastSuccess()
					} else {
						m.ToastFailure()
					}
				})
			}},
			REPOS: {Name: "repos proxy=mirrors.tencent.com", Help: "镜像", Hand: func(m *ice.Message, arg ...string) {
				switch {
				case strings.Contains(_release, ALPINE):
					defer m.PushStream().ToastProcess()()
					kit.If(m.Option("proxy"), func(p string) {
						m.Cmd(nfs.SAVE, ETC_APK_REPOS, strings.ReplaceAll(m.Cmdx(nfs.CAT, ETC_APK_REPOS), "dl-cdn.alpinelinux.org", p))
					})
					m.Cmdy(SYSTEM, "apk", "update")
				}
			}},
			ALPINE: {Name: "alpine cli cmd", Hand: func(m *ice.Message, arg ...string) { IsAlpine(m, arg...) }},
		}, mdb.ZoneAction(mdb.SHORT, CLI, mdb.FIELDS, "time,id,osid,cmd"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Table(func(value ice.Maps) {
					p := SystemFind(m, value[CLI])
					if m.Push(nfs.PATH, p); p == "" {
						m.PushButton(ADD)
					} else {
						m.PushButton("")
					}
				}).Action(REPOS).StatusTimeCount("release", _release)
			}
			switch {
			case strings.Contains(_release, ALPINE):
				m.Cmdy(nfs.CAT, ETC_APK_REPOS)
			}
			// m.EchoScript(kit.Format("cd %s; %s", kit.Path(""), kit.JoinCmds(kit.Simple(kit.Path(os.Args[0]), os.Args[1:])...)))
		}},
	})
}

var _release = ""

func release(m *ice.Message) string {
	list := []string{runtime.GOOS}
	if list[0] != LINUX || !nfs.Exists(m, ETC_OS_RELEASE) {
		return list[0]
	}
	m.Cmd(nfs.CAT, ETC_OS_RELEASE, kit.Dict(ice.MSG_USERROLE, aaa.ROOT), func(text string, _ int) string {
		if ls := kit.Split(text, mdb.EQ); len(ls) > 1 {
			kit.Switch(ls[0], []string{"ID", "ID_LIKE"}, func() { list = append(list, strings.TrimSpace(ls[1])) })
		}
		return text
	})
	_release = kit.JoinWord(list...)
	return _release
}
func insert(m *ice.Message, sys, cmd string, arg ...string) bool {
	if !strings.Contains(_release, sys) {
		return false
	}
	if len(arg) > 0 {
		m.GoSleep300ms(func() {
			m.Cmd(mdb.INSERT, kit.Keys(CLI, MIRRORS), "", mdb.ZONE, arg[0], OSID, sys, CMD, cmd+lex.SP+kit.Select(arg[0], arg, 1))
		})
	}
	return true
}
func IsAlpine(m *ice.Message, arg ...string) bool {
	return insert(m, ALPINE, "system apk add", arg...)
}
func IsRedhat(m *ice.Message, arg ...string) bool {
	return insert(m, "rhel", "system yum install -y", arg...)
}
func IsSystem(m *ice.Message, arg ...string) bool {
	return IsAlpine(m, arg...) || IsRedhat(m, arg...)
}
