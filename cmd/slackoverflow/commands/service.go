// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
)

// Service command for SlackOverflow.
func Service(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("service")
	cmd.SetShortDesc("SlackOverflow daemon commands.")
	cmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		//
	})
	return cmd
}
