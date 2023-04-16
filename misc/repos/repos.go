package repos

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	git "shylinux.com/x/go-git/v5"
	"shylinux.com/x/go-git/v5/plumbing"
	"shylinux.com/x/go-git/v5/plumbing/object"
	"shylinux.com/x/go-git/v5/plumbing/transport/file"
	"shylinux.com/x/go-git/v5/plumbing/transport/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _repos_path(m *ice.Message, p string, arg ...string) string {
	if p == path.Base(kit.Path("")) {
		return kit.Path("", arg...)
	}
	return path.Join(nfs.USR, p, path.Join(arg...))
}
func _repos_open(m *ice.Message, p string) *git.Repository {
	return mdb.HashSelectTarget(m, p, nil).(*git.Repository)
}
func _repos_insert(m *ice.Message, p string) {
	if repos, err := git.PlainOpen(p); err == nil {
		args := []string{REPOS, path.Base(p), nfs.PATH, p}
		if refer, err := repos.Head(); err == nil {
			args = append(args, BRANCH, refer.Name().String())
			if commit, err := repos.CommitObject(refer.Hash()); err == nil {
				args = append(args, mdb.TIME, commit.Author.When.Format(ice.MOD_TIME), COMMIT, commit.Message)
			}
		}
		if remote, err := repos.Remotes(); err == nil && len(remote) > 0 {
			args = append(args, ORIGIN, remote[0].Config().URLs[0])
		}
		mdb.HashCreate(m.Options(mdb.TARGET, repos), args)
	}
}
func _repos_each(m *ice.Message, title string, cb func(*git.Repository, ice.Maps) error) {
	msg := m.Cmd("")
	web.GoToast(m, title, func(toast func(string, int, int)) {
		list, count, total := []string{}, 0, msg.Length()
		msg.Table(func(value ice.Maps) {
			toast(value[REPOS], count, total)
			if err := cb(_repos_open(m, value[REPOS]), value); err != nil && err != git.NoErrAlreadyUpToDate {
				web.Toast(m, err.Error(), "error: "+value[REPOS], "", "3s")
				list = append(list, value[REPOS])
				m.Sleep3s()
			}
			count++
		})
		if len(list) > 0 {
			web.Toast(m, strings.Join(list, ice.NL), ice.FAILURE, "30s")
		} else {
			toast(ice.SUCCESS, count, total)
		}
	})

}
func _repos_log(m *ice.Message, repos *git.Repository) error {
	iter, err := repos.Log(&git.LogOptions{})
	if err != nil {
		return err
	}
	limit := 30
	defer m.StatusTimeCount()
	m.Push(mdb.TIME, m.Time())
	m.Push(COMMIT, INDEX)
	m.Push(aaa.USERNAME, m.Option(ice.MSG_USERNAME))
	m.Push(mdb.TEXT, "add some")
	m.Push("files", 0).Push("adds", 0).Push("dels", 0)
	return iter.ForEach(func(commit *object.Commit) error {
		if m.Length() > limit {
			return nil
		}
		m.Push(mdb.TIME, commit.Author.When)
		m.Push(COMMIT, commit.Hash.String())
		m.Push(aaa.USERNAME, commit.Author.Name)
		m.Push(mdb.TEXT, commit.Message)
		files, adds, dels := 0, 0, 0
		if stats, err := commit.Stats(); err == nil {
			for _, stat := range stats {
				files, adds, dels = files+1, adds+stat.Addition, dels+stat.Deletion
			}
		}
		m.Push("files", files).Push("adds", adds).Push("dels", adds)
		return nil
	})
}
func _repos_stats(m *ice.Message, repos *git.Repository, h string) error {
	commit, err := repos.CommitObject(plumbing.NewHash(h))
	if err != nil {
		return err
	}
	stats, err := commit.Stats()
	if err != nil {
		return err
	}
	defer m.StatusTimeCount()
	for _, stat := range stats {
		m.Push(nfs.FILE, stat.Name).Push("add", stat.Addition).Push("del", stat.Deletion)
	}
	return nil
}
func _repos_status(m *ice.Message, repos *git.Repository) error {
	work, err := repos.Worktree()
	if err != nil {
		return err
	}
	status, err := work.Status()
	if err != nil {
		return err
	}
	defer m.StatusTimeCount()
	for k, v := range status {
		switch kit.Ext(k) {
		case "swp", "swo":
			continue
		}
		m.Push(nfs.FILE, k).Push(STATUS, string(v.Worktree)+string(v.Staging))
		switch v.Worktree {
		case git.Untracked:
			m.PushButton(ADD, nfs.TRASH)
		case git.Modified:
			m.PushButton(ADD)
		default:
			m.PushButton(COMMIT)
		}
	}
	return nil
}

