package ctx

import (
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _command_list(m *ice.Message, name string) *ice.Message {
	if strings.HasPrefix(name, "can.") {
		return m.Push(mdb.INDEX, name).Push(mdb.NAME, name).Push(mdb.HELP, "").Push(mdb.META, "").Push(mdb.LIST, "")
	}
	m.Option(ice.MSG_NODENAME, ice.Info.Titles)
	m.Option(ice.MSG_NODEICON, m.Resource(ice.Info.NodeIcon))
	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		icon := kit.Format(kit.Value(cmd.Meta, kit.Keys(ice.CTX_ICONS, key)))
		m.Push(mdb.INDEX, kit.Keys(s.Prefix(), key))
		m.Push(mdb.ICONS, kit.Select(cmd.Icon, icon, !kit.HasPrefix(icon, "bi ", "{")))
		m.Push(mdb.NAME, kit.Format(cmd.Name)).Push(mdb.HELP, kit.Format(cmd.Help))
		m.Push(mdb.LIST, kit.Format(cmd.List)).Push(mdb.META, kit.Format(cmd.Meta))
		m.Push("_command", ShortCmd(kit.Keys(s.Prefix(), key)))
		if !nfs.Exists(m, kit.Split(cmd.FileLine(), nfs.DF)[0], func(p string) {
			m.Push("_fileline", m.FileURI(p))
		}) {
			m.Push("_fileline", "")
		}
		m.Push("_role", kit.Select("", ice.OK, aaa.Right(m.Spawn(), name)))
		m.Push("_help", GetCmdHelp(m, name))
	})
	return m
}
func _command_search(m *ice.Message, kind, name, text string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if IsOrderCmd(key) || !strings.Contains(s.Prefix(key), name) {
			return
		}
		m.PushSearch(ice.CTX, kit.PathName(1), ice.CMD, kit.FileName(1),
			kit.SimpleKV("", s.Prefix(), kit.Select(key, cmd.Name), cmd.Help),
			INDEX, kit.Keys(s.Prefix(), key))
	}).Sort(m.OptionFields())
}

const (
	INDEX   = "index"
	CMDS    = "cmds"
	ARGS    = "args"
	OPTS    = "opts"
	STYLE   = "style"
	DISPLAY = "display"
	PREVIEW = "preview"
	ACTION  = "action"
	TOOLS   = "tools"
	RUN     = "run"
	SHIP    = "ship"

	ICONS = ice.CTX_ICONS
	TRANS = ice.CTX_TRANS
	TITLE = ice.CTX_TITLE
	INPUT = html.INPUT
)
const COMMAND = "command"

func init() {
	Index.MergeCommands(ice.Commands{
		COMMAND: {Name: "command key auto", Help: "命令", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				TravelCmd(m, func(key, file, line string) { AddFileCmd(file, key) })
				m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
					kit.If(cmd.Actions == nil, func() { cmd.Actions = ice.Actions{} })
					if _, ok := cmd.Actions[COMMAND]; !ok {
						cmd.Actions[COMMAND] = &ice.Action{Hand: Command}
					}
					if _, ok := cmd.Actions[RUN]; !ok {
						cmd.Actions[RUN] = &ice.Action{Hand: Run}
					}
					if _, ok := cmd.Actions[mdb.INPUTS]; !ok {
						cmd.Actions[mdb.INPUTS] = &ice.Action{Hand: func(m *ice.Message, arg ...string) { mdb.HashInputs(m, arg) }}
					}
				})
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || len(arg) > 1 && arg[1] != "" {
					_command_search(m, arg[0], kit.Select("", arg, 1), kit.Select("", arg, 2))
				}
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
			"default": {Hand: func(m *ice.Message, arg ...string) {
				m.Spawn(m.Source()).Search(arg[0], func(key string, cmd *ice.Command) {
					if arg[1] == ACTION {
						list := kit.Value(cmd.Meta, arg[2])
						kit.For(arg[3:], func(k, v string) {
							kit.For(list, func(value ice.Map) {
								kit.If(value[mdb.NAME] == k, func() {
									value[mdb.VALUE] = v
								})
							})
						})
						return
					}
					for i, v := range arg[1:] {
						kit.Value(cmd.List, kit.Keys(i, mdb.VALUE), v)
					}
				})
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.OptionFields(INDEX)
				m.Cmdy("", mdb.SEARCH, COMMAND)
			} else {
				kit.For(arg, func(k string) { _command_list(m, k) })
			}
		}},
	})
}

