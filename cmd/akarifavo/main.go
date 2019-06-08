package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kechako/akarifavo"
)

var appID string

func init() {
	flag.StringVar(&appID, "appid", os.Getenv("YAHOO_APP_ID"), "Yahoo! Application ID.")
}

func printHelp() {
	flag.Usage()
	flag.PrintDefaults()
}

func run() (int, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s text\n\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() != 1 {
		printHelp()
		return 2, nil
	}

	if appID == "" {
		fmt.Fprint(os.Stderr, "Yahoo! Application ID is not specified.\n\n")
		printHelp()
		return 2, nil
	}

	a := akarifavo.New(appID)

	s, err := a.Say(context.Background(), flag.Arg(0))
	if err != nil {
		return 1, err
	}
	if s != "" {
		fmt.Println(s)
	}

	return 0, nil
}

func main() {
	code, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error : %v\n", err)
	}
	os.Exit(code)
}
