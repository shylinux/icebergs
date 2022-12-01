package ctx

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
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
	if nfs.ExistsFile(m, path.Join(ice.SRC, name)) {
		switch kit.Ext(name) {
		case nfs.GO:
			name = GetFileCmd(name)
		case nfs.JS:
			m.Push(DISPLAY, FileURI(name))
			name = kit.Select(CAN_PLUGIN, GetFileCmd(name))
		default:
			if msg := m.Cmd(mdb.RENDER, kit.Ext(name)); msg.Length() > 0 {
				m.Push(ARGS, kit.Format(kit.List(name)))
				name = kit.Keys(msg.Append(mdb.TEXT), msg.Append(mdb.NAME))
			}
		}
	}
	if strings.HasPrefix(name, "can.") {
		return m.Push(mdb.INDEX, name)
	}
	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Push(mdb.INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key))
		m.Push(mdb.NAME, kit.Format(cmd.Name))
		m.Push(mdb.HELP, kit.Format(cmd.Help))
		m.Push(mdb.META, kit.Format(cmd.Meta))
		m.Push(mdb.LIST, kit.Format(cmd.List))
	})
	return m
}
func _command_search(m *ice.Message, kind, name, text string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if IsOrderCmd(key) {
			return
		}
		if name != "" && !strings.HasPrefix(key, name) && !strings.Contains(s.Name, name) {
			return
		}
		m.PushSearch(ice.CTX, kit.PathName(1), ice.CMD, kit.FileName(1), kit.SimpleKV("", s.Cap(ice.CTX_FOLLOW), cmd.Name, cmd.Help),
			CONTEXT, s.Cap(ice.CTX_FOLLOW), COMMAND, key, INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key),
			mdb.HELP, cmd.Help, nfs.FILE, FileURI(cmd.GetFileLines()),
		)
	})
}

const (
	INDEX   = "index"
	ARGS    = "args"
	STYLE   = "style"
	DISPLAY = "display"
	ACTION  = "action"
	TRANS   = "trans"

	CAN_PLUGIN = "can.plugin"
)
const COMMAND = "command"

func init() {
	Index.MergeCommands(ice.Commands{
		COMMAND: {Name: "command key auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				TravelCmd(m, func(key, file, line string) {
					if strings.Contains(file, ice.ICEBERGS) {
						AddFileCmd(file, key)
					}
				})
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || len(arg) > 1 && arg[1] != "" {
					_command_search(m, arg[0], kit.Select("", arg, 1), kit.Select("", arg, 2))
				}
			}},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				TravelCmd(m, func(key, file, line string) {
					m.Push(mdb.NAME, key).Push(nfs.FILE, file).Push(nfs.LINE, line)
				}).Sort(mdb.NAME).Tables(func(value ice.Maps) {
					m.Echo(`%s	%s	%s;" f`+ice.NL, value[mdb.NAME], value[nfs.FILE], value[nfs.LINE])
				}).Cmd(nfs.SAVE, "tags", m.Result())
			}},
		}, aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "")
			}
			for _, key := range arg {
				_command_list(m, key)
			}
		}},
	})
}

var runChecker = []func(*ice.Message, string, string, ...string) bool{}

func AddRunChecker(cb func(*ice.Message, string, string, ...string) bool) {
	runChecker = append(runChecker, cb)
}
func Run(m *ice.Message, arg ...string) {
	if len(arg) > 3 && arg[1] == ACTION && arg[2] == CONFIG {
		for _, check := range runChecker {
			if check(m, arg[0], arg[3], arg...) {
				return
			}
		}
	}
	if Upload(m); !PodCmd(m, arg) && aaa.Right(m, arg) {
		m.Cmdy(arg)
	}
}

var Upload = func(*ice.Message) {}

func PodCmd(m *ice.Message, arg ...ice.Any) bool {
	if pod := m.Option(ice.POD); pod != "" {
		m.Option(ice.POD, "")
		m.Cmdy(append(kit.List(ice.SPACE, pod), arg...)...)
		return true
	}
	return false
}
func CmdHandler(args ...ice.Any) ice.Handler {
	return func(m *ice.Message, arg ...string) { m.Cmdy(args...) }
}
func CmdAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(args...),
		COMMAND: {Hand: func(m *ice.Message, arg ...string) {
			if !PodCmd(m, COMMAND, arg) {
				m.Cmdy(COMMAND, arg)
			}
		}}, ice.RUN: {Hand: Run},
	}
}

func FileURI(dir string) string {
	if dir == "" {
		return ""
	} else if strings.Contains(dir, "/go/pkg/mod/") {
		return path.Join(ice.PS, ice.REQUIRE, strings.Split(dir, "/go/pkg/mod/")[1])
	}
	if path.IsAbs(dir) {
		if strings.HasPrefix(dir, kit.Path("")+ice.PS) {
			dir = strings.TrimPrefix(dir, kit.Path("")+ice.PS)
		} else if ice.Info.Make.Path != "" && strings.HasPrefix(dir, ice.Info.Make.Path+ice.PS) {
			dir = strings.TrimPrefix(dir, ice.Info.Make.Path+ice.PS)
		}
	} else if nfs.ExistsFile(ice.Pulse, path.Join(ice.SRC, dir)) {
		dir = path.Join(ice.SRC, dir)
	}
	return path.Join(ice.PS, ice.REQUIRE, dir)
}
func FileCmd(dir string) string  { return FileURI(kit.ExtChange(strings.Split(dir, ice.DF)[0], nfs.GO)) }
func AddFileCmd(dir, key string) { ice.Info.File[FileCmd(dir)] = key }
func IsOrderCmd(key string) bool { return key[0] == '/' || key[0] == '_' }
func GetFileCmd(dir string) string {
	if strings.HasPrefix(dir, ice.REQUIRE+ice.PS) {
		dir = ice.PS + dir
	} else if strings.HasPrefix(dir, ice.ISH_PLUGED) {
		dir = path.Join(ice.PS, ice.REQUIRE, strings.TrimPrefix(dir, ice.ISH_PLUGED))
	}
	for _, dir := range []string{dir, path.Join(ice.PS, ice.REQUIRE, ice.Info.Make.Module, dir), path.Join(ice.PS, ice.REQUIRE, ice.Info.Make.Module, ice.SRC, dir)} {
		if cmd, ok := ice.Info.File[FileCmd(dir)]; ok {
			return cmd
		}
		p := path.Dir(dir)
		if cmd, ok := ice.Info.File[FileCmd(path.Join(p, path.Base(p)+ice.PT+nfs.GO))]; ok {
			return cmd
		}
	}
	return ""
}
func GetCmdFile(m *ice.Message, cmds string) (file string) {
	m.Search(cmds, func(key string, cmd *ice.Command) {
		if file = strings.TrimPrefix(FileURI(kit.Split(cmd.GetFileLines(), ice.DF)[0]), "/require/"); !nfs.ExistsFile(m, file) {
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
		if ls := kit.Split(cmd.GetFileLines(), ice.DF); len(ls) > 0 {
			cb(kit.Keys(s.Cap(ice.CTX_FOLLOW), key), strings.TrimPrefix(ls[0], kit.Path("")+ice.PS), kit.Select("1", ls, 1))
		}
	})
	return m
}