var PodCmd = func(m *ice.Message, arg ...ice.Any) bool { return false }

func Run(m *ice.Message, arg ...string) {
	kit.If(!PodCmd(m, arg) && aaa.Right(m, arg), func() { m.Cmdy(arg) })
}
func Command(m *ice.Message, arg ...string) {
	kit.If(!PodCmd(m, COMMAND, arg), func() { m.Cmdy(COMMAND, arg) })
}

func FileCmd(dir string) string {
	if strings.Index(dir, ":") == 1 {
		return ice.Pulse.FileURI(kit.ExtChange(strings.Join(kit.Slice(strings.Split(dir, ":"), 0, 2), ":"), nfs.GO))
	}
	return ice.Pulse.FileURI(kit.ExtChange(strings.Split(dir, nfs.DF)[0], nfs.GO))
}
func AddFileCmd(dir, key string) {
	if ls := strings.SplitN(path.Join(kit.Slice(kit.Split(FileCmd(dir), nfs.PS), 1, 4)...), mdb.AT, 2); len(ls) > 1 {
		_ls := strings.Split(FileCmd(dir), mdb.AT+ls[1]+nfs.PS)
		ice.Info.File[path.Join(nfs.P, nfs.USR, path.Base(_ls[0]), _ls[1])] = key
		ice.Info.Gomod[ls[0]] = ls[1]
	} else {
		ice.Info.File[FileCmd(dir)] = key
	}
}
func GetFileCmd(dir string) string {
	if strings.HasPrefix(dir, ice.REQUIRE+nfs.PS) {
		dir = nfs.PS + dir
	} else if strings.HasPrefix(dir, ice.ISH_PLUGED) {
		dir = path.Join(nfs.P, strings.TrimPrefix(dir, ice.ISH_PLUGED))
	}
	for _, dir := range []string{dir, path.Join(nfs.P, ice.Info.Make.Module, dir), path.Join(nfs.P, ice.Info.Make.Module, ice.SRC, dir)} {
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
func GetCmdHelp(m *ice.Message, cmds string) (file string) {
	file = kit.TrimPrefix(m.FileURI(kit.ExtChange(GetCmdFile(m, cmds), nfs.SHY)), nfs.P, nfs.REQUIRE)
	if !nfs.Exists(m, path.Join(nfs.USR_LEARNING_PORTAL, "commands", strings.TrimPrefix(file, nfs.USR_ICEBERGS)), func(p string) { file = p }) {
		kit.If(!nfs.Exists(m, file), func() { file = "" })
	}
	return
}
func GetCmdFile(m *ice.Message, cmds string) (file string) {
	m.Search(kit.Select(m.PrefixKey(), cmds), func(key string, cmd *ice.Command) {
		if file = kit.TrimPrefix(m.FileURI(kit.Split(cmd.FileLine(), nfs.DF)[0]), nfs.P); !nfs.Exists(m, file) {
			file = path.Join(nfs.P, file)
		}
	})
	return
}
func TravelCmd(m *ice.Message, cb func(key, file, line string)) *ice.Message {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if IsOrderCmd(key) {
			return
		}
		if runtime.GOOS == ice.WINDOWS {
			if ls := kit.Split(cmd.FileLine(), nfs.DF); len(ls) > 2 {
				cb(kit.Keys(s.Prefix(), key), strings.TrimPrefix(strings.Join(kit.Slice(ls, 0, -1), nfs.DF), kit.Path("")+nfs.PS), kit.Select("1", ls, -1))
				return
			}
		}
		if ls := kit.Split(cmd.FileLine(), nfs.DF); len(ls) > 0 {
			cb(kit.Keys(s.Prefix(), key), strings.TrimPrefix(ls[0], kit.Path("")+nfs.PS), kit.Select("1", ls, 1))
		}
	})
	return m
}
func IsOrderCmd(key string) bool {
	return key[0] == '/' || key[0] == '_'
}
func ShortCmd(key string) string {
	_key := kit.Select("", kit.Split(key, nfs.PT), -1)
	if _p, ok := ice.Info.Index[_key].(*ice.Context); ok && _p.Prefix(_key) == key {
		return _key
	}
	return key
}
func ResourceFile(m *ice.Message, file string, arg ...string) string {
	if kit.HasPrefix(file, nfs.PS, ice.HTTP) {
		return file
	} else if nfs.Exists(m, file) {
		return file
	} else {
		return path.Join(path.Dir(GetCmdFile(m, m.PrefixKey())), file)
	}
}
