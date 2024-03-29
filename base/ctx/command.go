package ctx

import (
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _command_list(m *ice.Message, name string) *ice.Message {
	if name == "" {
		for k, v := range m.Source().Commands {
			if IsOrderCmd(k) {
				continue
			}
			m.Push(mdb.KEY, k).Push(mdb.NAME, v.Name).Push(mdb.HELP, v.Help)
		}
		return m.Sort(mdb.KEY)
	}
	if strings.HasPrefix(name, "can.") {
		return m.Push(mdb.INDEX, name).Push(mdb.NAME, name).Push(mdb.HELP, "").Push(mdb.META, "").Push(mdb.LIST, "")
	}

	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Push(mdb.INDEX, kit.Keys(s.Prefix(), key))
		m.Push(mdb.NAME, kit.Format(cmd.Name))
		m.Push(mdb.HELP, kit.Format(cmd.Help))
		m.Push(mdb.META, FormatPretty(cmd.Meta, 0, 2))
		m.Push(mdb.LIST, FormatPretty(cmd.List, 0, 2))
	})
	return m
}
func _command_search(m *ice.Message, kind, name, text string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if IsOrderCmd(key) || !strings.Contains(s.Prefix(key), name) {
			return
		}
		m.PushSearch(ice.CTX, kit.PathName(1), ice.CMD, kit.FileName(1), kit.SimpleKV("", s.Prefix(), cmd.Name, cmd.Help),
			CONTEXT, s.Prefix(), COMMAND, key, INDEX, kit.Keys(s.Prefix(), key), mdb.HELP, cmd.Help, nfs.FILE, FileURI(cmd.FileLine()),
		)
	})
	m.Sort(m.OptionFields())
}

const (
	INDEX   = "index"
	OPTS    = "opts"
	ARGS    = "args"
	SHIP    = "ship"
	STYLE   = "style"
	DISPLAY = "display"
	ACTION  = "action"
	RUN     = "run"
)
const COMMAND = "command"

func init() {
	Index.MergeCommands(ice.Commands{
		COMMAND: {Name: "command key auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				TravelCmd(m, func(key, file, line string) { AddFileCmd(file, key) })
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || len(arg) > 1 && arg[1] != "" {
					_command_search(m, arg[0], kit.Select("", arg, 1), kit.Select("", arg, 2))
				}
			}},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				TravelCmd(m, func(key, file, line string) { m.Push(mdb.NAME, key).Push(nfs.FILE, file).Push(nfs.LINE, line) }).Sort(mdb.NAME).Table(func(value ice.Maps) {
					m.Echo(`%s	%s	%s;" f`+lex.NL, value[mdb.NAME], value[nfs.FILE], value[nfs.LINE])
				}).Cmd(nfs.SAVE, nfs.TAGS, m.Result())
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] != "" && arg[0] != ice.EXIT {
					m.Search(arg[0], func(key string, cmd *ice.Command) {
						field := kit.Format(kit.Value(cmd.List, kit.Format("%d.name", len(arg)-1)))
						if m.Cmdy(arg[0], mdb.INPUTS, field); m.Length() == 0 {
							m.Cmdy(arg).Cut(field)
						}
					})
				}
			}},
		}, CmdAction(), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy("", mdb.SEARCH, COMMAND, ice.OptionFields(INDEX)).StatusTimeCount()
				DisplayStory(m.Options(nfs.DIR_ROOT, "ice."), "spide.js?split=.")
				return
			}
			kit.If(len(arg) == 0, func() { arg = append(arg, "") })
			kit.For(arg, func(k string) { _command_list(m, k) })
		}},
	})
}

var Upload = func(*ice.Message) []string { return nil }

func CmdInputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case INDEX:
		m.Cmdy(COMMAND, mdb.SEARCH, COMMAND, ice.OptionFields(INDEX))
	case ARGS:
		m.Cmdy(m.Option(INDEX))
	}
}
func PodCmd(m *ice.Message, arg ...ice.Any) bool {
	Upload(m)
	if pod := m.Option(ice.POD); pod != "" {
		m.Options(ice.POD, []string{}, ice.MSG_USERPOD, pod).Cmdy(append(kit.List(ice.SPACE, pod), arg...)...)
		return true
	}
	return false
}
func Run(m *ice.Message, arg ...string) {
	kit.If(!PodCmd(m, arg) && aaa.Right(m, arg), func() { m.Cmdy(arg) })
}
func Command(m *ice.Message, arg ...string) {
	kit.If(!PodCmd(m, COMMAND, arg), func() { m.Cmdy(COMMAND, arg) })
}
func CmdHandler(args ...ice.Any) ice.Handler {
	return func(m *ice.Message, arg ...string) { m.Cmdy(args...) }
}
func CmdAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(args...), ice.RUN: {Hand: Run}, COMMAND: {Hand: Command}}
}
func CmdList(m *ice.Message) *ice.Message {
	return m.Cmdy(COMMAND, mdb.SEARCH, COMMAND, ice.OptionFields(INDEX))
}

