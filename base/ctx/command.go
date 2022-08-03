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

func _command_list(m *ice.Message, name string) {
	if nfs.ExistsFile(m, path.Join(ice.SRC, name)) {
		switch kit.Ext(name) {
		case nfs.JS:
			m.Push(DISPLAY, FileURI(name))
			name = kit.Select(CAN_PLUGIN, GetFileCmd(name))

		case nfs.GO:
			name = GetFileCmd(name)

		default:
			if msg := m.Cmd(mdb.RENDER, kit.Ext(name)); msg.Length() > 0 {
				m.Push(ARGS, kit.Format(kit.List(name)))
				name = kit.Keys(msg.Append(mdb.TEXT), msg.Append(mdb.NAME))
			}
		}
	}
	if strings.HasPrefix(name, "can.") {
		m.Push(mdb.INDEX, name)
		return
	}
	if name == "" { // 命令列表
		for k, v := range m.Source().Commands {
			if k[0] == '/' || k[0] == '_' {
				continue // 内部命令
			}

			m.Push(mdb.KEY, k)
			m.Push(mdb.NAME, v.Name)
			m.Push(mdb.HELP, v.Help)
		}
		m.Sort(mdb.KEY)
		return
	}

	// 命令详情
	m.Spawn(m.Source()).Search(name, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Push(mdb.INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key))
		m.Push(mdb.NAME, kit.Format(cmd.Name))
		m.Push(mdb.HELP, kit.Format(cmd.Help))
		m.Push(mdb.META, kit.Format(cmd.Meta))
		m.Push(mdb.LIST, kit.Format(cmd.List))
	})
}
func _command_search(m *ice.Message, kind, name, text string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if key[0] == '/' || key[0] == '_' {
			return // 内部命令
		}
		if name != "" && !strings.HasPrefix(key, name) && !strings.Contains(s.Name, name) {
			return
		}

		m.PushSearch(ice.CTX, kit.PathName(1), ice.CMD, kit.FileName(1),
			kit.SimpleKV("", s.Cap(ice.CTX_FOLLOW), cmd.Name, cmd.Help),
			CONTEXT, s.Cap(ice.CTX_FOLLOW), COMMAND, key,
			INDEX, kit.Keys(s.Cap(ice.CTX_FOLLOW), key),
			mdb.HELP, cmd.Help,
		)
	})
}

const (
	INDEX   = "index"
	ARGS    = "args"
	STYLE   = "style"
	DISPLAY = "display"
	ACTION  = "action"

	CAN_PLUGIN = "can.plugin"
)
const COMMAND = "command"

func init() {
	Index.MergeCommands(ice.Commands{
		COMMAND: {Name: "command key auto", Help: "命令", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Prefix(COMMAND))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, COMMAND)
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == m.CommandKey() || len(arg) > 1 && arg[1] != "" {
					_command_search(m, arg[0], kit.Select("", arg, 1), kit.Select("", arg, 2))
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, "")
			}
			for _, key := range arg {
				_command_list(m, key)
			}
		}},
	})
}

func CmdAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(args...),
		COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
			if !PodCmd(m, COMMAND, arg) {
				m.Cmdy(COMMAND, arg)
			}
		}},
		ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 3 && arg[1] == ACTION && arg[2] == CONFIG && arg[3] == "reset" {
				m.Cmd(CONFIG, "reset", arg[0])
				return
			}
			if aaa.Right(m, arg) && !PodCmd(m, arg) {
				m.Cmdy(arg)
			}
		}},
	}
}
func PodCmd(m *ice.Message, arg ...ice.Any) bool {
	if pod := m.Option(ice.POD); pod != "" {
		if m.Option(ice.POD, ""); m.Option(ice.MSG_UPLOAD) != "" {
			msg := m.Cmd("cache", "upload")
			m.Option(ice.MSG_UPLOAD, msg.Append(mdb.HASH), msg.Append(mdb.NAME), msg.Append(nfs.SIZE))
		}
		m.Cmdy(append(kit.List("space", pod), arg...))
		return true
	}
	return false
}

func FileURI(dir string) string {
	if strings.Contains(dir, "go/pkg/mod") {
		return path.Join("/require", strings.Split(dir, "go/pkg/mod")[1])
	}
	if ice.Info.Make.Path != "" && strings.HasPrefix(dir, ice.Info.Make.Path+ice.PS) {
		dir = strings.TrimPrefix(dir, ice.Info.Make.Path+ice.PS)
	}
	if strings.HasPrefix(dir, kit.Path("")+ice.PS) {
		dir = strings.TrimPrefix(dir, kit.Path("")+ice.PS)
	}
	if strings.HasPrefix(dir, ice.USR) {
		return path.Join("/require", dir)
	}
	if strings.HasPrefix(dir, ice.SRC) {
		return path.Join("/require", dir)
	}
	if kit.FileExists(path.Join(ice.SRC, dir)) {
		return path.Join("/require/src/", dir)
	}
	return dir
}
func FileCmd(dir string) string {
	dir = strings.Split(dir, ice.DF)[0]
	dir = strings.ReplaceAll(dir, ".js", ".go")
	dir = strings.ReplaceAll(dir, ".sh", ".go")
	return FileURI(dir)
}
func AddFileCmd(dir, key string) {
	ice.Info.File[FileCmd(dir)] = key
}
func GetFileCmd(dir string) string {
	if strings.HasPrefix(dir, "require/") {
		dir = "/" + dir
	}
	for _, dir := range []string{dir, path.Join("/require/", ice.Info.Make.Module, dir), path.Join("/require/", ice.Info.Make.Module, ice.SRC, dir)} {
		if cmd, ok := ice.Info.File[FileCmd(dir)]; ok {
			return cmd
		}
		p := path.Dir(dir)
		if cmd, ok := ice.Info.File[FileCmd(path.Join(p, path.Base(p)+".go"))]; ok {
			return cmd
		}
		for k, v := range ice.Info.File {
			if strings.HasPrefix(k, p) {
				return v
			}
		}
	}
	return ""
}
