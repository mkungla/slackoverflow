// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

var (
	// Name of the application
	Name = "slackoverflow"
	// Version number
	Version = "17.11.10"
	// ShortDesc of the app
	ShortDesc = "Slackoverflow enables you to post tagged Stack Overflow questions to Slack, updated using reaction emojis."
	// BuildDate of the application
	BuildDate = "2017-11-10T01:18:26+02:00"
	// Contributors of the application
	Contributors = []string{
		"Marko Kungla <marko.kungla@gmail.com>",
	}
	// CLIheader template
	CLIheader = `
################################################################################
# {{ .Title }}{{ if .CopyRight}}
#  {{ .CopyRight }}{{end}}
# {{if .Version}}
#   Version:    {{ .Version }}{{end}}{{if .BuildDate}}
#   Build date: {{ .BuildDate | funcDate }}{{end}}
################################################################################
`
	// CLIfooter template
	CLIfooter = `
################################################################################
# elapsed: {{ funcElapsed }}
################################################################################`
)
