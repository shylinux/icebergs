package cli

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func IsAlpine(m *ice.Message, arg ...string) bool {
	if strings.Contains(m.Conf(RUNTIME, "host.OSID"), ALPINE) {
		if len(arg) > 0 {
			m.Cmd(MIRROR, mdb.CREATE, "cli", arg[0], "cmd", arg[1])
		}
		return true
	}
	return false
}

const (
	OSID   = "OSID"
	ALPINE = "alpine"
	CENTOS = "centos"
	UBUNTU = "ubuntu"
)

const MIRROR = "mirror"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		MIRROR: {Name: "mirror cli auto", Help: "软件镜像", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					m.Sleep("1s")
					IsAlpine(m, "curl", "system apk add curl")
					IsAlpine(m, "make", "system apk add make")
					IsAlpine(m, "gcc", "system apk add gcc")
					IsAlpine(m, "vim", "system apk add vim")
					IsAlpine(m, "tmux", "system apk add tmux")

					if IsAlpine(m, "git", "system apk add git"); !IsAlpine(m, "go", "system apk add git go") {
						m.Cmd(MIRROR, mdb.CREATE, kit.SimpleKV("cli,cmd", "go", "install download https://golang.google.cn/dl/go1.15.5.linux-amd64.tar.gz usr/local"))
					}

					IsAlpine(m, "node", "system apk add nodejs")
					IsAlpine(m, "java", "system apk add openjdk8")
					IsAlpine(m, "javac", "system apk add openjdk8")
					IsAlpine(m, "mvn", "system apk add openjdk8 maven")
					IsAlpine(m, "python", "system apk add python2")
					IsAlpine(m, "python2", "system apk add python2")
					IsAlpine(m, "python3", "system apk add python3")
				})
			}},
			mdb.CREATE: {Name: "create cli cmd", Help: "创建"},
		}, mdb.HashAction(mdb.SHORT, "cli", mdb.FIELD, "time,cli,cmd"))},
	}})
}