const (
	CLONE  = "clone"
	PULL   = "pull"
	PUSH   = "push"
	STATUS = "status"
	ADD    = "add"
	COMMIT = "commit"
	LOG    = "log"

	ORIGIN = "origin"
	BRANCH = "branch"
	INDEX  = "index"
)
const (
	REPOS = "repos"
)
const GIT = "git"

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Help: "代码库", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 4 {
				m.RenderStatusBadRequest()
			} else if path.Join(arg[:3]...) == ice.Info.Make.Module && nfs.Exists(m, path.Join(arg[3:]...)) {
				m.RenderDownload(path.Join(arg[3:]...))
			} else {
				p := path.Join(kit.Select(ice.USR_REQUIRE, m.Cmdx(cli.SYSTEM, "go", "env", "GOMODCACHE")), path.Join(arg...))
				if !nfs.Exists(m, p) {
					if p = path.Join(ice.USR_REQUIRE, path.Join(arg...)); !nfs.Exists(m, p) {
						ls := strings.SplitN(path.Join(arg[:3]...), ice.AT, 2)
						to := path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...))
						_, err := git.PlainClone(to, false, &git.CloneOptions{URL: "https://" + ls[0], ReferenceName: plumbing.NewBranchReferenceName(kit.Select(ice.Info.Gomod[ls[0]], ls, 1))})
						m.Warn(err)
					}
				}
				m.RenderDownload(p)
			}
		}},
	})
	code.Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos commit:text file:text auto", Help: "仓库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, nfs.USR, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
				_repos_insert(m, kit.Path(""))
			}},
			"upload-pack": {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeUploadPack(arg[0]); err != nil && err != io.EOF {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					os.Exit(128)
				}
			}},
			"receive-pack": {Hand: func(m *ice.Message, arg ...string) {
				if err := file.ServeReceivePack(arg[0]); err != nil && err != io.EOF {
					fmt.Fprintln(os.Stderr, "ERR:", err)
					fmt.Fprintln(os.Stderr, arg[0])
					os.Exit(128)
				}
			}},
			CLONE: {Name: "clone origin* branch name path", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, path.Base(m.Option(ORIGIN)))
				m.OptionDefault(nfs.PATH, path.Join(path.Join(nfs.USR, m.Option(mdb.NAME))))
				_, err := git.PlainClone(m.Option(nfs.PATH), false, &git.CloneOptions{URL: m.Option(ORIGIN)})
				m.Warn(err)
			}},
			PULL: {Hand: func(m *ice.Message, arg ...string) {
				_repos_each(m, "repos pull", func(repos *git.Repository, value ice.Maps) error {
					if value[ORIGIN] == "" {
						return nil
					} else if work, err := repos.Worktree(); err != nil {
						return err
					} else {
						return work.Pull(&git.PullOptions{})
					}
				})
			}},
			PUSH: {Hand: func(m *ice.Message, arg ...string) {
				list := map[string]*url.URL{}
				m.Cmd(nfs.CAT, kit.HomePath(".git-credentials"), func(line string) {
					u := kit.ParseURL(line)
					list[u.Host] = u
				})
				_repos_each(m, "repos push", func(repos *git.Repository, value ice.Maps) error {
					if value[ORIGIN] == "" {
						return nil
					}
					u := list[kit.ParseURL(value[ORIGIN]).Host]
					if password, ok := u.User.Password(); !ok {
						return errors.New("not found password")
					} else {
						return repos.Push(&git.PushOptions{Auth: &http.BasicAuth{Username: u.User.Username(), Password: password}})
					}
				})
			}},
			STATUS: {Hand: func(m *ice.Message, arg ...string) {
				_repos_each(m, "repos status", func(repos *git.Repository, value ice.Maps) error { return _repos_status(m, repos) })
			}},
			ADD: {Hand: func(m *ice.Message, arg ...string) {
				if work, err := _repos_open(m, m.Option(REPOS)).Worktree(); !m.Warn(err) {
					_, err := work.Add(m.Option(nfs.FILE))
					m.Warn(err)
				}
			}},
			COMMIT: {Name: "commit actions=add,opt,fix comment*=some", Hand: func(m *ice.Message, arg ...string) {
				if work, err := _repos_open(m, m.Option(REPOS)).Worktree(); !m.Warn(err) {
					_, err := work.Commit(m.Option("actions")+ice.SP+m.Option("comment"), &git.CommitOptions{})
					m.Warn(err)
				}
			}},
			LOG: {Hand: func(m *ice.Message, arg ...string) {
				_repos_log(m, _repos_open(m, kit.Select(m.Option(REPOS), arg, 0)))
			}},
			code.VIMER: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] != ice.RUN {
					arg = []string{path.Join(arg[:2]...), kit.Select("README.md", arg, 2)}
				} else if kit.Select("", arg, 1) != ctx.ACTION {
					ls := kit.Split(kit.Select(arg[1], m.Option(nfs.DIR_ROOT)), nfs.PS)
					if ls[1] == INDEX {
						if len(arg) < 3 {
							m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, _repos_path(m, ls[0])))
						} else {
							m.Cmdy(nfs.CAT, _repos_path(m, ls[0], arg[2]))
						}
					} else if commit, err := _repos_open(m, ls[0]).CommitObject(plumbing.NewHash(ls[1])); m.Warn(err) {
						return
					} else if len(arg) < 3 {
						if iter, err := commit.Files(); !m.Warn(err) {
							iter.ForEach(func(file *object.File) error {
								m.Push(nfs.PATH, file.Name)
								return nil
							})
						}
					} else {
						if file, err := commit.File(arg[2]); !m.Warn(err) {
							if content, err := file.Contents(); !m.Warn(err) {
								m.Echo(content)
							}
						}
					}
					ctx.DisplayLocal(m, "code/vimer.js")
					return
				}
				ctx.ProcessField(m, "", arg, arg...)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(!m.Warn(m.Option(REPOS) == ""), func() { m.Cmdy(nfs.TRASH, _repos_path(m, m.Option(REPOS), m.Option(nfs.FILE))) })
			}},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit,origin"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Action(CLONE, PULL, PUSH, STATUS)
			} else if len(arg) == 1 {
				_repos_log(m, _repos_open(m, arg[0]))
				// _repos_status(m, _repos_open(m, arg[0]))
			} else if len(arg) == 2 {
				if repos := _repos_open(m, arg[0]); arg[1] == INDEX {
					_repos_status(m, repos)
				} else {
					_repos_stats(m, repos, arg[1])
				}
			} else {
				m.Cmdy("", code.VIMER, arg)
			}
		}},
	})
}
