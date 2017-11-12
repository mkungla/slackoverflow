// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"os"
	"os/signal"
	"time"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/addon/application/plugin/cli/flags"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
	"github.com/robfig/cron"
)

var fromDate time.Time

// StackExchange ...
func StackExchange(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("stackexchange")
	cmd.SetShortDesc("Stack Exchange related commands see slackoverflow stackexchange --help for more info.")
	cmd.AddSubcommand(StackExchangeQuestions(so))
	cmd.AddSubcommand(StackExchangeWatch(so))

	return cmd
}

// StackExchangeQuestions ...
func StackExchangeQuestions(so *internal.SlackOverflow) cli.Command {
	scmd := cli.NewCommand("questions")
	scmd.SetShortDesc("Work with stackexchange questions based on the config.")

	gnqFlag := flags.NewBoolFlag("get")
	gnqFlag.SetUsage("Get new questions from configured Stack Exchange Site")
	scmd.AddFlag(gnqFlag)

	uqFlag := flags.NewBoolFlag("update")
	uqFlag.SetUsage("Update information about existing questions")
	scmd.AddFlag(uqFlag)

	sFlag := flags.NewBoolFlag("sync")
	sFlag.SetUsage("Get new questions and update information about existing questions")
	scmd.AddFlag(sFlag)

	scmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		get, err := w.Flag("get")
		if err != nil {
			w.Fail(err.Error())
			return
		}

		update, err := w.Flag("update")
		if err != nil {
			w.Fail(err.Error())
			return
		}
		sync, err := w.Flag("sync")
		if err != nil {
			w.Fail(err.Error())
			return
		}

		if sync.Present() {
			getNewQuestions(w, so)
			updateQuestions(w, so)
		} else if get.Present() {
			getNewQuestions(w, so)
		} else if update.Present() {
			updateQuestions(w, so)
		} else {
			w.Fail("atleast one flag must be provided")
			return
		}
	})
	scmd.AfterAlways(func(w *cli.Worker) {
		if err := so.DB.Close(); err != nil {
			w.Log.Error(err)
		}
		w.Log.Infof(
			"Stack Exchange Quota usage (%d/%d)",
			so.StackExchange.GetQuotaRemaining(),
			so.StackExchange.GetQuotaMax(),
		)
	})
	return scmd
}

// StackExchangeWatch ...
func StackExchangeWatch(so *internal.SlackOverflow) cli.Command {
	scmd := cli.NewCommand("watch")
	scmd.SetShortDesc("Watch new questions from Stack Exchange site (updated every minute nothing stored to db or posted to slack)")
	scmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		w.Log.Ok("Waiting for new questions!")
		fromDate = time.Now().UTC().Add(-30 * time.Minute)
		startWatching(w, so)
		watch := cron.New()
		watch.AddFunc("@every 1m", func() {
			startWatching(w, so)
		})
		go watch.Start()
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig
	})
	scmd.AfterAlways(func(w *cli.Worker) {
		if err := so.DB.Close(); err != nil {
			w.Log.Error(err)
		}
		w.Log.Infof(
			"Stack Exchange Quota usage (%d/%d)",
			so.StackExchange.GetQuotaRemaining(),
			so.StackExchange.GetQuotaMax(),
		)
	})
	return scmd
}

func getNewQuestions(w *cli.Worker, so *internal.SlackOverflow) {
	w.Log.Info("Stack Exchange: Checking for new questions.")

	var empty bool
	latest, _ := so.DB.LatestStackExchangeQuestion()
	if latest.QID == 0 {
		empty = true
	}

	if empty {
		w.Log.Info("There are no questions in database,")
		w.Log.Info("That is ok if current execution is first time you run slackoverflow")
		w.Log.Infof("Or there has been no questions tagged with %q on site %q",
			so.Config.StackExchange.SearchAdvanced["tagged"],
			so.Config.StackExchange.Site,
		)
		latest.CreationDate = time.Now().UTC().Add(-4 * time.Hour)
	}

	now := time.Now().UTC().Unix()
	diff := latest.CreationDate.Unix() - now
	// Allow maximum 4 h old question as from date otherwise we may exhaust
	// rate limit if last tracked questions fromdata is to far past
	// and slackoverflow has not been running for a while.
	if diff > (4 * 60) {
		latest.CreationDate = time.Now().UTC().Add(-4 * time.Hour)
	}

	w.Log.Infof("Checking new questions since %s", latest.CreationDate.String())

	// Check for New Questions from Stack Exchange
	searchAdvanced := so.StackExchange.SearchAdvanced()

	// Set it here so that it is allowed to override by config
	searchAdvanced.Parameters.Set("site", so.Config.StackExchange.Site)

	// Set all parameters from config
	for param, value := range so.Config.StackExchange.SearchAdvanced {
		searchAdvanced.Parameters.Set(param, value)
	}

	searchAdvanced.Parameters.Set("fromdate", latest.CreationDate.Unix()+1)

	// Output query as table
	d, _ := w.Flag("debug")
	debbuging := d.Present()
	if debbuging {
		searchAdvanced.DrawQuery(w)
	}

	fetchQuestions := true
	var lastQuestion internal.QuestionObj
	for fetchQuestions {
		w.Log.Debugf("Fetching page %d", searchAdvanced.GetCurrentPageNr())
		if results, err := searchAdvanced.Get(); results {
			// Questions recieved
			for _, q := range searchAdvanced.Result.Items {

				w.Log.Infof("Question: %s", q.Title)
				w.Log.Infof("Url:      %s", q.ShareLink)

				if debbuging {
					newq := internal.NewTable("Question ID", "Time", "Answers", "Comments", "Score", "Views", "Username")
					newq.AddRow(
						q.QID,
						time.Unix(q.CreationDate, 0).UTC().Format("15:04:05 Mon Jan _2 2006"),
						q.AnswerCount,
						q.CommentCount,
						q.Score,
						q.ViewCount,
						q.Owner.DisplayName,
					)
					newq.Print()
				}
				// Skip sync if there are locally no questions
				if empty {
					lastQuestion = q
					continue
				}
				so.SyncQuestion(w, q)
			}
			if err != nil {
				fetchQuestions = false
				w.Log.Error(err)
			}
		}

		// Done go to next page
		if searchAdvanced.GetCurrentPageNr() > 10 {
			fetchQuestions = false
		} else if searchAdvanced.HasMore() {
			searchAdvanced.NextPage()
		} else {
			fetchQuestions = false
			w.Log.Debug("There are no more new questions.")
		}
	}
	if empty && lastQuestion.QID > 0 {
		so.SyncQuestion(w, lastQuestion)
	}
}

