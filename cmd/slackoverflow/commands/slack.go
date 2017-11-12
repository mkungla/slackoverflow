// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package commands

import (
	"fmt"

	"github.com/howi-ce/howi/addon/application/plugin/cli"
	"github.com/howi-ce/howi/addon/application/plugin/cli/flags"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
	"github.com/nlopes/slack"
)

const (
	msgNotAnswered = "#B7E0ED"
	msgIsAnswewed  = "#30AC1F"
	thumbUp        = ":+1:"
	thumbDown      = ":-1:"
)

// Slack command for SlackOverflow.
func Slack(so *internal.SlackOverflow) cli.Command {
	cmd := cli.NewCommand("slack")
	cmd.SetShortDesc("Slack related commands see slackoverflow slack --help for more info.")
	cmd.AddSubcommand(SlackChannels(so))
	cmd.AddSubcommand(SlackQuestions(so))
	return cmd
}

// SlackChannels returns Slack channels command
func SlackChannels(so *internal.SlackOverflow) cli.Command {
	scmd := cli.NewCommand("channels")
	scmd.SetShortDesc("This command returns a list of all Slack channels in the team.")
	scmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		api := slack.New(so.Config.Slack.Token)
		channels, err := api.GetChannels(true)
		if err != nil {
			w.Fail(err.Error())
			return
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
			w.Fail("Unable to fetch any channels with provided credentials")
			return
		}
	})
	return scmd
}

