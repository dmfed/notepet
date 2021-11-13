package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/dmfed/notepet"
	"github.com/dmfed/termtools"
)

const version string = `notepet v0.4.0
Copyright 2021 by Dmitry Fedotov
Redistributable under MIT license`

var prnt = termtools.PrintSuite{}

//Output detailed usage info
func displayHelpLong() { //TODO: write proper help
	name := os.Args[0]
	prnt.Printf(`Usage: %v <options> <command> <arguments>
  Commands are: show, put, new, sticky, del, edit, search, export
	
  Example: 
  Argument to get and del commands is index of Note to printout or delete
  show shows all notes and their indices
  search looks up the input string and returns matching results
  Example: 
	%v show returns all notes in storage. You can provide a single index
	   or slice. show 3 - will show note No 3. show 2: will show all notes
	   starting from 2. show :4 will show first four note inclusive.	
	%v put "Hello" "Hello world" "tag1 tag2" - adds note with title
	   "Hello", body "Hello world" and two tags. IF only one argument is 
	   present after put command it will be considered the body of note.
	%v del 1 - deletes note with index 1
  
  Options:
`, name, name, name, name)
	flag.PrintDefaults()
}

func main() {
	homedir, _ := os.UserHomeDir()
	configpath := filepath.Join(homedir, ".notepet.conf")
	var (
		flagConfigFile = flag.String("conf", configpath, "Config file to use")
		flagVerbose    = flag.Bool("v", false, "Verbose output")
		flagColor      = flag.Bool("c", false, "Colored output")
		flagVersion    = flag.Bool("version", false, "Print version and exit")
		flagIP         = flag.String("ip", "", "ip address to connect to")
		flagPort       = flag.String("port", "", "port to connect to")
		flagAPIpath    = flag.String("path", "", "api base path")
		// flagUpdateIDs  = flag.Bool("generate", false, "recalculate IDs of all notes")
	)
	flag.Usage = displayHelpLong
	flag.Parse()

	if *flagVersion {
		prnt.Println(version)
		os.Exit(0)
	}
	if len(os.Args) < 2 { //Will do nothing if no command provided
		displayHelpLong()
		os.Exit(0)
	}
	conf := readAndParseConfig(*flagConfigFile)
	if *flagVerbose {
		conf.verbose = *flagVerbose
	}
	if *flagColor {
		conf.color = *flagColor
	}
	if *flagIP != "" {
		conf.server = *flagIP
	}
	if *flagPort != "" {
		conf.port = *flagPort
	}
	if *flagAPIpath != "" {
		conf.path = *flagAPIpath
	}
	storage, err := notepet.NewAPIClient(conf.server, conf.port, conf.path, conf.token)
	if err != nil {
		prnt.Printf("error initializing api client: %v", err)
		return
	}
	defer storage.Close()

	if err := runCLI(storage, conf); err != nil {
		prnt.Println(err)
	}
}
