// Copyright © 2016 -2017 A-Frame authors.
// Use of this source code is governed by a MIT License
// that can be found in the LICENSE file.

package internal

import (
	"io/ioutil"
	"os"

	"github.com/howi-ce/howi/std/errors"
	"github.com/nlopes/slack"

	yaml "gopkg.in/yaml.v2"
)

// Config of SlackOverflow
type Config struct {
	loaded        bool
	file          string
	Slack         SlackConfig         `yaml:"slack"`
	StackExchange StackExchangeConfig `yaml:"stackexchange"`
}

// IsLoaded returns true if configuration has been loaded
func (c *Config) IsLoaded() bool {
	return c.loaded
}

// Save the configuration file
func (c *Config) Save() error {
	contents, _ := yaml.Marshal(&c)
	err := ioutil.WriteFile(c.file, []byte(contents), 0644)
	if err != nil {
		return errors.Newf("Failed to write: %s", c.file)
	}
	return nil
}

// readConfig - Read yaml into struct
func (c *Config) readConfig() error {
	f, err := os.Open(c.file)
	if err != nil {
		return err
	}
	defer f.Close()
	inputBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(inputBytes, &c); err != nil {
		return err
	}
	c.loaded = true
	return nil
}

// SlackConfig for Slack Overflow
type SlackConfig struct {
	Enabled     bool           `yaml:"enabled"`
	TeamURL     bool           `yaml:"team-url"`
	Channel     string         `yaml:"channel"`
	ChannelName string         `yaml:"channel-name"`
	Token       string         `yaml:"token"`
	APIHost     string         `yaml:"api-host"`
	TeamInfo    slack.TeamInfo `yaml:"team-info"`
}

// Enable posting and updating to Slack
func (s *SlackConfig) Enable() {
	s.Enabled = true
}

// Disable posting and updating to Slack
func (s *SlackConfig) Disable() {
	s.Enabled = false
}

// SetAPIhost for Slack API
func (s *SlackConfig) SetAPIhost(h string) {
	s.APIHost = h
}

// SetToken sets Slack API token
func (s *SlackConfig) SetToken(t string) {
	s.Token = t
}

// SetChannel sets the slack channel where questions will be posted
func (s *SlackConfig) SetChannel(ch string) {
	s.Channel = ch
}

// SetChannelName sets the name of the slack channel where questions will be posted
func (s *SlackConfig) SetChannelName(n string) {
	s.ChannelName = n
}

// SetTeamInfo for current Slack configuration
func (s *SlackConfig) SetTeamInfo(t *slack.TeamInfo) {
	s.TeamInfo = *t
}

// StackExchangeConfig for Slack Overflow
type StackExchangeConfig struct {
	Enabled          bool              `yaml:"enabled"`
	Key              string            `yaml:"key"`
	APIVersion       string            `yaml:"api-version"`
	APIHost          string            `yaml:"api-host"`
	Site             string            `yaml:"site"`
	QuestionsToWatch int               `yaml:"questions-to-watch"`
	SearchAdvanced   map[string]string `yaml:"search-advanced"`
	Questions        map[string]string `yaml:"questions"`
}

// Enable Stack Exchange
func (s *StackExchangeConfig) Enable() {
	s.Enabled = true
}

// Disable Stack Exchange
func (s *StackExchangeConfig) Disable() {
	s.Enabled = false
}

// SetAPIhost for Stack Exchange API
func (s *StackExchangeConfig) SetAPIhost(h string) {
	s.APIHost = h
}

// SetAPIVersion for Stack Exchange API
func (s *StackExchangeConfig) SetAPIVersion(v string) {
	s.APIVersion = v
}
