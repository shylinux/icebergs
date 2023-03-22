package ice

import (
	"bufio"
	"os"
	"path"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) ActionKey() string {
	return strings.TrimPrefix(strings.TrimSuffix(m._sub, PS), PS)
}
func (m *Message) CommandKey() string {
	return strings.TrimPrefix(strings.TrimSuffix(m._key, PS), PS)
}
func (m *Message) PrefixRawKey(arg ...Any) string {
	return kit.Keys(m.Prefix(m._key), kit.Keys(arg...))
}
func (m *Message) PrefixKey(arg ...Any) string {
	return kit.Keys(m.Prefix(m.CommandKey()), kit.Keys(arg...))
}
func (m *Message) Prefix(arg ...string) string {
	return m.Target().PrefixKey(arg...)
}
func (m *Message) PrefixPath(arg ...Any) string {
	return strings.TrimPrefix(path.Join(strings.ReplaceAll(m.PrefixRawKey(arg...), PT, PS)), "web") + PS
}

func (m *Message) Config(key string, arg ...Any) string {
	return kit.Format(m.Configv(key, arg...))
}
func (m *Message) Configv(key string, arg ...Any) Any {
	if len(arg) > 0 {
		m.Confv(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Confv(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) ConfigSimple(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		res = append(res, k, m.Config(k))
	}
	return
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
func SaveImportant(m *Message, arg ...string) {
	if Info.Important != true {
		return
	}
	for i, v := range arg {
		if v == "" || strings.Contains(v, SP) {
			arg[i] = "\"" + v + "\""
		}
	}
	m.Cmd("nfs.push", VAR_DATA_IMPORTANT, kit.Join(arg, SP), NL)
}
func removeImportant(m *Message) { os.Remove(VAR_DATA_IMPORTANT) }
