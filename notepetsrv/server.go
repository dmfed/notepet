package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dmfed/notepet"
	"github.com/dmfed/notepet/storage"
)

const version string = `notepet v0.1.0
Copyright 2021 by Dmitry Fedotov
Redistributable under MIT license`

// ReadTokensFile accepts filename to parse. It reads
// file and adds each non-empty line found as token.
func readTokensFromFile(filename string) (tokens []string, err error) {
	data, err := os.ReadFile(filename)
	tokens = []string{}
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Trim(line, " \n")
		if line != "" {
			tokens = append(tokens, line)
		}
	}
	return
}

func main() {
	var (
		flagIPAddr      = flag.String("ip", "127.0.0.1", "ip address to listen on")
		flagPort        = flag.String("port", "10000", "port to listen on")
		flagTokensFile  = flag.String("tokens", "/usr/local/share/notepetsrv/tokens.conf", "tokens file to use")
		flagStorageFile = flag.String("storage", "/usr/local/share/notepetsrv/storage.json", "storage file to use")
		flagCertFile    = flag.String("cert", "", "certificate file to use")
		flagKeyFile     = flag.String("key", "", "key file to use")
		flagAppToken    = flag.String("t", "", "provide app token via command line")
		flagVersion     = flag.Bool("v", false, "print version and exit")
	)
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		return
	}

	// Open storage
	var st notepet.Storage
	st, err := storage.OpenJSONFileStorage(*flagStorageFile)
	if err != nil {
		log.Printf("could not open storage: %v exiting", err)
		return
	}

	// Get tokens
	var tokens = []string{}
	if *flagTokensFile != "" {
		if tks, err := readTokensFromFile(*flagTokensFile); err == nil {
			tokens = append(tokens, tks...)
		}
	}
	if *flagAppToken != "" {
		tokens = append(tokens, *flagAppToken)
	}

	// Configure the server
	srv, err := notepet.NewNotepetServer(*flagIPAddr, *flagPort, st, tokens...)
	if err != nil {
		st.Close()
		return
	}

	// Fire up the server
	if *flagCertFile != "" && *flagKeyFile != "" {
		log.Fatal(srv.ListenAndServeTLS(*flagCertFile, *flagKeyFile))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
	log.Println("Notepet server stopped")
}
