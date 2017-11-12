// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/std/errors"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
	"github.com/nlopes/slack"
)

// Reconfigure command for SlackOverflow
func Reconfigure(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("reconfigure")
	cmd.SetShortDesc("Interactive configuration of stackoverflow.")
	cmd.Do(func(w *cli.Worker) {
		err := so.Load(w)
		if err != nil {
			w.Fail(err.Error())
			return
		}
		ok, _ := so.IsConfigured()
		w.Log.Okf("path set to %q", so.Path.Abs())
		if ok {
			w.Log.Warning("SlcakOverflow seems to be configured and proceeding will override existing configuration.")
			printConfig(w, so)

			if c := w.AskForConfirmation("Start interactive confguration?"); !c {
				w.Log.Ok("Reconfigure canceled by user")
				return
			}
		}

		err = os.MkdirAll(so.Path.Abs(), os.ModePerm)
		if err != nil {
			w.Fail(err.Error())
			return
		}
		slack := w.AskForConfirmation("Would you like to Configure Slack?")
		if slack {
			err := configureSlack(w, so)
			if err != nil {
				w.Fail(err.Error())
				return
			}
			w.Log.Ok("Slack Slack API Client is configured")
		}

		stackexchange := w.AskForConfirmation("Would you like to Configure Stack Exchange?")
		if stackexchange {
			configureStackExchange(w, so)
			w.Log.Ok("Slack Stack Exchange API Client is configured")
		}
		w.Log.Ok("configuration done")
	})
	return cmd
}

// ConfigureSlack for SlackOverflow
func configureSlack(w *cli.Worker, so *internal.SlackOverflow) error {
	w.Log.Notice("Configuring Slack API Client")

	// Set Slack Defaults
	so.Config.Slack.SetAPIhost("https://slack.com/api")
	so.Config.Slack.Enable()

	reader := bufio.NewReader(os.Stdin)

	w.Log.Line("Enter your @stackoverflow Slack BOT API Token.")
	w.Log.Line("You can create a bot and get token at https://<your-team>.slack.com/apps/manage/custom-integrations")

	token, _ := reader.ReadString('\n')
	so.Config.Slack.SetToken(strings.TrimSpace(token))

	w.Log.Line("Fetching available Slack channels.")
	api := slack.New(so.Config.Slack.Token)
	channels, err := api.GetChannels(true)
	if err != nil {
		return err
	}

	if len(channels) > 0 {
		chlist := internal.NewTable("ID", "Name", "Created")
		for _, channel := range channels {
			chlist.AddRow(
				channel.ID,
				channel.Name,
				fmt.Sprintf("%d", channel.Created),
			)
		}
		chlist.Print()
	} else {
		return errors.New("Unable to fetch any channels with provided credentials")
	}

	w.Log.Line("Enter Channel ID which you want to post the questions")
	ch, _ := reader.ReadString('\n')
	ch = strings.TrimSpace(ch)
	so.Config.Slack.SetChannel(ch)
	for _, channel := range channels {
		if channel.ID == ch {
			so.Config.Slack.SetChannelName(channel.Name)
		}
	}

	team, err := api.GetTeamInfo()
	if err != nil {
		return err
	}
	so.Config.Slack.SetTeamInfo(team)
	return so.Config.Save()
}

// ConfigureStackExchange for SlackOverflow
func configureStackExchange(w *cli.Worker, so *internal.SlackOverflow) error {
	w.Log.Notice("Configuring Stack Exchange API Client")

	reader := bufio.NewReader(os.Stdin)

	if so.Config.StackExchange.SearchAdvanced == nil {
		so.Config.StackExchange.SearchAdvanced = make(map[string]string)
	}

	so.Config.StackExchange.Enable()
	so.Config.StackExchange.SetAPIhost("https://api.stackexchange.com")
	so.Config.StackExchange.SetAPIVersion("2.2")

	w.Log.Line("Name one Stack Exchange site where you want to track questions from e.g: stackoverflow.")
	w.Log.Line("For full list of available sites check: http://stackexchange.com/sites")
	site, _ := reader.ReadString('\n')
	so.Config.StackExchange.Site = strings.TrimSpace(site)

	w.Log.Line("Set tag which you want to track from selected site.")
	w.Log.Line("You can also set multible tags separated with (;) e.g: aframe;three.js")

	tagged, _ := reader.ReadString('\n')
	so.Config.StackExchange.SearchAdvanced["tagged"] = strings.TrimSpace(tagged)

	// Number of questions to watch
	w.Log.Line("Set the value for how many latest questions you want to track and update.")
	w.Log.Line("Good value is (25) which means that besides checking new qustions in defined stack exchange site")
	w.Log.Line("also last (n) questions will be checked for comment count, view count, answer count, score and is question accepted or not.")
	w.Log.Line("Emoijs of these stats will be removed from older than (n) questions.")
	fmt.Scan(&so.Config.StackExchange.QuestionsToWatch)

	w.Log.Line("Without having Stack Exchange API APP key's you can make 300 requests per day.")
	w.Log.Line("When you register for an Stack Exchange API App Key you can make 10000 requests per day")
	w.Log.Line("You can register for an APP KEY here: http://stackapps.com/apps/oauth/register")
	registerStackExchange := w.AskForConfirmation("Do you want to set keys now?")
	if registerStackExchange {
		w.Log.Line("Enter your Client Key")
		clientKey, _ := reader.ReadString('\n')
		so.Config.StackExchange.Key = strings.TrimSpace(clientKey)
	}
	return so.Config.Save()
}
