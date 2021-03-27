package main

import (
	"github.com/dmfed/conf"
)

type notepetConfig struct {
	editor  string
	verbose bool
	color   bool
	server  string
	port    string
	token   string
}

// ReadAndParseConfig returns pointer to notepetConfig
// It takes path (including filename) as argument and tries to read
// configuration file. If it fails - it fall back to default
// configuration
func readAndParseConfig(filename string) *notepetConfig {
	config := notepetConfig{editor: "nano"}
	parsed, err := conf.ParseFile(filename)
	if err != nil {
		return &config
	}
	if ed, err := parsed.Find("editor"); err == nil {
		config.editor = ed.String()
	}
	config.server = parsed.Get("server").String()
	config.port = parsed.Get("port").String()
	config.token = parsed.Get("token").String()
	config.verbose = parsed.HasOption("verbose")
	config.color = parsed.HasOption("color")
	return &config
}
