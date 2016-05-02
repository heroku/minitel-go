package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/heroku/minitel-go"
)

var targetID = flag.String("target", "", "Target UUID")
var followupID = flag.String("followup", "", "followup")
var argType = flag.String("type", "", "Target Type (app, user)")
var title = flag.String("title", "Default Title", "Title")

func getBody() string {
	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func getClient() minitel.Client {
	url := os.Getenv("TELEX_URL")
	pos := flag.Args()
	if len(pos) > 0 {
		url = pos[0]
	}

	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	client, err := minitel.New(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid telex URL: %q", err)
		os.Exit(1)
	}
	return client
}

func post() {
	var targetType minitel.Type
	if *targetID == "" || *argType == "" {
		flag.Usage()
		os.Exit(1)
	} else {
		switch *argType {
		case "app":
			targetType = minitel.App
		case "user":
			targetType = minitel.User
		default:
			fmt.Fprintf(os.Stderr, "unknown target type: %q", *argType)
			os.Exit(1)
		}
	}

	payload := minitel.Payload{
		Title: *title,
		Body:  getBody(),
	}
	payload.Target.Id = *targetID
	payload.Target.Type = targetType

	res, err := getClient().Notify(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "received the following error: %q", err)
		os.Exit(1)
	}
	fmt.Printf("Posted message. ID=%q", res.Id)
}

func followup() {
	res, err := getClient().Followup(*followupID, getBody())
	if err != nil {
		fmt.Fprintf(os.Stderr, "received the following error: %q", err)
		os.Exit(1)
	}
	fmt.Printf("Posted followup message to %q. ID=%q", *followupID, res.Id)
	return
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <telex URL>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *followupID != "" {
		followup()
	} else {
		post()
	}

}
