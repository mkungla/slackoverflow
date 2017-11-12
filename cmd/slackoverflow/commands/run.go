// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"os"
	"os/signal"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/addon/application/plugin/cli/flags"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
	"github.com/robfig/cron"
)

// Run command for SlackOverflow.
func Run(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("run")
	cmd.SetShortDesc("Run SlackOverflow once.")

	kaFlag := flags.NewBoolFlag("keep-alive")
	kaFlag.SetUsage("Keep on rumning every minute")
	cmd.AddFlag(kaFlag)

	cmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		keepAlive, err := w.Flag("keep-alive")
		if err != nil {
			w.Fail(err.Error())
			return
		}
		runFull(w, so)
		if keepAlive.Present() {
			cr := cron.New()
			cr.AddFunc("@every 1m", func() {
				runFull(w, so)
			})
			go cr.Start()
			sig := make(chan os.Signal)
			signal.Notify(sig, os.Interrupt, os.Kill)
			<-sig
		}
	})
	return cmd
}

func runFull(w *cli.Worker, so *internal.SlackOverflow) {
	getNewQuestions(w, so)
	updateQuestions(w, so)
	slackPostNewQuestions(w, so)
	slackUpdateQuestions(w, so)
}