// SlackQuestions returns Slack channels command
func SlackQuestions(so *internal.SlackOverflow) cli.Command {
	scmd := cli.NewCommand("questions")
	scmd.SetShortDesc("Post new or update tracked Stack Exchange questions on Slack channel.")

	pnqFlag := flags.NewBoolFlag("post-new")
	pnqFlag.SetUsage("Post new questions origin from configured Stack Exchange Site")
	scmd.AddFlag(pnqFlag)

	uqpnqFlag := flags.NewBoolFlag("update-old")
	uqpnqFlag.SetUsage("Update information about questions already posted to slack")
	scmd.AddFlag(uqpnqFlag)

	sFlag := flags.NewBoolFlag("sync-slack")
	sFlag.SetUsage("Get new questions and update information about existing questions")
	scmd.AddFlag(sFlag)

	scmd.Do(func(w *cli.Worker) {
		if err := so.Session(w); err != nil {
			w.Fail(err.Error())
			return
		}
		post, err := w.Flag("post-new")
		if err != nil {
			w.Fail(err.Error())
			return
		}

		update, err := w.Flag("update-old")
		if err != nil {
			w.Fail(err.Error())
			return
		}
		sync, err := w.Flag("sync-slack")
		if err != nil {
			w.Fail(err.Error())
			return
		}

		if sync.Present() {
			slackPostNewQuestions(w, so)
			slackUpdateQuestions(w, so)
		} else if post.Present() {
			slackPostNewQuestions(w, so)
		} else if update.Present() {
			slackUpdateQuestions(w, so)
		} else {
			w.Fail("atleast one flag must be provided")
			return
		}
	})
	return scmd
}
func slackPostNewQuestions(w *cli.Worker, so *internal.SlackOverflow) {
	w.Log.Info("Slack: Posting new questions.")

	tracked, questionCount := so.DB.StackExchangeQuestionsTracked(
		so.Config.StackExchange.QuestionsToWatch)
	if questionCount == 0 {
		w.Log.Notice("Slack: There are no questions in database,")
		return
	}

	// Process questions
	for _, question := range tracked {
		slackQuestion := so.DB.FindSlackQuestion(question.QID)
		if slackQuestion.QID == 0 {
			user := so.DB.FindStackExchangeUser(question.UID)
			params := slack.NewPostMessageParameters()
			params.Parse = "full"
			params.LinkNames = 1
			params.UnfurlLinks = true
			params.UnfurlMedia = false
			params.Username = fmt.Sprintf("%s asked on %s:",
				user.DisplayName,
				question.Site,
			)
			params.AsUser = false
			params.IconURL = user.ProfileImage
			params.Markdown = true
			params.EscapeText = true

			color := msgNotAnswered
			if question.IsAnswered {
				color = msgIsAnswewed
			}
			thumb := thumbUp
			if question.Score < 0 {
				thumb = thumbDown
			}
			ficon := "https://aframe.io/images/aframe-logo-192.png"
			if val, ok := so.Config.Slack.TeamInfo.Icon["image_132"].(string); ok {
				ficon = val
			}
			attachment := slack.Attachment{
				Fallback:  question.Title,
				Title:     question.Title,
				TitleLink: question.ShareLink,
				Color:     color,
				Text: fmt.Sprintf(":pencil: %d :speech_balloon: %d %s %d :eye: %d",
					question.AnswerCount,
					question.CommentCount,
					thumb,
					question.Score,
					question.ViewCount,
				),
				Footer:     "slackoverflow",
				FooterIcon: ficon,
			}
			params.Attachments = []slack.Attachment{attachment}
			api := slack.New(so.Config.Slack.Token)
			channelID, timestamp, err := api.PostMessage(so.Config.Slack.Channel, "", params)
			if err != nil {
				w.Log.Error(err.Error())
				return
			}
			// Store the link
			slackQuestion.QID = question.QID
			slackQuestion.Channel = channelID
			slackQuestion.TS = timestamp
			msg, err := so.DB.SlackQuestionCreate(slackQuestion)
			if err != nil {
				w.Log.Errorf("Slack channel (%s): %s %s", channelID, msg, err.Error())
			} else {
				w.Log.Infof("Slack channel (%s): %s and question posted", channelID, msg)
			}
		} else {
			w.Log.Debugf("Slack: Question %d already exists", question.QID)
		}
	}
	w.Log.Debug("No more new questions to post")
}
func slackUpdateQuestions(w *cli.Worker, so *internal.SlackOverflow) {
	w.Log.Info("Slack: Updating questions.")

	links, count := so.DB.SlackQuestionGetAll()
	if count == 0 {
		w.Log.Debug("No questions to update.")
		return
	}

	track := 0
	for _, ql := range links {
		stackQuestion := so.DB.FindStackExchangeQuestion(ql.QID)
		if stackQuestion.QID == 0 {
			w.Log.Warning("Could not find question with ID: %d.", ql.QID)
			continue
		}
		track++
		if track <= so.Config.StackExchange.QuestionsToWatch {
			color := msgNotAnswered
			if stackQuestion.IsAnswered {
				color = msgIsAnswewed
			}
			thumb := thumbUp
			if stackQuestion.Score < 0 {
				thumb = thumbDown
			}
			ficon := "https://aframe.io/images/aframe-logo-192.png"
			if val, ok := so.Config.Slack.TeamInfo.Icon["image_132"].(string); ok {
				ficon = val
			}
			attachment := slack.Attachment{
				Fallback:  stackQuestion.Title,
				Title:     stackQuestion.Title,
				TitleLink: stackQuestion.ShareLink,
				Color:     color,
				Text: fmt.Sprintf(":pencil: %d :speech_balloon: %d %s %d :eye: %d",
					stackQuestion.AnswerCount,
					stackQuestion.CommentCount,
					thumb,
					stackQuestion.Score,
					stackQuestion.ViewCount,
				),
				Footer:     "slackoverflow",
				FooterIcon: ficon,
			}

			api := slack.New(so.Config.Slack.Token)
			channelID, _, _, err := api.SendMessage(so.Config.Slack.Channel,
				slack.MsgOptionUpdate(ql.TS),
				slack.MsgOptionAsUser(false),
				slack.MsgOptionAttachments(attachment),
			)
			if err != nil {
				w.Log.Errorf("Slack channel (%s): %s", channelID, err.Error())
			} else {
				w.Log.Infof("Slack channel (%s) updated: %s", channelID, stackQuestion.Title)
			}

		} else {
			color := msgNotAnswered
			if stackQuestion.IsAnswered {
				color = msgIsAnswewed
			}
			attachment := slack.Attachment{
				Fallback:  stackQuestion.Title,
				Title:     stackQuestion.Title,
				TitleLink: stackQuestion.ShareLink,
				Color:     color,
			}
			api := slack.New(so.Config.Slack.Token)
			channelID, _, _, err := api.SendMessage(so.Config.Slack.Channel,
				slack.MsgOptionUpdate(ql.TS),
				slack.MsgOptionAsUser(false),
				slack.MsgOptionAttachments(attachment),
			)
			so.DB.StackExchangeQuestionDelete(stackQuestion)
			so.DB.SlackQuestionDelete(ql)
			if err != nil {
				w.Log.Errorf("Slack channel (%s): %s", channelID, err.Error())
			} else {
				w.Log.Infof("Quesstion: not tracking anymore. %s", stackQuestion.Title)
			}
		}
	}
}
