package ice

import (
	"bufio"
	"os"
	"path"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) ActionKey() string  { return strings.TrimPrefix(strings.TrimSuffix(m._sub, PS), PS) }
func (m *Message) CommandKey() string { return strings.TrimPrefix(strings.TrimSuffix(m._key, PS), PS) }
func (m *Message) PrefixKey() string  { return m.Prefix(m.CommandKey()) }
func (m *Message) PrefixPath(arg ...Any) string {
	return strings.TrimPrefix(path.Join(strings.ReplaceAll(m.PrefixRawKey(arg...), PT, PS)), "web") + PS
}
func (m *Message) PrefixRawKey(arg ...Any) string { return m.Prefix(m._key, kit.Keys(arg...)) }
func (m *Message) Prefix(arg ...string) string    { return m.Target().Prefix(arg...) }

func SaveImportant(m *Message, arg ...string) {
	if Info.Important != true {
		return
	}
	for i, v := range arg {
		kit.If(v == "" || strings.Contains(v, SP), func() { arg[i] = "\"" + v + "\"" })
	}
	m.Cmd("nfs.push", VAR_DATA_IMPORTANT, kit.Join(arg, SP), NL)
}
func loadImportant(m *Message) {
	if f, e := os.Open(VAR_DATA_IMPORTANT); e == nil {
		defer f.Close()
		for bio := bufio.NewScanner(f); bio.Scan(); {
			if bio.Text() == "" || strings.HasPrefix(bio.Text(), "# ") {
				continue
			}
			m.Cmd(kit.Split(bio.Text()))
		}
	}
	Info.Important = true
}
func removeImportant(m *Message) { os.Remove(VAR_DATA_IMPORTANT) }
