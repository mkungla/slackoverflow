// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

import (
	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/lib/filesystem/path"
	"github.com/howi-ce/howi/std/errors"
)

const (
	// ErrNotConfigured is used when SlackOverflow is not configured
	ErrNotConfigured = "You must execute 'slackoverflow reconfigure' or correct errors in ~/.slackoverflow/slackoverflow.yaml"
)

// NewSlackOverflow instance
func NewSlackOverflow() *SlackOverflow {
	return &SlackOverflow{}
}

// SlackOverflow application instance
type SlackOverflow struct {
	// Path retursn SlackOverflow config path
	Path             path.Obj
	ConfigFilePath   path.Obj
	DatabaseFilePath path.Obj
	Config           Config
	DB               Database
	StackExchange    StackExchangeClient
}

// Load SlackOverflow and try to load configuration from given path
func (so *SlackOverflow) Load(w *cli.Worker) (err error) {
	p, _ := w.Flag("config-path")
	if p.Present() {
		so.Path, err = path.New(p.Value().String())
	} else {
		so.Path, err = path.New("~/.slackoverflow")
	}
	if err != nil {
		return err
	}
	so.ConfigFilePath, err = path.New(so.Path.Join("slackoverflow.yaml"))
	if err != nil {
		return err
	}
	so.DatabaseFilePath, err = path.New(so.Path.Join("slackoverflow.db3"))
	if err != nil {
		return err
	}
	so.Config.file = so.ConfigFilePath.Abs()
	if so.Path.Exists() {
		if !so.Path.IsDir() {
			return errors.Newf("%s exists, but is not a directory", so.Path.Abs())
		}
		err = so.Config.readConfig()
	}
	if err != nil {
		return err
	}
	so.DB.SetPath(so.DatabaseFilePath)
	if so.DatabaseFilePath.Exists() {
		if so.DatabaseFilePath.IsDir() {
			return errors.Newf("%s exists, but is a directory", so.Path.Abs())
		}
	}

	return err
}

// IsConfigured returns false with error if SlackOverflow is not configured
func (so *SlackOverflow) IsConfigured() (bool, error) {
	if so.Path.Exists() && so.Config.IsLoaded() {
		return true, nil
	}
	return false, errors.New(ErrNotConfigured)
}

// Session loads everything for SlackOverfloe
func (so *SlackOverflow) Session(w *cli.Worker) error {
	if err := so.Load(w); err != nil {
		return err
	}
	if ok, err := so.IsConfigured(); !ok && err != nil {
		return err
	}
	err := so.DB.VerifyTables(w)

	so.StackExchange.SetHost(so.Config.StackExchange.APIHost)
	so.StackExchange.SetAPIVersion(so.Config.StackExchange.APIVersion)
	so.StackExchange.SetKey(so.Config.StackExchange.Key)

	return err
}

// SyncQuestion questions
func (so *SlackOverflow) SyncQuestion(w *cli.Worker, q QuestionObj) {
	var ok string
	var err error
	// Create or Update user
	ok, err = so.DB.SyncStackExchangeUserShallowUser(q.Owner)
	if err != nil {
		w.Log.Error(err)
	} else {
		w.Log.Ok(ok)
	}
	// Create or Update user
	ok, err = so.DB.SyncStackExchangeQuestion(q, so.Config.StackExchange.Site)
	if err != nil {
		w.Log.Error(err)
	} else {
		w.Log.Ok(ok)
	}
}
