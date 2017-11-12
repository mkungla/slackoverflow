// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
)

// Config command for SlackOverflow
func Config(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("config")
	cmd.SetShortDesc("Display SlackOverflow configuration.")
	cmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		printConfig(w, so)
	})
	return cmd
}

func printConfig(w *cli.Worker, so *internal.SlackOverflow) {

	si := internal.NewTable("Session Info", " ")
	si.AddRow("Project path", so.Path.Abs())
	si.AddRow("Config file", so.ConfigFilePath.Abs())
	si.AddRow("Database file", so.DatabaseFilePath.Abs())
	si.Print()

	slack := internal.NewTable("Slack Configuration", " ")
	slack.AddRow("API host", so.Config.Slack.APIHost)
	slack.AddRow("Token", so.Config.Slack.Token)
	slack.AddRow("Team name", so.Config.Slack.TeamInfo.Name)
	slack.AddRow("Team domain", so.Config.Slack.TeamInfo.Domain)
	slack.AddRow("Team icon", so.Config.Slack.TeamInfo.Icon["image_original"])
	slack.AddRow("Team email domain", so.Config.Slack.TeamInfo.EmailDomain)

	slack.AddRow("Channel", so.Config.Slack.Channel)
	slack.AddRow("Channel name", so.Config.Slack.ChannelName)
	slack.Print()

	stackexchange := internal.NewTable("StackExchange Configuration", " ")
	stackexchange.AddRow("API Host", so.Config.StackExchange.APIHost)
	stackexchange.AddRow("API Version", so.Config.StackExchange.APIVersion)

	stackexchange.AddRow("Key", so.Config.StackExchange.Key)

	stackexchange.AddRow("Site", so.Config.StackExchange.Site)
	stackexchange.AddRow("Tagged", so.Config.StackExchange.SearchAdvanced["tagged"])
	stackexchange.AddRow("Questions to watch", so.Config.StackExchange.QuestionsToWatch)
	stackexchange.Print()
}
