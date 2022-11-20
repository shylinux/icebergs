package cli

import (
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
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
)

const MIRRORS = "mirrors"

func init() {
	Index.MergeCommands(ice.Commands{
		MIRRORS: {Name: "mirrors cli auto", Help: "软件镜像", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH), "")
				IsAlpine(m, "curl")
				IsAlpine(m, "make")
				IsAlpine(m, "gcc")
				IsAlpine(m, "vim")
				IsAlpine(m, "tmux")
				if IsAlpine(m, "git"); !IsAlpine(m, "go", "go git") {
					mdb.ZoneInsert(m, CLI, "go", CMD, kit.Format("install download https://golang.google.cn/dl/go1.15.5.%s-%s.tar.gz usr/local", runtime.GOOS, runtime.GOARCH))
				}
			}},
			mdb.INSERT: {Name: "insert cli osid cmd"},
			ALPINE: {Name: "alpine cli cmd", Hand: func(m *ice.Message, arg ...string) { IsAlpine(m, arg...) }},
			CMD: {Name: "cmd cli osid", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
				osid := kit.Select(m.Conf(RUNTIME, kit.Keys(HOST, OSID)), m.Option(OSID))
				mdb.ZoneSelectCB(m, m.Option(CLI), func(value ice.Map) {
					if osid != "" && strings.Contains(osid, kit.Format(value[OSID])) {
						m.Cmdy(kit.Split(kit.Format(value[CMD])))
					}
				})
			}},
		}, mdb.ZoneAction(mdb.SHORT, CLI, mdb.FIELD, "time,id,osid,cmd"))},
	})
}

func osid(m *ice.Message, sys string) bool {
	osid := runtime.GOOS
	m.Option(ice.MSG_USERROLE, aaa.ROOT)
	m.Cmd(nfs.CAT, "/etc/os-release", func(text string) {
		if ls := kit.Split(text, ice.EQ); len(ls) > 1 {
			switch ls[0] {
			case "ID", "ID_LIKE":
				osid = strings.TrimSpace(ls[1] + ice.SP + osid)
			}
		}
	})
	if strings.Contains(osid, sys) {
		return true
	}
	return false
}
func insert(m *ice.Message, cmd string, arg ...string) bool {
	if len(arg) > 0 {
		m.Go(func() {
			m.Sleep300ms()
			m.Cmd(mdb.INSERT, kit.Keys(CLI, MIRRORS), "", mdb.ZONE, arg[0], OSID, ALPINE, CMD, cmd+ice.SP+kit.Select(arg[0], arg, 1))
		})
	}
	return true
}
func IsAlpine(m *ice.Message, arg ...string) bool { return osid(m, ALPINE) && insert(m, "system apk add", arg...) }
func IsCentos(m *ice.Message, arg ...string) bool { return osid(m, CENTOS) && insert(m, "yum install -y", arg...) }
func IsUbuntu(m *ice.Message, arg ...string) bool { return osid(m, UBUNTU) && insert(m, "apt get -y", arg...) }
