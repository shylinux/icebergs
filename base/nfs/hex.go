package nfs

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/hex"
	"io"
	"os"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const HEX = "hex"

func init() {
	Index.MergeCommands(ice.Commands{HEX: {Name: "hex path compress=raw,gzip,zlib size auto", Help: "二进制", Hand: func(m *ice.Message, arg ...string) {
		if len(arg) == 0 || arg[0] == "" || strings.HasSuffix(arg[0], ice.PS) {
			m.Cmdy(DIR, kit.Slice(arg, 0, 1))
			return
		}
		if f, e := os.Open(arg[0]); !m.Warn(e, ice.ErrNotFound, arg[0]) {
			defer f.Close()
			s, _ := f.Stat()
			var r io.Reader = f
			switch arg[1] {
			case "gzip":
				if g, e := gzip.NewReader(r); !m.Warn(e) {
					r = g
				}
			case "zlib":
				if z, e := zlib.NewReader(r); !m.Warn(e) {
					r = z
				}
			}
			buf := make([]byte, kit.Int(kit.Select("1024", arg, 2)))
			n, _ := r.Read(buf)
			for i := 0; i < n; i++ {
				if i%8 == 0 {
					m.Push("n", kit.Format("%04d", i))
				}
				if m.Push(kit.Format(i%8), hex.EncodeToString(buf[i:i+1])); i%8 == 7 {
					m.Push("text", string(buf[i-7:i+1]))
				}
			}
			m.Status(mdb.TIME, s.ModTime().Format(ice.MOD_TIME), FILE, arg[0], SIZE, kit.FmtSize(s.Size()))
		}
	}}})
}
