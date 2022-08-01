package cli

import (
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	CMD    = "cmd"
	OSID   = "osid"
	ALPINE = "alpine"
	CENTOS = "centos"
	UBUNTU = "ubuntu"
)

const MIRRORS = "mirrors"

func init() {
	Index.MergeCommands(ice.Commands{
		MIRRORS: {Name: "mirrors cli auto", Help: "软件镜像", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH), "")
				IsAlpine(m, "curl")
				IsAlpine(m, "make")
				IsAlpine(m, "gcc")
				IsAlpine(m, "vim")
				IsAlpine(m, "tmux")

				IsAlpine(m, "git")
				mdb.ZoneInsert(m, CLI, "go", CMD, kit.Format("install download https://golang.google.cn/dl/go1.15.5.%s-%s.tar.gz usr/local", runtime.GOOS, runtime.GOARCH))

				IsAlpine(m, "node", "nodejs")
				IsAlpine(m, "java", "openjdk8")
				IsAlpine(m, "javac", "openjdk8")
				IsAlpine(m, "mvn", "openjdk8 maven")
				IsAlpine(m, "python", "python2")
				IsAlpine(m, "python2")
				IsAlpine(m, "python3")
			}},
			mdb.INSERT: {Name: "insert cli osid cmd", Help: "添加"},
			CMD: {Name: "cmd cli osid", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				osid := kit.Select(m.Conf(RUNTIME, kit.Keys(HOST, OSID)), m.Option(OSID))
				mdb.ZoneSelectCB(m, m.Option(CLI), func(value ice.Map) {
					if osid != "" && strings.Contains(osid, kit.Format(value[OSID])) {
						m.Cmdy(kit.Split(kit.Format(value[CMD])))
					}
				})
			}},
			ALPINE: {Name: "alpine cli cmd", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				IsAlpine(m, arg...)
			}},
		}, mdb.ZoneAction(mdb.SHORT, CLI, mdb.FIELD, "time,id,osid,cmd"))},
	})
}

func IsAlpine(m *ice.Message, arg ...string) bool {
	if strings.Contains(m.Conf(RUNTIME, kit.Keys(HOST, OSID)), ALPINE) {
		if len(arg) > 0 {
			mdb.ZoneInsert(m, CLI, arg[0], OSID, ALPINE, CMD, "system apk add "+kit.Select(arg[0], arg, 1))
		}
		return true
	}
	return false
}
