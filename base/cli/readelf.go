package cli

import (
	"debug/elf"
	"debug/macho"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		"readelf": {Name: "readelf path=bin/ice.bin auto", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], nfs.PS) {
				m.Cmdy(nfs.DIR, arg)
				return
			}
			if f, e := nfs.OpenFile(m, arg[0]); !m.Warn(e) {
				defer f.Close()
				buf := make([]byte, 1024)
				n, e := f.Read(buf)
				if m.Warn(e) {
					return
				}
				switch runtime.GOOS {
				case LINUX:
					f, _ := elf.Open(arg[0])
					m.Echo("%v", kit.Formats(f))
				case DARWIN:
					f, _ := macho.Open(arg[0])
					m.Echo("%v", kit.Formats(f))
				}
				for i := 0; i < n; i++ {
					kit.If(i%16 == 0, func() { m.Push("addr", kit.Format("%04x", i)) })
					m.Push(kit.Format("%02x", i%16), kit.Format("%02x", buf[i]))
				}
				m.StatusTimeCount()
			}
		}},
	})
}
