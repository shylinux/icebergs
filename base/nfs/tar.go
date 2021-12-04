package nfs

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

const TAR = "tar"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TAR: {Name: "tar file path auto", Help: "打包", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(cli.CMD_DIR, m.Option(DIR_ROOT))
			m.Cmdy(cli.SYSTEM, "tar", "zcvf", arg)
			return

			file, err := os.Create(arg[0])
			m.Assert(err)
			defer file.Close()

			var out io.WriteCloser = file
			if strings.HasSuffix(arg[0], ".gz") {
				out = gzip.NewWriter(out)
				defer out.Close()
			}
			t := tar.NewWriter(out)
			defer t.Close()

			dir_root := m.Option(DIR_ROOT)
			var count, total int64
			for _, k := range arg[1:] {
				m.Option(DIR_TYPE, TYPE_CAT)
				m.Option(DIR_DEEP, ice.TRUE)
				m.Cmdy(DIR, k, func(f os.FileInfo, p string) {
					total += f.Size()

					header, err := tar.FileInfoHeader(f, p)
					if m.Warn(err != nil, err) {
						return
					}

					header.Name = strings.TrimPrefix(p, dir_root+ice.PS)
					if err = t.WriteHeader(header); m.Warn(err != nil, "err: %v", err) {
						return
					}

					file, err := os.Open(p)
					if m.Warn(err != nil, "err: %v", err) {
						return
					}
					defer file.Close()

					m.PushNoticeGrow(kit.Format("%v %v %v\n", header.Name, kit.FmtSize(f.Size()), kit.FmtSize(total)))
					if _, err = io.Copy(t, file); m.Warn(err != nil, "err: %v", err) {
						return
					}

					count++
					m.Toast(kit.Format("%v %v %v", count, m.Cost(), kit.FmtSize(total)))
				})
			}
			m.StatusTimeCountTotal(kit.FmtSize(total))
		}},
	}})
}