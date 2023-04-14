package git

import (
	"os"
	"path"
	"strings"

	"shylinux.com/x/gogit"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _repos_path(name string) string {
	return kit.Select(path.Join(ice.USR, name)+ice.PS, nfs.PWD, name == path.Base(kit.Pwd()))
}
func _repos_cmd(m *ice.Message, name string, arg ...string) *ice.Message {
	return m.Copy(_git_cmd(m.Options(cli.CMD_DIR, _repos_path(name)), arg...))
}
func _repos_init(m *ice.Message, dir string) string {
	os.MkdirAll(path.Join(dir, REFS_HEADS), ice.MOD_DIR)
	os.MkdirAll(path.Join(dir, "objects/info/"), ice.MOD_DIR)
	return m.Cmdx(nfs.SAVE, path.Join(dir, "HEAD"), "ref: refs/heads/master")
}
func _repos_insert(m *ice.Message, name string, path string) bool {
	if repos, e := gogit.OpenRepository(_git_dir(path)); e == nil {
		origin := kit.Select("", kit.Split(repos.GetOrigin()), -1)
		kit.If(origin == "", func() { origin = _configs_read(m, _git_dir(path, "config"))["remote.origin.url"] })
		if ci, e := repos.GetCommit(); e == nil {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, ci.Author.When.Format(ice.MOD_TIME), COMMIT, strings.TrimSpace(ci.Message), BRANCH, repos.GetBranch(), ORIGIN, origin)
		} else {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, m.Time(), BRANCH, repos.GetBranch(), ORIGIN, origin)
		}
		return true
	}
	return false
}
func _repos_branch(m *ice.Message, dir string) {
	if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
		nfs.DirDeepAll(m, path.Join(dir, REFS_HEADS), "", func(value ice.Maps) {
			if refer, e := repos.LookupReference(REFS_HEADS + value[nfs.PATH]); !m.Warn(e, ice.ErrNotValid, value[nfs.PATH]) {
				if ci, e := repos.LookupCommit(refer.Oid); !m.Warn(e, ice.ErrNotValid, refer.Oid.String()) {
					m.Push(mdb.TIME, ci.Author.When.Format(ice.MOD_TIME))
					m.Push(BRANCH, value[nfs.PATH])
					m.Push(COMMIT, ci.Oid.Short())
					m.Push(AUTHOR, ci.Author.Name)
					m.Push(MESSAGE, ci.Message)
				}
			}
		}, nfs.PATH)
	}
}
func _repos_commit(m *ice.Message, dir, branch string, cb func(*gogit.Commit, *gogit.Repository) bool) {
	if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
		if refer, e := repos.LookupReference(REFS_HEADS + branch); !m.Warn(e, ice.ErrNotFound, branch) {
			if cb == nil {
				m.Push(mdb.TIME, m.Time())
				m.Push(COMMIT, cli.PWD)
				m.Push(AUTHOR, kit.Select(m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERNICK)))
				m.Push(MESSAGE, "opt some")
			}
			for oid := refer.Oid; oid != nil; {
				if ci, e := repos.LookupCommit(oid); !m.Warn(e, ice.ErrNotValid, oid.String()) {
					if cb == nil {
						m.Push(mdb.TIME, ci.Author.When.Format(ice.MOD_TIME))
						m.Push(COMMIT, ci.Oid.Short())
						m.Push(AUTHOR, ci.Author.Name)
						m.Push(MESSAGE, ci.Message)
					} else if cb(ci, repos) {
						break
					}
					if p := ci.ParentCommit(0); p != nil {
						oid = p.Oid
						continue
					}
				}
				break
			}
		}
	}
}
func _repos_dir(m *ice.Message, dir, branch, commit, file string, cb func(*gogit.TreeEntry, *gogit.Repository) bool) {
	if commit == cli.PWD && strings.HasPrefix(dir, ".git") {
		nfs.DirDeepAll(m, path.Dir(dir), file, nil, "time,line,path")
		m.Option(cli.CMD_DIR, path.Dir(dir))
		m.Echo(_git_cmds(m, DIFF))
		return
	} else if commit == cli.PWD {
		if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
			if refer, e := repos.LookupReference(REFS_HEADS + branch); !m.Warn(e, ice.ErrNotFound, branch) {
				commit = refer.Oid.String()
			}
		}
	} else if file == nfs.PWD {
		file = ""
	}
	if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
		if refer, e := repos.LookupReference(REFS_TAGS + commit); e == nil {
			commit = refer.Oid.String()
		}
	}
	_repos_commit(m, dir, branch, func(ci *gogit.Commit, repos *gogit.Repository) bool {
		if !strings.HasPrefix(ci.Oid.String(), commit) {
			return false
		}
		prev := ice.Maps{}
		if p := ci.ParentCommit(0); p != nil {
			if ci, e := repos.LookupCommit(p.Oid); e == nil {
				if tree, e := repos.LookupTree(ci.TreeId()); !m.Warn(e, ice.ErrNotValid, ci.TreeId().String) {
					tree.Walk(func(p string, v *gogit.TreeEntry) bool {
						prev[path.Join(p, v.Name)+kit.Select("", ice.PS, v.Type == gogit.ObjectTree)] = v.Oid.String()
						return false
					})
				}
			}
		}
		if tree, e := repos.LookupTree(ci.TreeId()); !m.Warn(e, ice.ErrNotValid, ci.TreeId().String) {
			m.Logs(mdb.SELECT, REPOS, dir, BRANCH, branch, COMMIT, commit, "tree", tree.Oid.Short())
			tree.Walk(func(p string, v *gogit.TreeEntry) bool {
				if pp := path.Join(p, v.Name) + kit.Select("", ice.PS, v.Type == gogit.ObjectTree); strings.HasPrefix(pp, file) {
					if cb == nil {
						if path.Dir(file) != path.Dir(pp) {

						} else if v.Type == gogit.ObjectTree {

						} else if id, ok := prev[pp]; ok && id == v.Oid.String() {
							if m.Option("_index") == web.CODE_INNER {
								m.Push(mdb.HASH, v.Oid.Short()).Push(nfs.PATH, pp).Push(mdb.STATUS, "")
							}
						} else if ok {
							m.Push(mdb.HASH, v.Oid.Short()).Push(nfs.PATH, pp).Push(mdb.STATUS, "~~~")
						} else {
							m.Push(mdb.HASH, v.Oid.Short()).Push(nfs.PATH, pp).Push(mdb.STATUS, "+++")
						}
						delete(prev, pp)
					} else if cb(v, repos) {
						return true
					}
				}
				return false
			})
		}
		kit.For(prev, func(pp, id string) {
			if path.Dir(file) != path.Dir(pp) {
				return
			}
			m.Push(mdb.HASH, id[:6]).Push(nfs.PATH, pp).Push(mdb.STATUS, "---")
		})
		if m.Sort(kit.Fields(mdb.STATUS, nfs.PATH), ice.STR_R, ice.STR); cb == nil {
			m.Option(cli.CMD_DIR, dir)
			m.Echo(_git_cmds(m, DIFF, ci.Oid.String()+"^", ci.Oid.String()))
			m.Status(mdb.TIME, ci.Author.When.Format(ice.MOD_TIME), DIFF, _git_cmds(m, DIFF, "--shortstat", ci.Oid.String()+"^", ci.Oid.String()), MESSAGE, ci.Message)
		}
		return true
	})
}
func _repos_cat(m *ice.Message, dir, branch, commit, file string) {
	if commit == cli.PWD && strings.HasPrefix(dir, ".git") {
		m.Cmdy(nfs.CAT, path.Join(path.Dir(dir), file))
		return
	} else if commit == cli.PWD {
		if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
			if refer, e := repos.LookupReference(REFS_HEADS + branch); !m.Warn(e, ice.ErrNotFound, branch) {
				commit = refer.Oid.String()
			}
		}
	}
	_repos_dir(m, dir, branch, commit, file, func(v *gogit.TreeEntry, repos *gogit.Repository) bool {
		if blob, e := repos.LookupBlob(v.Oid); e == nil {
			m.Logs(mdb.IMPORT, REPOS, dir, BRANCH, branch, COMMIT, commit, "blob", v.Oid.Short())
			m.Echo(string(blob.Contents()))
		} else {
			m.Option(cli.CMD_DIR, dir)
			m.Echo(_git_cmds(m, "cat-file", "-p", v.Oid.String()))
		}
		return true
	})
}

