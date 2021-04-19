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