func updateQuestions(w *cli.Worker, so *internal.SlackOverflow) {
	w.Log.Info("Stack Exchange: Updating existing questions.")

	questionIds, questionIdsCount := so.DB.StackExchangeQuestionTrackedIds(
		so.Config.StackExchange.QuestionsToWatch)

	// Check do we already have some questions do obtain from_date for next request
	if questionIdsCount == 0 {
		w.Log.Info("There are no questions in database,")
		w.Log.Info("That is ok if current execution is first time you run slackoverflow")
		w.Log.Info("and this case run 'slackoverflow stackechange guestions' --get first")
	}

	w.Log.Infof("Checking updates for %d questions. Max to be tracked (%d)",
		questionIdsCount, so.Config.StackExchange.QuestionsToWatch)

	// Check for New Questions from Stack Exchange
	updateQuestions := so.StackExchange.Questions()

	// Set it here so that it is allowed to override by config
	updateQuestions.Parameters.Set("site", so.Config.StackExchange.Site)

	for param, value := range so.Config.StackExchange.Questions {
		updateQuestions.Parameters.Set(param, value)
	}

	// Output query as table
	d, _ := w.Flag("debug")
	debbuging := d.Present()
	if debbuging {
		updateQuestions.DrawQuery(w, questionIds)
	}

	fetchQuestions := true
	for fetchQuestions {
		w.Log.Debugf("Fetching page %d", updateQuestions.GetCurrentPageNr())
		if results, err := updateQuestions.Get(questionIds); results {
			// Questions recieved
			for _, q := range updateQuestions.Result.Items {

				w.Log.Infof("Question: %s", q.Title)
				w.Log.Infof("Url:      %s", q.ShareLink)
				if debbuging {
					newq := internal.NewTable("Question ID", "Time", "Answers", "Comments", "Score", "Views", "Username")
					newq.AddRow(
						q.QID,
						time.Unix(q.CreationDate, 0).UTC().Format("15:04:05 Mon Jan _2 2006"),
						q.AnswerCount,
						q.CommentCount,
						q.Score,
						q.ViewCount,
						q.Owner.DisplayName,
					)
					newq.Print()
				}

				so.SyncQuestion(w, q)
			}
			if err != nil {
				fetchQuestions = false
				w.Log.Error(err.Error())
			}
		}

		// Done go to next page
		if updateQuestions.HasMore() {
			updateQuestions.NextPage()
		} else {
			fetchQuestions = false
			w.Log.Debug("There are no more questions to update.")
		}
	}
}

func startWatching(w *cli.Worker, so *internal.SlackOverflow) {
	// Check for New Questions from Stack Exchange
	searchAdvanced := so.StackExchange.SearchAdvanced()

	// Set it here so that it is allowed to override by config
	searchAdvanced.Parameters.Set("site", so.Config.StackExchange.Site)

	// Set all parameters from config
	for param, value := range so.Config.StackExchange.SearchAdvanced {
		searchAdvanced.Parameters.Set(param, value)
	}
	searchAdvanced.Parameters.Set("tagged", "php")
	searchAdvanced.Parameters.Set("fromdate", fromDate.Unix()+1)

	fetchQuestions := true
	for fetchQuestions {
		w.Log.Debugf("Fetching page %d", searchAdvanced.GetCurrentPageNr())
		if results, err := searchAdvanced.Get(); results {
			// Questions recieved
			for _, q := range searchAdvanced.Result.Items {
				fromDate = time.Unix(q.CreationDate, 0).UTC()
				w.Log.Linef("Question: %s", q.Title)
				w.Log.Linef("Url:      %s", q.ShareLink)
				newq := internal.NewTable("Question ID", "Time", "Answers", "Comments", "Score", "Views", "Username")
				newq.AddRow(
					q.QID,
					time.Unix(q.CreationDate, 0).Local().Format("15:04:05 Mon Jan _2 2006"),
					q.AnswerCount,
					q.CommentCount,
					q.Score,
					q.ViewCount,
					q.Owner.DisplayName,
				)
				newq.Print()
			}
			if err != nil {
				fetchQuestions = false
				w.Log.Error(err.Error())
			}

			if len(searchAdvanced.Result.Items) > 0 {
				w.Log.Ok("Waiting for new questions!")
			}
		}

		// Done go to next page
		if searchAdvanced.HasMore() || searchAdvanced.GetCurrentPageNr() > 10 {
			searchAdvanced.NextPage()
		} else {
			fetchQuestions = false
		}
	}

	w.Log.Infof(
		"Stack Exchange Quota usage (%d/%d)",
		so.StackExchange.GetQuotaRemaining(),
		so.StackExchange.GetQuotaMax(),
	)
}
