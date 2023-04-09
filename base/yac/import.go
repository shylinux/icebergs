package yac

import (
	"io"
	"net/url"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	PACKAGE = "package"
	IMPORT  = "import"
)

func init() {
	Index.MergeCommands(ice.Commands{
		PACKAGE: {Name: "package main", Hand: func(m *ice.Message, arg ...string) {
			s := _parse_stack(m)
			s.skip = len(s.rest)
		}},
		IMPORT: {Name: "import ice shylinux.com/x/icebergs", Hand: func(m *ice.Message, arg ...string) {
			load := func(pre string, u *url.URL, p string, r io.Reader) {
				if kit.Ext(p) == nfs.SHY {
					s := _parse_stack(m)
					f := s.pushf(m, "", pre, p)
					defer s.popf(m)
					kit.For(u.Query(), func(k string, v []string) { f.value[k] = v[0] })
					s.parse(m, p, r)
					kit.If(pre != "_", func() { kit.For(f.value, func(k string, v Any) { s.frame[0].value[kit.Keys(pre, k)] = v }) })
				}
			}
			find := func(pre, url string) {
				kit.If(url == "\"shylinux.com/x/ice\"", func() { pre, url = "ice", "\"shylinux.com/x/release\"" })
				u := kit.ParseURL(strings.TrimSuffix(strings.TrimPrefix(url, "\""), "\""))
				pre = kit.Select(path.Base(u.Path), pre)
				kit.If(pre == nfs.PT, func() { pre = "" })
				m.Debug("import %v %v", pre, url)
				if ls := kit.Split(u.Path, nfs.PS); path.Join(kit.Slice(ls, 0, 3)...) == ice.Info.Make.Module && nfs.Exists(m, path.Join(kit.Slice(ls, 3)...)) {
					nfs.Open(m, path.Join(kit.Slice(ls, 3)...)+nfs.PS, func(r io.Reader, p string) { load(pre, u, p, r) })
				} else if p := path.Join(ice.USR_REQUIRE, u.Path); nfs.Exists(m, p) {
					nfs.Open(m, p+nfs.PS, func(r io.Reader, p string) { load(pre, u, p, r) })
				} else if p := nfs.USR + path.Join(kit.Slice(ls, 2)...); nfs.Exists(m, p) {
					nfs.Open(m, p+nfs.PS, func(r io.Reader, p string) { load(pre, u, p, r) })
				}
			}
			s := _parse_stack(m)
			if p := s.next(m); p == OPEN {
				for s.token() != CLOSE {
					if list := s.nextLine(m); s.token() != CLOSE {
						pos := s.Position
						find(kit.Select("", list[0], len(list) > 1), kit.Select("", list, -1))
						s.Position = pos
					}
				}
			} else {
				find(kit.Select("", s.rest[1], len(s.rest) > 2), kit.Select("", s.rest, -1))
			}
		}},
	})
}
