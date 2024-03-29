package git

import (
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
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/lex"
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
func _repos_recent(m *ice.Message, repos *git.Repository) (r *plumbing.Reference) {
	max := 0
	if iter, err := repos.Tags(); err == nil {
		for {
			refer, err := iter.Next()
			if err != nil {
				break
			}
			ls := kit.Split(refer.Name().Short(), "v.")
			if len(ls) < 2 {
				continue
			}
			if n := kit.Int(ls[0])*1000000 + kit.Int(ls[1])*1000 + kit.Int(ls[2]); n > max {
				max, r = n, refer
			}
		}
	}
	return
}
func _repos_forword(m *ice.Message, repos *git.Repository, version string) int {
	if refer, err := repos.Head(); err == nil {
		if commit, err := repos.CommitObject(refer.Hash()); err == nil {
			n := 0
			for {
				if commit.Hash.String() == version {
					break
				}
				if commit, err = commit.Parent(0); err != nil || commit == nil {
					break
				}
				n++
			}
			return n
		}
	}
	return 0
}
func _repos_insert(m *ice.Message, p string) {
	if repos, err := git.PlainOpen(p); err == nil {
		args := []string{REPOS, path.Base(p), nfs.PATH, p}
		if head, err := repos.Head(); err == nil {
			args = append(args, BRANCH, head.Name().Short())
			if commit, err := repos.CommitObject(head.Hash()); err == nil {
				args = append(args, mdb.TIME, commit.Author.When.Format(ice.MOD_TIME), COMMENT, strings.TrimSuffix(commit.Message, lex.NL))
			}
		}
		if refer := _repos_recent(m, repos); refer != nil {
			if n := _repos_forword(m, repos, refer.Hash().String()); n > 0 {
				args = append(args, VERSION, kit.Format("%s-%d-%s", refer.Name().Short(), n, kit.Cut(refer.Hash().String(), 6)))
			} else {
				args = append(args, VERSION, refer.Name().Short())
			}
		}
		args = append(args, ORIGIN, _repos_origin(m, repos))
		mdb.HashCreate(m.Options(mdb.TARGET, repos), args)
	}
}
func _repos_origin(m *ice.Message, repos *git.Repository) string {
	if remote, err := repos.Remote(ORIGIN); err == nil {
		return remote.Config().URLs[0]
	} else if remote, err := repos.Remotes(); err == nil && len(remote) > 0 {
		return remote[0].Config().URLs[0]
	}
	return ""
}
func _repos_path(m *ice.Message, p string, arg ...string) string {
	if p == path.Base(kit.Path("")) {
		return kit.Path("", arg...)
	}
	if nfs.Exists(m, path.Join(nfs.USR, p, ".git")) {
		return path.Join(nfs.USR, p, path.Join(arg...))
	}
	if nfs.Exists(m, path.Join(nfs.USR_LOCAL_WORK, p, ".git")) {
		return path.Join(nfs.USR_LOCAL_WORK, p, path.Join(arg...))
	}
	return p
}
func _repos_open(m *ice.Message, p string) *git.Repository {
	return mdb.HashSelectTarget(m, p, nil).(*git.Repository)
}
func _repos_each(m *ice.Message, title string, cb func(*git.Repository, ice.Maps) error) {
	msg := m.Cmd("")
	if msg.Length() == 0 {
		return
	}
	web.GoToast(m, kit.Select(m.CommandKey()+lex.SP+m.ActionKey(), title), func(toast func(string, int, int)) (list []string) {
		count, total := 0, msg.Length()
		msg.Table(func(value ice.Maps) {
			toast(value[REPOS], count, total)
			if err := cb(_repos_open(m, value[REPOS]), value); err != nil && err != git.NoErrAlreadyUpToDate {
				web.Toast(m, err.Error(), "error: "+value[REPOS], "", "3s")
				list = append(list, value[REPOS])
				m.Sleep3s()
			}
			count++
		})
		return
	})

}
func _repos_auth(m *ice.Message, origin string) *http.BasicAuth {
	list := _repos_credentials(m)
	if insteadof := mdb.Config(m, INSTEADOF); insteadof != "" {
		origin = insteadof + path.Base(origin)
	}
	if u, ok := list[kit.ParseURL(origin).Host]; !ok {
		return nil
	} else if password, ok := u.User.Password(); !ok {
		return nil
	} else {
		return &http.BasicAuth{Username: u.User.Username(), Password: password}
	}
}
func _repos_each_origin(m *ice.Message, title string, cb func(*git.Repository, string, *http.BasicAuth, ice.Maps) error) {
	_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
		if value[ORIGIN] == "" {
			return nil
		} else if remote, err := repos.Remote(ORIGIN); err != nil {
			return err
		} else {
			remoteURL := remote.Config().URLs[0]
			if insteadof := mdb.Config(m, INSTEADOF); insteadof != "" {
				remoteURL = insteadof + path.Base(remoteURL)
			}
			return cb(repos, remoteURL, _repos_auth(m, remote.Config().URLs[0]), value)
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
			m.Push(BRANCH, refer.Name().Short())
			m.Push(aaa.USERNAME, commit.Author.Name)
			m.Push(mdb.TEXT, commit.Message)
		}
		return nil
	})
	return nil
}
func _repos_log(m *ice.Message, hash plumbing.Hash, repos *git.Repository) error {
	iter, err := repos.Log(&git.LogOptions{From: hash})
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
		if kit.IsIn(k, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO, ice.ETC_LOCAL_SHY) {
			continue
		} else if kit.IsIn(kit.Ext(k), "swp", "swo") || kit.IsIn(kit.Split(k, nfs.PS)[0], ice.BIN, ice.VAR, ice.USR) && !strings.HasPrefix(k, ice.USR_LOCAL_EXPORT) {
			continue
		}
		if m.Push(REPOS, p).Push(STATUS, string(v.Worktree)+string(v.Staging)).Push(nfs.FILE, k); m.Option(ice.MSG_MODE) == mdb.ZONE {
			ls := nfs.SplitPath(m, kit.Path(_repos_path(m, p), k))
			m.Push(nfs.PATH, ls[0]).Push(mdb.TEXT, string(v.Worktree)+string(v.Staging)+lex.SP+ls[0]+ls[1])
		}
		switch v.Worktree {
		case git.Untracked:
			m.PushButton(ADD, nfs.TRASH)
		case git.Modified:
			m.PushButton(COMMIT)
		default:
			m.PushButton(COMMIT)
		}
	}
	if p == path.Base(kit.Path("")) {
		var tree *object.Tree
		if refer, err := repos.Head(); err == nil {
			if commit, err := repos.CommitObject(refer.Hash()); err == nil {
				tree, err = commit.Tree()
			}
		}
		m.Cmd(nfs.DIR, ice.USR_LOCAL_EXPORT, kit.Dict(ice.MSG_FILES, nfs.DiskFile, nfs.DIR_DEEP, ice.TRUE, nfs.DIR_TYPE, nfs.TYPE_CAT), func(value ice.Maps) {
			if _, ok := status[value[nfs.PATH]]; ok {
				return
			} else if tree != nil {
				if file, err := tree.File(value[nfs.PATH]); err == nil {
					if content, err := file.Contents(); err == nil && strings.TrimSpace(content) == strings.TrimSpace(m.Cmdx(nfs.CAT, value[nfs.PATH])) {
						return
					} else {
						m.Push(REPOS, p).Push(STATUS, "M").Push(nfs.FILE, value[nfs.PATH]).PushButton(ADD)
						return
					}
				}
			}
			m.Push(REPOS, p).Push(STATUS, "??").Push(nfs.FILE, value[nfs.PATH]).PushButton(ADD, nfs.TRASH)
		})
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
func _repos_inner(m *ice.Message, _repos_path func(m *ice.Message, p string, arg ...string) string, arg ...string) {
	if len(arg) == 0 || arg[0] != ice.RUN {
		arg = []string{path.Join(arg[:3]...) + nfs.PS, kit.Select("README.md", arg, 3)}
	} else if kit.Select("", arg, 1) != ctx.ACTION {
		if ls := kit.Split(path.Join(m.Option(nfs.DIR_ROOT), arg[1]), nfs.PS); len(ls) < 2 || ls[2] == INDEX {
			if repos := _repos_open(m, ls[0]); len(arg) < 3 {
				// m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, _repos_path(m, ls[0])))
				if work, err := repos.Worktree(); err == nil {
					if status, err := work.Status(); err == nil {
						for k := range status {
							m.Push(nfs.PATH, k)
							// m.Echo(k)
						}
					}
				}
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
					// m.Push(nfs.PATH, file.Name)
					return nil
				})
			}
			if stats, err := commit.Stats(); err == nil {
				for _, stat := range stats {
					m.Push(nfs.PATH, stat.Name)
					// m.Echo(stat.Name)
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
		ctx.DisplayLocal(m, "code/inner.js", "style", "output")
		return
	}
	ctx.ProcessField(m, "", arg, arg...)
}
func _repos_credentials(m *ice.Message) map[string]*url.URL {
	list := map[string]*url.URL{}
	m.Cmd(nfs.CAT, kit.HomePath(".git-credentials"), func(line string) {
		u := kit.ParseURL(strings.ReplaceAll(line, "%3a", ":"))
		list[u.Host] = u
	})
	return list
}

const (
	INIT   = "init"
	CLONE  = "clone"
	PULL   = "pull"
	PUSH   = "push"
	LOG    = "log"
	TAG    = "tag"
	ADD    = "add"
	STASH  = "stash"
	COMMIT = "commit"
	BRANCH = "branch"

	REMOTE    = "remote"
	ORIGIN    = "origin"
	MASTER    = "master"
	INDEX     = "index"
	INSTEADOF = "insteadof"
)
const REPOS = "repos"

func init() {
	cache := ice.USR_REQUIRE
	web.Index.MergeCommands(ice.Commands{
		web.PP(ice.REQUIRE): {Name: "/require/shylinux.com/x/volcanos/proto.js", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(cli.SystemFind(m, code.GO), func() { cache = code.GoCache(m) })
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 4 {
				m.RenderStatusBadRequest()
			} else if path.Join(arg[:3]...) == ice.Info.Make.Module && nfs.Exists(m, path.Join(arg[3:]...)) {
				m.RenderDownload(path.Join(arg[3:]...))
			} else if p := path.Join(nfs.USR, kit.Split(arg[2], mdb.AT)[0], path.Join(arg[3:]...)); nfs.Exists(m, p) && kit.Select("", kit.Split(arg[2], mdb.AT), 1) == ice.Info.Gomod[path.Join(arg[0], arg[1], kit.Split(arg[2], mdb.AT)[0])] {
				m.RenderDownload(p)
			} else {
				if p = path.Join(cache, path.Join(arg...)); !nfs.Exists(m, p) {
					if p = path.Join(ice.USR_REQUIRE, path.Join(arg...)); !nfs.Exists(m, p) {
						ls := strings.SplitN(path.Join(arg[:3]...), mdb.AT, 2)
						_, err := git.PlainClone(path.Join(ice.USR_REQUIRE, path.Join(arg[:3]...)), false, &git.CloneOptions{
							URL: "http://" + ls[0], Depth: 1, ReferenceName: plumbing.NewTagReferenceName(kit.Select(ice.Info.Gomod[ls[0]], ls, 1)),
						})
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
				m.Cmd(nfs.DIR, nfs.USR_LOCAL_WORK, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
				m.Cmd(nfs.DIR, nfs.USR, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
				_repos_insert(m, kit.Path(""))
				m.Cmd(CONFIGS, func(value ice.Maps) {
					if strings.HasSuffix(value[mdb.NAME], ".insteadof") && strings.HasPrefix(ice.Info.Make.Remote, value[mdb.VALUE]) {
						mdb.Config(m, INSTEADOF, strings.TrimPrefix(strings.TrimSuffix(value[mdb.NAME], ".insteadof"), "url."))
					}
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case COMMENT:
					ls := kit.Split(m.Option(nfs.FILE), " /")
					m.Push(arg[0], kit.Join(kit.Slice(ls, -1), nfs.PS))
					m.Push(arg[0], kit.Join(kit.Slice(ls, -2), nfs.PS))
					m.Push(arg[0], m.Option(nfs.FILE))
				case VERSION:
					m.Push(VERSION, _status_tag(m, m.Option(TAGS)))
				}
			}},
			CLONE: {Name: "clone origin* branch name path", Help: "克隆", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(mdb.NAME, path.Base(m.Option(ORIGIN)))
				m.OptionDefault(nfs.PATH, path.Join(path.Join(nfs.USR, m.Option(mdb.NAME))))
				if _, err := git.PlainClone(m.Option(nfs.PATH), false, &git.CloneOptions{URL: m.Option(ORIGIN), Auth: _repos_auth(m, m.Option(ORIGIN))}); m.Warn(err) {
					_repos_insert(m, m.Option(nfs.PATH))
				}
			}},
			INIT: {Name: "clone origin* branch name path", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DEFS, kit.Path(".git/config"), nfs.Template(m, "config", m.Option(ORIGIN)))
				git.PlainInit(m.Option(nfs.PATH), false)
				_repos_insert(m, kit.Path(""))
				m.ProcessRefresh()
			}},
			PULL: {Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				_repos_each_origin(m, "", func(repos *git.Repository, remoteURL string, auth *http.BasicAuth, value ice.Maps) error {
					work, err := repos.Worktree()
					if err != nil {
						return err
					}
					return work.Pull(&git.PullOptions{RemoteURL: remoteURL, Auth: auth})
				})
			}},
			PUSH: {Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				_repos_each_origin(m, "", func(repos *git.Repository, remoteURL string, auth *http.BasicAuth, value ice.Maps) error {
					m.Cmd(cli.SYSTEM, GIT, PUSH, "--tags", kit.Dict(cli.CMD_DIR, path.Join(ice.USR_LOCAL_REPOS, value[REPOS])))
					return repos.Push(&git.PushOptions{RemoteURL: remoteURL, Auth: auth, FollowTags: true})
				})
			}},
			LOG: {Hand: func(m *ice.Message, arg ...string) {
				repos := _repos_open(m, kit.Select(m.Option(REPOS), arg, 0))
				if branch, err := repos.Branch(kit.Select(m.Option(BRANCH), arg, 1)); !m.Warn(err) {
					if refer, err := repos.Reference(branch.Merge, true); !m.Warn(err) {
						_repos_log(m, refer.Hash(), repos)
					}
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
			COMMIT: {Name: "commit actions=add,fix,opt message*=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if work, err := _repos_open(m, m.Option(REPOS)).Worktree(); !m.Warn(err) {
					opt := &git.CommitOptions{All: true}
					if cfg, err := config.LoadConfig(config.GlobalScope); err == nil {
						if cfg.Author.Email == "" || cfg.Author.Name == "" {
							opt.Author = &object.Signature{
								Email: kit.Select(m.Option(ice.MSG_USERNAME)+"@shylinux.com", mdb.Config(m, aaa.EMAIL)),
								Name:  kit.Select(m.Option(ice.MSG_USERNAME), mdb.Config(m, aaa.USERNAME)),
								When:  time.Now(),
							}
						}
					}
					_, err := work.Commit(m.Option("actions")+lex.SP+m.Option("message"), opt)
					m.Warn(err)
				}
			}},
			STATUS: {Help: "状态", Hand: func(m *ice.Message, arg ...string) {
				if repos := kit.Select(m.Option(REPOS), arg, 0); repos != "" {
					_repos_status(m, repos, _repos_open(m, repos))
				} else {
					last, password, list := "", "", _repos_credentials(m)
					_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
						if refer, err := repos.Head(); err == nil {
							if commit, err := repos.CommitObject(refer.Hash()); err == nil {
								_last := commit.Author.When.Format(ice.MOD_TIME)
								kit.If(_last > last, func() { last = _last })
							}
						}
						return _repos_status(m, value[REPOS], repos)
					})
					remote := ice.Info.Make.Remote
					if repos, ok := mdb.HashSelectTarget(m, path.Base(kit.Path("")), nil).(*git.Repository); ok {
						remote = kit.Select(remote, _repos_origin(m, repos))
					}
					if insteadof := mdb.Config(m, INSTEADOF); insteadof != "" {
						remote = insteadof + path.Base(remote)
					}
					if u, ok := list[kit.ParseURL(remote).Host]; ok {
						password, _ = u.User.Password()
					}
					m.Sort("repos,status,file").Status(mdb.TIME, last, REMOTE, remote, kit.Select(aaa.TECH, aaa.VOID, password == ""), m.Option(aaa.EMAIL), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], kit.MDB_COST, m.FormatCost())
				}
			}},
			TOTAL: {Hand: func(m *ice.Message, arg ...string) {
				stats := map[string]int{}
				if repos := kit.Select(m.Option(REPOS), arg, 0); repos == "" {
					var from *time.Time
					_repos_each(m, "", func(repos *git.Repository, value ice.Maps) error {
						if kit.IsIn(value[REPOS], "go-git", "go-qrcode", "websocket", "webview", "wubi-dict", "word-dict") {
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
			INSTEADOF: {Name: "insteadof remote", Help: "代理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(CONFIGS, func(value ice.Maps) {
					if strings.HasSuffix(value[mdb.NAME], ".insteadof") && strings.HasPrefix(ice.Info.Make.Remote, value[mdb.VALUE]) {
						_git_cmd(m, CONFIG, "--global", "--unset", value[mdb.NAME])
					}
				})
				if mdb.Config(m, INSTEADOF, m.Option(REMOTE)); m.Option(REMOTE) != "" {
					_git_cmd(m, CONFIG, "--global", "url."+m.Option(REMOTE)+".insteadof", strings.TrimSuffix(ice.Info.Make.Remote, path.Base(ice.Info.Make.Remote)))
				}
			}},
			"remoteURL": {Hand: func(m *ice.Message, arg ...string) {
				remoteURL := ""
				if repos, ok := mdb.HashSelectTarget(m, path.Base(kit.Path("")), nil).(*git.Repository); ok {
					remoteURL = kit.Select(remoteURL, _repos_origin(m, repos))
				}
				if insteadof := mdb.Config(m, INSTEADOF); insteadof != "" {
					remoteURL = insteadof + path.Base(remoteURL)
				}
				m.Echo(remoteURL)
			}},
			REMOTE: {Hand: func(m *ice.Message, arg ...string) {
				repos := _repos_open(m, kit.Select(path.Base(kit.Path("")), kit.Select(m.Option(REPOS), arg, 0)))
				if _remote, err := repos.Remote(ORIGIN); err == nil {
					m.Push(REMOTE, kit.Select("", _remote.Config().URLs, 0))
				}
				if refer := _repos_recent(m, repos); refer != nil {
					m.Push(nfs.VERSION, refer.Name().Short())
					m.Push("forword", _repos_forword(m, repos, refer.Hash().String()))
				}
				if refer, err := repos.Head(); err == nil {
					m.Push(BRANCH, refer.Name().Short())
					m.Push(mdb.HASH, refer.Hash().String())
					if commit, err := repos.CommitObject(refer.Hash()); err == nil {
						m.Push(aaa.EMAIL, commit.Author.Email)
						m.Push("author", commit.Author.Name)
						m.Push("when", commit.Author.When)
						m.Push("message", commit.Message)
					}
				}
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
					p = strings.Split(p, mdb.QS)[0]
					kit.If(!strings.Contains(p, "://"), func() { p = web.UserHost(m) + "/x/" + p })
					if ice.Info.System == cli.LINUX {
						p = strings.Replace(p, "https", "http", 1)
					}
					m.Cmd("", CLONE, ORIGIN, p, nfs.PATH, m.Option(cli.CMD_DIR), ice.Maps{cli.CMD_DIR: ""})
				})
			}},
			web.DREAM_TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", mdb.REMOVE, kit.Dict(REPOS, m.Option(mdb.NAME)))
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if !kit.IsIn(m.Option(mdb.TYPE), web.WORKER, web.SERVER) {
					return
				} else if !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME), ".git")) {
					return
				}
				m.PushButton(kit.Dict(m.CommandKey(), "源码"))
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, []string{}, arg...) }},
			code.INNER:       {Hand: func(m *ice.Message, arg ...string) { _repos_inner(m, _repos_path, arg...) }},
		}, aaa.RoleAction(REMOTE), gdb.EventsAction(web.DREAM_CREATE, web.DREAM_TRASH), mdb.ClearOnExitHashAction(), mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,version,comment,origin")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Sort(REPOS).PushAction(STATUS, mdb.REMOVE).Action(CLONE, PULL, PUSH, STATUS)
			} else if len(arg) == 1 {
				_repos_branch(m, _repos_open(m, arg[0]))
			} else if len(arg) == 2 {
				repos := _repos_open(m, arg[0])
				if branch, err := repos.Branch(arg[1]); !m.Warn(err) {
					if refer, err := repos.Reference(branch.Merge, true); !m.Warn(err) {
						_repos_log(m, refer.Hash(), repos)
					}
				}
			} else if len(arg) == 3 {
				if repos := _repos_open(m, arg[0]); arg[2] == INDEX {
					_repos_status(m, arg[0], repos)
				} else {
					_repos_stats(m, repos, arg[2])
				}
			} else {
				m.Cmdy("", code.INNER, arg)
			}
		}},
	})
}
func ReposList(m *ice.Message) *ice.Message { return m.Cmd(REPOS, ice.OptionFields("repos,path")) }
func ReposClone(m *ice.Message, arg ...string) *ice.Message {
	return m.Cmdy("web.code.git.repos", "clone", arg)
}
