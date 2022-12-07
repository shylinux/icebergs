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
	kit "shylinux.com/x/toolkits"
)

func _repos_path(name string) string {
	if strings.Contains(name, ":\\") {
		return name
	}
	return kit.Select(path.Join(ice.USR, name)+ice.PS, nfs.PWD, name == path.Base(kit.Pwd()))
}
func _repos_cmd(m *ice.Message, name string, arg ...string) *ice.Message {
	m.Option(cli.CMD_DIR, _repos_path(name))
	return m.Copy(_git_cmd(m, arg...))
}
func _repos_init(m *ice.Message, p string) string {
	os.MkdirAll(path.Join(p, "refs/heads/"), ice.MOD_DIR)
	os.MkdirAll(path.Join(p, "objects/info/"), ice.MOD_DIR)
	return m.Cmdx(nfs.SAVE, path.Join(p, "HEAD"), "ref: refs/heads/master")
}
func _repos_insert(m *ice.Message, name string, path string) bool {
	if repos, e := gogit.OpenRepository(_git_dir(path)); e == nil {
		if ci, e := repos.GetCommit(); e == nil {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, ci.Author.When.Format(ice.MOD_TIME), COMMIT, strings.TrimSpace(ci.Message),
				BRANCH, repos.GetBranch(), ORIGIN, kit.Select("", kit.Slice(kit.Split(repos.GetOrigin()), -1), 0))
		} else {
			mdb.HashCreate(m, REPOS, name, nfs.PATH, path, mdb.TIME, m.Time(),
				BRANCH, repos.GetBranch(), ORIGIN, kit.Select("", kit.Slice(kit.Split(repos.GetOrigin()), -1), 0))
		}
		return true
	}
	return false
}
func _repos_branch(m *ice.Message, dir string) {
	if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
		nfs.DirDeepAll(m, path.Join(dir, "refs/heads/"), "", func(value ice.Maps) {
			if refer, e := repos.LookupReference("refs/heads/" + value[nfs.PATH]); !m.Warn(e, ice.ErrNotValid, value[nfs.PATH]) {
				if ci, e := repos.LookupCommit(refer.Oid); !m.Warn(e, ice.ErrNotValid, refer.Oid.String()) {
					m.Push(mdb.TIME, ci.Author.When.Format(ice.MOD_TIME))
					m.Push(BRANCH, value[nfs.PATH])
					m.Push(COMMIT, ci.Oid.String()[:6])
					m.Push(AUTHOR, ci.Author.Name)
					m.Push(mdb.TEXT, ci.Message)
				}
			}
		}, nfs.PATH)
	}
}
func _repos_commit(m *ice.Message, dir, branch string, cb func(*gogit.Commit, *gogit.Repository) bool) {
	if repos, e := gogit.OpenRepository(dir); !m.Warn(e, ice.ErrNotFound, dir) {
		if refer, e := repos.LookupReference("refs/heads/" + branch); !m.Warn(e, ice.ErrNotFound, branch) {
			if cb == nil {
				m.Push(mdb.TIME, m.Time())
				m.Push(COMMIT, cli.PWD)
				m.Push(AUTHOR, kit.Select(m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERNICK)))
				m.Push(mdb.TEXT, "opt some")
			}
			for oid := refer.Oid; oid != nil; {
				if ci, e := repos.LookupCommit(oid); !m.Warn(e, ice.ErrNotValid, oid.String()) {
					if cb == nil {
						m.Push(mdb.TIME, ci.Author.When.Format(ice.MOD_TIME))
						m.Push(COMMIT, ci.Oid.String()[:6])
						m.Push(AUTHOR, ci.Author.Name)
						m.Push(mdb.TEXT, ci.Message)
					} else if cb(ci, repos) {
						break
					}
					if p := ci.Parent(0); p != nil {
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
	if commit == cli.PWD {
		nfs.DirDeepAll(m, path.Dir(dir), file, nil, "time,line,path")
		return
	} else if file == nfs.PWD {
		file = ""
	}
	_repos_commit(m, dir, branch, func(ci *gogit.Commit, repos *gogit.Repository) bool {
		if strings.HasPrefix(ci.Oid.String(), commit) {
			if tree, e := repos.LookupTree(ci.TreeId()); !m.Warn(e, ice.ErrNotValid, ci.TreeId().String) {
				m.Logs(mdb.SELECT, REPOS, dir, BRANCH, branch, COMMIT, commit, "tree", tree.Oid.String()[:6])
				tree.Walk(func(p string, v *gogit.TreeEntry) int {
					if strings.HasPrefix(path.Join(p, v.Name), file) {
						if cb == nil {
							m.Push(mdb.HASH, v.Id.String()[:6])
							m.Push(nfs.PATH, path.Join(p, v.Name)+kit.Select("", ice.PS, v.Type == gogit.ObjectTree))
						} else if cb(v, repos) {
							return -1
						}
					}
					return 0
				})
			}
			return true
		}
		return false
	})
}
func _repos_cat(m *ice.Message, dir, branch, commit, file string) {
	if commit == cli.PWD {
		m.Cmdy(nfs.CAT, path.Join(path.Dir(dir), file))
		return
	}
	_repos_dir(m, dir, branch, commit, file, func(v *gogit.TreeEntry, repos *gogit.Repository) bool {
		if blob, e := repos.LookupBlob(v.Id); e == nil {
			m.Logs(mdb.IMPORT, REPOS, dir, BRANCH, branch, COMMIT, commit, "blob", v.Id.String()[:6])
			m.Echo(string(blob.Contents()))
		} else {
			m.Option(cli.CMD_DIR, dir)
			m.Echo(_git_cmds(m, "cat-file", "-p", v.Id.String()))
		}
		return true
	})
}

const (
	ORIGIN = "origin"
	BRANCH = "branch"
	MASTER = "master"
	AUTHOR = "author"
	INIT   = "init"
)
const REPOS = "repos"

func init() {
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos branch commit path auto create inner", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR, "name,path", func(value ice.Maps) { _repos_insert(m, value[mdb.NAME], value[nfs.PATH]) })
				_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())
				cli.IsSystem(m, GIT)
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
					m.Cmd(nfs.SAVE, path.Join(dir, "config"), kit.Format(_repos_config, m.Option(ORIGIN)))
					_git_cmd(m, PULL, ORIGIN, m.OptionDefault(BRANCH, MASTER))
				}
			}},
			"inner": {Help: "编辑器", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{_repos_path(arg[0]), kit.Select("README.md", arg, 3)}
				} else if kit.Select("", arg, 1) != ctx.ACTION {
					if ctx.DisplayLocal(m, "code/inner.js"); len(arg) < 3 {
						_repos_dir(m, _git_dir(_repos_path(m.Option(REPOS))), m.Option(BRANCH), m.Option(COMMIT), kit.Select("", arg, 1), nil)
					} else {
						_repos_cat(m, _git_dir(_repos_path(m.Option(REPOS))), m.Option(BRANCH), m.Option(COMMIT), arg[2])
					}
					return
				}
				ctx.ProcessField(m, "web.code.inner", arg, arg...)
			}},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit,origin"), mdb.ClearHashOnExitAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				mdb.HashSelect(m, arg...)
			} else if len(arg) == 1 || arg[1] == "" {
				_repos_branch(m, _git_dir(_repos_path(arg[0])))
			} else if len(arg) == 2 || arg[2] == "" {
				_repos_commit(m, _git_dir(_repos_path(arg[0])), arg[1], nil)
			} else if len(arg) == 3 || arg[3] == "" || strings.HasSuffix(arg[3], ice.PS) {
				_repos_dir(m, _git_dir(_repos_path(arg[0])), arg[1], arg[2], kit.Select("", arg, 3), nil)
			} else {
				m.Cmdy("", "inner", arg)
			}
			m.StatusTimeCount()
		}},
	})
}
func ReposList(m *ice.Message) *ice.Message { return m.Cmd(REPOS, ice.OptionFields("repos,path")) }

var _repos_config = `
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "master"]
	remote = origin
	merge = refs/heads/master
`
