// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
)

// Validate command for SlackOverflow.
func Validate(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("validate")
	cmd.SetShortDesc("Validate stackoverflow configuration.")
	cmd.Do(func(w *cli.Worker) {
		w.Fail("cmd not implemented.")
	})
	return cmd
}