func IsOrderCmd(key string) bool { return key[0] == '/' || key[0] == '_' }
func FileURI(dir string) string {
	if runtime.GOOS == "windows" {
		dir = strings.ReplaceAll(dir, "\\", "/")
	}
	if dir == "" {
		return ""
	} else if strings.Contains(dir, "/pkg/mod/") {
		dir = strings.Split(dir, "/pkg/mod/")[1]
	} else if ice.Info.Make.Path != "" && strings.HasPrefix(dir, ice.Info.Make.Path) {
		dir = strings.TrimPrefix(dir, ice.Info.Make.Path)
	} else if strings.HasPrefix(dir, kit.Path("")+nfs.PS) {
		dir = strings.TrimPrefix(dir, kit.Path("")+nfs.PS)
	}
	return path.Join(nfs.PS, ice.REQUIRE, dir)
}
func FileCmd(dir string) string {
	if strings.Index(dir, ":") == 1 {
		return FileURI(kit.ExtChange(strings.Join(kit.Slice(strings.Split(dir, ":"), 0, 2), ":"), nfs.GO))
	}
	return FileURI(kit.ExtChange(strings.Split(dir, nfs.DF)[0], nfs.GO))
}
func AddFileCmd(dir, key string) {
	ice.Info.File[FileCmd(dir)] = key
	if ls := strings.SplitN(path.Join(kit.Slice(kit.Split(FileCmd(dir), nfs.PS), 1, 4)...), mdb.AT, 2); len(ls) > 1 {
		ice.Info.Gomod[ls[0]] = ls[1]
	}
}
func GetFileCmd(dir string) string {
	if strings.HasPrefix(dir, ice.REQUIRE+nfs.PS) {
		dir = nfs.PS + dir
	} else if strings.HasPrefix(dir, ice.ISH_PLUGED) {
		dir = path.Join(nfs.PS, ice.REQUIRE, strings.TrimPrefix(dir, ice.ISH_PLUGED))
	}
	for _, dir := range []string{dir, path.Join(nfs.PS, ice.REQUIRE, ice.Info.Make.Module, dir), path.Join(nfs.PS, ice.REQUIRE, ice.Info.Make.Module, ice.SRC, dir)} {
		if cmd, ok := ice.Info.File[FileCmd(dir)]; ok {
			return cmd
		}
		p := path.Dir(dir)
		if cmd, ok := ice.Info.File[FileCmd(path.Join(p, path.Base(p)+nfs.PT+nfs.GO))]; ok {
			return cmd
		}
	}
	return ""
}
func GetCmdFile(m *ice.Message, cmds string) (file string) {
	m.Search(cmds, func(key string, cmd *ice.Command) {
		if file = strings.TrimPrefix(FileURI(kit.Split(cmd.FileLine(), nfs.DF)[0]), "/require/"); !nfs.Exists(m, file) {
			file = path.Join(ice.ISH_PLUGED, file)
		}
	})
	return
}
func TravelCmd(m *ice.Message, cb func(key, file, line string)) *ice.Message {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if IsOrderCmd(key) {
			return
		}
		if runtime.GOOS == "windows" {
			if ls := kit.Split(cmd.FileLine(), nfs.DF); len(ls) > 2 {
				cb(kit.Keys(s.Prefix(), key), strings.TrimPrefix(strings.Join(kit.Slice(ls, 0, -1), nfs.DF), kit.Path("")+nfs.PS), kit.Select("1", ls, -1))
				return
			}
		}
		if ls := kit.Split(cmd.FileLine(), nfs.DF); len(ls) > 0 && cmd.Name != "" {
			cb(kit.Keys(s.Prefix(), key), strings.TrimPrefix(ls[0], kit.Path("")+nfs.PS), kit.Select("1", ls, 1))
		}
	})
	return m
}
