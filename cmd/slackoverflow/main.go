// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/howi-ce/howi/addon/application"
	"github.com/howi-ce/howi/addon/application/plugin/cli/flags"
	"github.com/howi-ce/howi/std/log"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/commands"
	"github.com/mkungla/slackoverflow/cmd/slackoverflow/internal"
)

func main() {
	app := application.NewAddon()

	buildDate, err := time.Parse(time.RFC3339, internal.BuildDate)
	if err != nil {
		buildDate = time.Now().UTC()
	}
	// Set application info
	info := app.Info()
	info.SetName(internal.Name)
	info.SetShortDesc(internal.ShortDesc)
	info.SetCopyRightInfo(2016, "A-Frame authors")
	info.SetURL("https://github.com/aframevr/slackoverflow")
	info.SetVersion(internal.Version)
	info.SetBuildDate(buildDate)
	for _, contributor := range internal.Contributors {
		info.AddContributor(contributor)
	}

	// Configure CLI
	appcli := app.CLI()
	appcli.Log.Colors()
	appcli.Log.SetPrimaryColor("magenta")
	appcli.Log.SetLogLevel(log.NOTICE)

	// Application header
	appcli.Header.SetTemplate(internal.CLIheader)
	// Application footer
	appcli.Footer.SetTemplate(internal.CLIfooter)

	so := internal.NewSlackOverflow()

	// Flags
	cnfFlag := flags.NewStringFlag("config-path")
	cnfFlag.SetUsage("set alternative path  to configuration directory. defaults ~/.slackoverflow")
	appcli.AddFlag(cnfFlag)

	// Attach Commands
	appcli.AddCommand(commands.Config(so))
	appcli.AddCommand(commands.Reconfigure(so))
	appcli.AddCommand(commands.Run(so))
	appcli.AddCommand(commands.Service(so))
	appcli.AddCommand(commands.Slack(so))
	appcli.AddCommand(commands.StackExchange(so))
	appcli.AddCommand(commands.Validate(so))

	app.Start()
}
