package git

import (
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	git "shylinux.com/x/go-git/v5"
	"shylinux.com/x/go-git/v5/config"
	"shylinux.com/x/go-git/v5/plumbing"
	"shylinux.com/x/go-git/v5/plumbing/object"
	"shylinux.com/x/go-git/v5/plumbing/transport/http"
	"shylinux.com/x/go-git/v5/utils/diffmatchpatch"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _repos_cmd(m *ice.Message, p string, arg ...string) *ice.Message {
	return m.Copy(_git_cmd(m.Options(cli.CMD_DIR, _repos_path(m, p)), arg...))
}
func _repos_init(m *ice.Message, p string) { git.PlainInit(p, true) }
func _repos_insert(m *ice.Message, p string) {
	if repos, err := git.PlainOpen(p); err == nil {
		args := []string{REPOS, path.Base(p), nfs.PATH, p}
		if refer, err := repos.Head(); err == nil {
			args = append(args, BRANCH, strings.TrimPrefix(refer.Name().String(), "refs/heads/"))
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
func _repos_path(m *ice.Message, p string, arg ...string) string {
	if p == path.Base(kit.Path("")) {
		return kit.Path("", arg...)
	}
	return path.Join(nfs.USR, p, path.Join(arg...))
}
func _repos_open(m *ice.Message, p string) *git.Repository {
	return mdb.HashSelectTarget(m, p, nil).(*git.Repository)
}
func _repos_each(m *ice.Message, title string, cb func(*git.Repository, ice.Maps) error) {
	msg := m.Cmd("")
	web.GoToast(m, kit.Select(m.CommandKey()+ice.SP+m.ActionKey(), title), func(toast func(string, int, int)) {
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
func _repos_branch(m *ice.Message, repos *git.Repository) error {
	iter, err := repos.Branches()
	if err != nil {
		return err
	}
	defer m.StatusTimeCount()
	iter.ForEach(func(refer *plumbing.Reference) error {
		if commit, err := repos.CommitObject(refer.Hash()); err == nil {
			m.Push(mdb.TIME, commit.Author.When.Format(ice.MOD_TIME))
			m.Push(BRANCH, strings.TrimPrefix(refer.Name().String(), "refs/heads/"))
			m.Push(aaa.USERNAME, commit.Author.Name)
			m.Push(mdb.TEXT, commit.Message)
		}
		return nil
	})
	return nil
}
func _repos_log(m *ice.Message, branch *config.Branch, repos *git.Repository) error {
	refer, err := repos.Reference(branch.Merge, true)
	if err != nil {
		return err
	}
	iter, err := repos.Log(&git.LogOptions{From: refer.Hash()})
	if err != nil {
		return err
	}
	limit := 30
	defer m.StatusTimeCount()
	m.Push(mdb.TIME, m.Time()).Push(COMMIT, INDEX)
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
	adds, dels := 0, 0
	for _, stat := range stats {
		m.Push(nfs.FILE, stat.Name).Push("add", stat.Addition).Push("del", stat.Deletion)
		adds += stat.Addition
		dels += stat.Deletion
	}
	m.StatusTimeCount("adds", adds, "dels", dels)
	return nil
}
func _repos_status(m *ice.Message, p string, repos *git.Repository) error {
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
		if kit.IsIn(kit.Ext(k), "swp", "swo") {
			continue
		}
		switch m.Push(REPOS, p).Push(STATUS, string(v.Worktree)+string(v.Staging)).Push(nfs.FILE, k); v.Worktree {
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
func _repos_total(m *ice.Message, p string, repos *git.Repository, stats map[string]int) *time.Time {
	iter, err := repos.Log(&git.LogOptions{})
	if err != nil {
		return nil
	}
	from, cmts, adds, dels := time.Now(), 0, 0, 0
	iter.ForEach(func(commit *object.Commit) error {
		from, cmts = commit.Author.When, cmts+1
		if stats, err := commit.Stats(); err == nil {
			for _, stat := range stats {
				adds, dels = adds+stat.Addition, dels+stat.Deletion
			}
		}
		return nil
	})
	days := kit.Int(time.Now().Sub(from) / time.Hour / 24)
	m.Push(REPOS, p).Push("from", from.Format(ice.MOD_TIME)).Push("days", days)
	m.Push("commits", cmts).Push("adds", adds).Push("dels", dels).Push("rest", adds-dels)
	stats["cmts"] += cmts
	stats["adds"] += adds
	stats["dels"] += dels
	stats["rest"] += adds - dels
	if days > stats["days"] {
		stats["days"] = days
		return &from
	}
	return nil
}
func _repos_vimer(m *ice.Message, _repos_path func(m *ice.Message, p string, arg ...string) string, arg ...string) {
	if len(arg) == 0 || arg[0] != ice.RUN {
		arg = []string{path.Join(arg[:3]...), kit.Select("README.md", arg, 3)}
	} else if kit.Select("", arg, 1) != ctx.ACTION {
		if ls := kit.Split(path.Join(m.Option(nfs.DIR_ROOT), arg[1]), nfs.PS); len(ls) < 2 || ls[2] == INDEX {
			if repos := _repos_open(m, ls[0]); len(arg) < 3 {
				if work, err := repos.Worktree(); err == nil {
					if status, err := work.Status(); err == nil {
						for k := range status {
							m.Echo(k)
						}
					}
				}
				m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, _repos_path(m, ls[0])))
			} else {
				m.Cmdy(nfs.CAT, _repos_path(m, ls[0], arg[2]))
				if refer, err := repos.Head(); err == nil {
					if commit, err := repos.CommitObject(refer.Hash()); err == nil {
						if file, err := commit.File(arg[2]); !m.Warn(err) {
							if content, err := file.Contents(); !m.Warn(err) {
								for _, diff := range diffmatchpatch.New().DiffMain(content, m.Result(), true) {
									switch diff.Type {
									case diffmatchpatch.DiffDelete:
										m.Push(mdb.TYPE, mdb.DELETE)
									case diffmatchpatch.DiffInsert:
										m.Push(mdb.TYPE, mdb.INSERT)
									default:
										m.Push(mdb.TYPE, "")
									}
									m.Push(mdb.TEXT, diff.Text)
								}
							}
						}
					}
				}
			}
		} else if commit, err := _repos_open(m, ls[0]).CommitObject(plumbing.NewHash(ls[2])); m.Warn(err) {
			return
		} else if len(arg) < 3 {
			if iter, err := commit.Files(); !m.Warn(err) {
				iter.ForEach(func(file *object.File) error {
					m.Push(nfs.PATH, file.Name)
					return nil
				})
			}
			if stats, err := commit.Stats(); err == nil {
				for _, stat := range stats {
					m.Echo(stat.Name)
				}
			}
		} else {
			if file, err := commit.File(arg[2]); !m.Warn(err) {
				if content, err := file.Contents(); !m.Warn(err) {
					if parent, err := commit.Parent(0); err == nil {
						if file0, err := parent.File(arg[2]); err == nil {
							if content0, err := file0.Contents(); err == nil {
								for _, diff := range diffmatchpatch.New().DiffMain(content0, content, true) {
									switch diff.Type {
									case diffmatchpatch.DiffDelete:
										m.Push(mdb.TYPE, mdb.DELETE)
									case diffmatchpatch.DiffInsert:
										m.Push(mdb.TYPE, mdb.INSERT)
									default:
										m.Push(mdb.TYPE, "")
									}
									m.Push(mdb.TEXT, diff.Text)
								}
							}
						}
					}
					m.Echo(content)
				}
			}
		}
		ctx.DisplayLocal(m, "code/vimer.js")
		return
	}
	ctx.ProcessField(m, "", arg, arg...)
}

const (
	CLONE  = "clone"
	PULL   = "pull"
	PUSH   = "push"
	LOG    = "log"
	TAG    = "tag"
	ADD    = "add"
	STASH  = "stash"
	COMMIT = "commit"

	BRANCH = "branch"

	ORIGIN = "origin"
	MASTER = "master"
	INDEX  = "index"
)
const REPOS = "repos"

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Hand: func(m *ice.Message, arg ...string) {
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
	Index.MergeCommands(ice.Commands{
		REPOS: {Name: "repos repos branch:text commit:text file:text auto", Help: "代码库", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, nfs.USR, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
				_repos_insert(m, kit.Path(""))
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.PushSearch(mdb.TYPE, web.LINK, mdb.NAME, m.CommandKey(), mdb.TEXT, m.MergePodCmd("", "", log.DEBUG, ice.TRUE))
				}
			}},
			CLONE: {Name: "clone origin* branch name path", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, path.Base(m.Option(ORIGIN)))
				m.OptionDefault(nfs.PATH, path.Join(path.Join(nfs.USR, m.Option(mdb.NAME))))
				if _, err := git.PlainClone(m.Option(nfs.PATH), false, &git.CloneOptions{URL: m.Option(ORIGIN)}); m.Warn(err) {
					_repos_insert(m, m.Option(nfs.PATH))
				}
			}},
			PULL: {Hand: func(m *ice.Message, arg ...string) {
				_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
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
				_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
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
			LOG: {Hand: func(m *ice.Message, arg ...string) {
				repos := _repos_open(m, kit.Select(m.Option(REPOS), arg, 0))
				if branch, err := repos.Branch(kit.Select(m.Option(BRANCH), arg, 1)); !m.Warn(err) {
					_repos_log(m, branch, repos)
				}
			}},
			TAG: {Name: "tag version", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(VERSION) == "", func() { m.Option(VERSION, _status_tag(m, m.Option(TAGS))) })
				repos := _repos_open(m, m.Option(REPOS))
				if refer, err := repos.Head(); !m.Warn(err) {
					_, err := repos.CreateTag(m.Option(VERSION), refer.Hash(), &git.CreateTagOptions{})
					m.Warn(err)
				}
			}},
			ADD: {Hand: func(m *ice.Message, arg ...string) {
				if work, err := _repos_open(m, m.Option(REPOS)).Worktree(); !m.Warn(err) {
					_, err := work.Add(m.Option(nfs.FILE))
					m.Warn(err)
				}
			}},
			STASH: {Hand: func(m *ice.Message, arg ...string) { _repos_cmd(m, kit.Select(m.Option(REPOS), arg, 0), STASH) }},
			COMMIT: {Name: "commit actions=add,opt,fix comment*=some", Hand: func(m *ice.Message, arg ...string) {
				if work, err := _repos_open(m, m.Option(REPOS)).Worktree(); !m.Warn(err) {
					_, err := work.Commit(m.Option("actions")+ice.SP+m.Option("comment"), &git.CommitOptions{All: true})
					m.Warn(err)
				}
			}},
			STATUS: {Hand: func(m *ice.Message, arg ...string) {
				if repos := kit.Select(m.Option(REPOS), arg, 0); repos != "" {
					_repos_status(m, repos, _repos_open(m, repos))
				} else {
					_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
						return _repos_status(m, value[REPOS], repos)
					})
				}
			}},
			TOTAL: {Hand: func(m *ice.Message, arg ...string) {
				stats := map[string]int{}
				if repos := kit.Select(m.Option(REPOS), arg, 0); repos == "" {
					var from *time.Time
					_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
						if kit.IsIn(value[REPOS], "go-git", "go-qrcode", "websocket") {
							return nil
						}
						t := _repos_total(m, value[REPOS], repos, stats)
						kit.If(t != nil, func() { from = t })
						return nil
					})
					m.Push(REPOS, TOTAL)
					m.Push("from", from.Format(ice.MOD_TIME))
					m.Push("days", stats["days"])
					m.Push("commits", stats["cmts"])
					m.Push("adds", stats["adds"])
					m.Push("dels", stats["dels"])
					m.Push("rest", stats["rest"])
					m.SortIntR("rest")
				} else {
					_repos_total(m, repos, _repos_open(m, repos), stats)
				}
				m.StatusTimeCount()
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				if !m.Warn(m.Option(REPOS), ice.ErrNotValid, REPOS) && !m.Warn(m.Option(nfs.FILE), ice.ErrNotValid, nfs.FILE) {
					nfs.Trash(m, _repos_path(m, m.Option(REPOS), m.Option(nfs.FILE)))
				}
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				if !m.Warn(m.Option(REPOS), ice.ErrNotValid, REPOS) {
					nfs.Trash(m, _repos_path(m, m.Option(REPOS)))
					mdb.HashRemove(m, m.Option(REPOS))
				}
			}},
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(REPOS), func(p string) {
					m.Cmd("", CLONE, ORIGIN, p, nfs.PATH, m.Option(cli.CMD_DIR), ice.Maps{cli.CMD_DIR: ""})
				})
			}},
			code.VIMER: {Hand: func(m *ice.Message, arg ...string) { _repos_vimer(m, _repos_path, arg...) }},
		}, mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,commit,origin"), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Sort(REPOS).Action(CLONE, PULL, PUSH)
			} else if len(arg) == 1 {
				_repos_branch(m, _repos_open(m, arg[0]))
			} else if len(arg) == 2 {
				repos := _repos_open(m, arg[0])
				if branch, err := repos.Branch(arg[1]); !m.Warn(err) {
					_repos_log(m, branch, repos)
				}
			} else if len(arg) == 3 {
				if repos := _repos_open(m, arg[0]); arg[2] == INDEX {
					_repos_status(m, arg[0], repos)
				} else {
					_repos_stats(m, repos, arg[2])
				}
			} else {
				m.Cmdy("", code.VIMER, arg)
			}
		}},
	})
}
func ReposList(m *ice.Message) *ice.Message { return m.Cmd(REPOS, ice.OptionFields("repos,path")) }
