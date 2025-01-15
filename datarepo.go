package main

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log/slog"
	"os"
	"time"
)

func (r *DataRepo) Initiate() error {
	slog.Info("Initiating repo at " + r.BaseFolder)
	repo, err := git.PlainClone(r.BaseFolder, false, &git.CloneOptions{
		URL:      r.GitAddress,
		Progress: os.Stdout,
		Auth:     r.authFunc(),
	})

	r.Repository = repo

	return err
}

func (r *DataRepo) IsExist() bool {
	_, err := os.Stat(r.BaseFolder)
	return err == nil //TODO: correct error handling
}

func (r *DataRepo) OpenRepository() error {
	repo, err := git.PlainOpen(r.BaseFolder)
	r.Repository = repo
	return err
}

func (r *DataRepo) CheckForNewFiles() (error, bool) {
	wt, err := r.Repository.Worktree()
	if err != nil {
		return err, false
	}

	err = wt.AddGlob("*")
	if err != nil {
		return err, false
	}

	status, err := wt.Status()
	if err != nil {
		return err, false
	}

	return nil, !status.IsClean()
}

func (r *DataRepo) Pull() error {
	wt, err := r.Repository.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{Auth: r.authFunc()})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

func (r *DataRepo) DoTheThings() error {
	err := r.Pull()
	if err != nil {
		return err
	}

	return r.CommitAndPush()
}

func createDataRepos(d *Daemon) error {
	for index := range d.Config.Entry {
		dr, err := NewDataRepo(&daemon.Config.Entry[index])

		if err != nil {
			slog.Error("error while initiating", "msg", err.Error())
			return err
		}

		if dr.IsExist() {
			err = dr.OpenRepository()
		} else {
			err = dr.Initiate()
		}

		d.Repos = append(d.Repos, *dr)

		if err != nil {
			slog.Error("error while initiating", "msg", err.Error())
			return err
		}
	}

	return nil
}

func NewDataRepo(cd *ConfigDirectory) (*DataRepo, error) {
	rp := &DataRepo{
		BaseFolder: cd.Path,
		GitAddress: cd.Repository,
		authFunc: func() transport.AuthMethod {
			return nil
		},
		configDirectory: cd,
	}

	if cd.Auth != nil {
		if cd.Auth.Type == "http" || cd.Auth.Type == "https" {
			rp.authFunc = func() transport.AuthMethod {
				return &http.BasicAuth{
					Username: cd.Auth.Username,
					Password: cd.Auth.Password,
				}
			}
		} else if cd.Auth.Type == "agent" {
			auth, err := ssh.NewSSHAgentAuth("git")
			if err != nil {
				return rp, err
			}

			rp.authFunc = func() transport.AuthMethod {
				return auth
			}
		} else {
			slog.Warn("Unknown auth type", "value", cd.Auth.Type)
		}
	}

	return rp, nil
}

func (r *DataRepo) CommitAndPush() error {
	wt, err := r.Repository.Worktree()
	if err != nil {
		return err
	}

	err = wt.AddGlob("*")
	if err != nil {
		return err
	}

	status, err := wt.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		slog.Debug("There is nothing to commit.") //TODO: additional data
		return nil                                //nothing to do
	}

	commitOpts := &git.CommitOptions{
		All: true,
	}

	if r.configDirectory.Author != nil {
		commitOpts.Author = &object.Signature{
			When: time.Now(),
		}

		if r.configDirectory.Author.Name != nil {
			commitOpts.Author.Name = *(r.configDirectory.Author.Name)
		}
		if r.configDirectory.Author.Email != nil {
			commitOpts.Author.Email = *(r.configDirectory.Author.Email)
		}
	}

	_, err = wt.Commit(status.String(), commitOpts)
	if err != nil {
		return err
	}

	err = r.Repository.Push(&git.PushOptions{
		Auth: r.authFunc(),
	})
	return err
}
