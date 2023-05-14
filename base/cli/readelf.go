package cli

import (
	"bytes"
	"encoding/binary"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type elf struct {
	EI_CLASS    int
	EI_DATA     int
	EI_VERSION  int
	e_type      uint16
	e_machine   uint16
	e_version   uint32
	e_entry     uint64
	e_phoff     uint64
	e_shoff     uint64
	e_flags     uint32
	e_ehsize    uint16
	e_phentsize uint16
	e_phnum     uint16
	e_shentsize uint16
	e_shnum     uint16
	e_shstrndx  uint16
}

func read2(buf []byte, offset int) (uint16, int) {
	return binary.LittleEndian.Uint16(buf[offset : offset+2]), offset + 2
}
func read4(buf []byte, offset int) (uint32, int) {
	return binary.LittleEndian.Uint32(buf[offset : offset+4]), offset + 4
}
func read8(buf []byte, offset int) (uint64, int) {
	return binary.LittleEndian.Uint64(buf[offset : offset+8]), offset + 8
}
func readelf(buf []byte) (elf elf) {
	i := 16
	elf.EI_CLASS = int(buf[4])
	elf.EI_DATA = int(buf[5])
	elf.EI_VERSION = int(buf[6])
	elf.e_type, i = read2(buf, i)
	elf.e_machine, i = read2(buf, i)
	elf.e_version, i = read4(buf, i)
	elf.e_entry, i = read8(buf, i)
	elf.e_phoff, i = read8(buf, i)
	elf.e_shoff, i = read8(buf, i)
	elf.e_flags, i = read4(buf, i)
	elf.e_ehsize, i = read2(buf, i)
	elf.e_phentsize, i = read2(buf, i)
	elf.e_phnum, i = read2(buf, i)
	elf.e_shentsize, i = read2(buf, i)
	elf.e_shnum, i = read2(buf, i)
	elf.e_shstrndx, i = read2(buf, i)
	return elf
}

func init() {
	Index.MergeCommands(ice.Commands{
		"readelf": {Name: "readelf path=usr/publish/ice.linux.amd64 auto", Hand: func(m *ice.Message, arg ...string) {
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
				if bytes.Equal(buf[:4], []byte{0x7f, 0x45, 0x4c, 0x46}) {
					m.Echo("elf %#v", readelf(buf))
				}
				for i := 0; i < n; i++ {
					kit.If(i%16 == 0, func() { m.Push("addr", kit.Format("%04x", i)) })
					m.Push(kit.Format("%02x", i%16), kit.Format("%02x", buf[i]))
				}
			}
		}},
	})
}