const (
	REFS_HEADS = "refs/heads/"
	REFS_TAGS  = "refs/tags/"

	INIT    = "init"
	CONFIG  = "config"
	ORIGIN  = "origin"
	BRANCH  = "branch"
	MASTER  = "master"
	AUTHOR  = "author"
	MESSAGE = "message"
)
const REPOS = "repos"

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 4 {
				m.RenderStatusBadRequest()
				return
			} else if path.Join(arg[:3]...) == ice.Info.Make.Module && nfs.Exists(m, path.Join(arg[3:]...)) {
				m.RenderDownload(path.Join(arg[3:]...))
				return
			}
			p := path.Join(kit.Select(ice.USR_REQUIRE, m.Cmdx(cli.SYSTEM, "go", "env", "GOMODCACHE")), path.Join(arg...))
			if !nfs.Exists(m, p) {
				if p = path.Join(ice.USR_REQUIRE, path.Join(arg...)); !nfs.Exists(m, p) {
					ls := strings.SplitN(path.Join(arg[:3]...), ice.AT, 2)
					if v := kit.Select(ice.Info.Gomod[ls[0]], ls, 1); v == "" {
						_git_cmd(m, "clone", "https://"+ls[0], path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...)))
					} else {
						_git_cmd(m, "clone", "-b", v, "https://"+ls[0], path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...)))
					}
				}
			}
			m.RenderDownload(p)
		}},
	})
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos@key branch@key commit@key path@key auto", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR, "name,path", func(value ice.Maps) { _repos_insert(m, value[mdb.NAME], value[nfs.PATH]) })
				_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())
				cli.IsSystem(m, GIT)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case REPOS:
					mdb.HashSelect(m).Cut(REPOS)
				case BRANCH:
					m.Cmdy("", m.Option(REPOS)).Cut(BRANCH)
				case COMMIT:
					m.Cmdy("", m.Option(REPOS), m.Option(BRANCH)).Cut("commit,author,message,time")
				case nfs.PATH:
					m.Cmdy("", m.Option(REPOS), m.Option(BRANCH), m.Option(COMMIT)).Cut("path,hash,status")
				default:
					mdb.HashInputs(m, arg)
				}
			}},
			mdb.CREATE: {Name: "create origin name path", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, strings.TrimSuffix(path.Base(m.Option(ORIGIN)), ".git"))
				m.OptionDefault(nfs.PATH, path.Join(ice.USR, m.Option(mdb.NAME)))
				if _repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH)) {
					return
				}
				m.Cmd("", INIT, kit.Dict(cli.CMD_DIR, m.Option(nfs.PATH)))
				_repos_insert(m, m.Option(mdb.NAME), m.Option(nfs.PATH))
			}},
			INIT: {Hand: func(m *ice.Message, arg ...string) {
				if dir := _repos_init(m, _git_dir(m.Option(cli.CMD_DIR))); m.Option(ORIGIN, kit.Select("", kit.Split(m.Option(ORIGIN)), -1)) != "" {
					m.Cmd(nfs.SAVE, path.Join(dir, CONFIG), kit.Format(nfs.TemplateText(m, CONFIG), m.Option(ORIGIN)))
					_git_cmd(m, PULL, ORIGIN, m.OptionDefault(BRANCH, MASTER))
				}
			}},
			"inner": {Help: "编辑器", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{_repos_path(arg[0]), kit.Select("README.md", arg, 3)}
				} else if kit.Select("", arg, 1) != ctx.ACTION {
					if dir := _git_dir(_repos_path(m.Option(REPOS))); len(arg) < 3 {
						_repos_dir(m, dir, m.Option(BRANCH), m.Option(COMMIT), kit.Select("", arg, 1), nil)
					} else {
						_repos_cat(m, dir, m.Option(BRANCH), m.Option(COMMIT), arg[2])
						// ctx.DisplayLocal(m, "code/inner.js")
					}
					return
				}
				ctx.ProcessField(m, "", arg, arg...)
			}},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit,origin"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				mdb.HashSelect(m, arg...).Action(mdb.CREATE)
			} else if dir := _git_dir(_repos_path(arg[0])); len(arg) == 1 || arg[1] == "" {
				_repos_branch(m, dir)
			} else if len(arg) == 2 || arg[2] == "" {
				_repos_commit(m, dir, arg[1], nil)
			} else if len(arg) == 3 || arg[3] == "" || strings.HasSuffix(arg[3], ice.PS) {
				_repos_dir(m, dir, arg[1], arg[2], kit.Select("", arg, 3), nil)
				return
			} else {
				m.Cmdy("", "inner", arg)
			}
			m.StatusTimeCount()
		}},
	})
}
func ReposList(m *ice.Message) *ice.Message { return m.Cmd(REPOS, ice.OptionFields("repos,path")) }
