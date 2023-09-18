package cli

import (
	"io"
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
	REPOS   = "repos"
	UBUNTU  = "ubuntu"
	CENTOS  = "centos"
	ALPINE  = "alpine"
	BUSYBOX = "busybox"
	RELEASE = "release"

	ETC_OS_RELEASE = "/etc/os-release"
	ETC_APK_REPOS  = "/etc/apk/repositories"
)

const MIRRORS = "mirrors"

func init() {
	Index.MergeCommands(ice.Commands{
		MIRRORS: {Name: "mirrors cli auto repos", Help: "软件镜像", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert cli* osid cmd*"},
			CMD: {Name: "cmd cli osid", Hand: func(m *ice.Message, arg ...string) {
				osid := kit.Select(mdb.Conf(m, RUNTIME, kit.Keys(HOST, OSID)), m.Option(OSID))
				mdb.ZoneSelectCB(m, m.Option(CLI), func(value ice.Map) {
					kit.If(strings.Contains(osid, kit.Format(value[OSID])), func() { m.Cmdy(kit.Split(kit.Format(value[CMD]))) })
				})
			}},
			ALPINE: {Name: "alpine cli cmd", Hand: func(m *ice.Message, arg ...string) { IsAlpine(m, arg...) }},
			REPOS: {Help: "镜像源", Hand: func(m *ice.Message, arg ...string) {
				switch {
				case strings.Contains(release(m.Spawn()), ALPINE):
					ice.Info.PushStream(m)
					m.Optionv(CMD_OUTPUT).(io.Writer).Write([]byte("\n"))
					defer ice.Info.PushNotice(m, "toast", "success")
					m.Cmd(nfs.SAVE, ETC_APK_REPOS, strings.ReplaceAll(m.Cmdx(nfs.CAT, ETC_APK_REPOS), "dl-cdn.alpinelinux.org", "mirrors.tencent.com"))
					m.Cmdy(SYSTEM, "apk", "update")
					m.StatusTimeCount()
				}
			}},
			"add": {Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				mdb.ZoneSelect(m, m.Option(CLI)).Table(func(value ice.Maps) {
					ice.Info.PushStream(m)
					ice.Info.PushNotice(m, "toast", "process", "", "-1")
					defer ice.Info.PushNotice(m, "toast", "success")
					m.Push("res", m.Cmdx(kit.Split(value[CMD])))
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, CLI, mdb.FIELD, "time,id,osid,cmd"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Table(func(value ice.Maps) {
					p := SystemFind(m, value[CLI])
					m.Push("path", p)
					if p == "" {
						m.PushButton("add")
					} else {
						m.PushButton("")
					}
				})
				m.StatusTimeCount("release", release(m.Spawn()))
			}
			switch {
			case strings.Contains(release(m.Spawn()), ALPINE):
				m.Cmdy(nfs.CAT, ETC_APK_REPOS)
			}
		}},
	})
}

var _release = ""

func release(m *ice.Message) string {
	list := []string{runtime.GOOS}
	if list[0] != LINUX || !nfs.Exists(m, ETC_OS_RELEASE) {
		return list[0]
	}
	m.Option(nfs.CAT_CONTENT, _release)
	_release = m.Cmdx(nfs.CAT, ETC_OS_RELEASE, kit.Dict(ice.MSG_USERROLE, aaa.ROOT), func(text string, _ int) string {
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
