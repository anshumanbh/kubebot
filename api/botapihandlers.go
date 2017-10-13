package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/gorilla/schema"
	"github.com/subosito/gotenv"
)

var r1 *pubsub.PublishResult
var decoder = schema.NewDecoder()

// Slackbot struct
type Slackbot struct {
	Channelid   string `schema:"channel_id"`
	Channelname string `schema:"channel_name"`
	Command     string `schema:"command"`
	Responseurl string `schema:"response_url"`
	Teamdomain  string `schema:"team_domain"`
	Teamid      string `schema:"team_id"`
	Text        string `schema:"text"`
	Token       string `schema:"token"`
	Userid      string `schema:"user_id"`
	Username    string `schema:"user_name"`
	Triggerid   string `schema:"trigger_id"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Index API func
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
	fmt.Fprint(w, "The Kubebot is going to be EPIC!\n")
	fmt.Fprint(w, "Please contact anshuman dot bhartiya at gmail dot com for more details.")
}

//Apitest POST and GET API v1 TEST function
func Apitest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "API v1 Test Successful\n")
}

//RunTool API func
func RunTool(w http.ResponseWriter, r *http.Request) {

	err2 := gotenv.Load("../.env")
	if err2 != nil {
		log.Fatal("Error loading .env file")
	}

	defer r.Body.Close()

	err := r.ParseForm()
	check(err)

	var slackbot Slackbot

	err1 := decoder.Decode(&slackbot, r.PostForm)
	check(err1)

	//Verify the request is originating from slack, belongs to the correct domain/team and the correct user is sending the request
	if (slackbot.Teamdomain == os.Getenv("slackteamdomain")) && (slackbot.Username == os.Getenv("slackusername") && (slackbot.Token == os.Getenv("slackverifytoken"))) {
		//Takes in the text and publishes message(s) against the target(s) in PubSub for the Subscription Workers to pick up
		bb := strings.Split(slackbot.Text, "|")
		tname, options, targets, responseurl := bb[0], bb[1], strings.Split(bb[2], ","), slackbot.Responseurl

		proj := os.Getenv("project_id")
		ctx := context.Background()

		psclient, err := pubsub.NewClient(ctx, proj)
		check(err)

		topicName := os.Getenv("tools_topicname")
		topic := psclient.Topic(topicName)

		for _, element := range targets {

			mapD := map[string]string{"toolname": tname, "target": element, "options": options, "responseurl": responseurl}
			mapB, _ := json.Marshal(mapD)

			r1 = topic.Publish(ctx, &pubsub.Message{
				Data: []byte(mapB),
			})

			id, err := r1.Get(ctx)
			check(err)

			fmt.Fprintf(w, "Job submitted to run "+tname+" with options "+options+" against "+element+". Published Message ID: "+id+" \n")

		}
	} else {
		fmt.Fprintf(w, "Make sure the Slack team domain, Username and Verify token are all correct...")
	}

}

//RunAutomation API func
func RunAutomation(w http.ResponseWriter, r *http.Request) {

	err2 := gotenv.Load("../.env")
	if err2 != nil {
		log.Fatal("Error loading .env file")
	}

	defer r.Body.Close()

	err := r.ParseForm()
	check(err)

	var slackbot Slackbot

	err1 := decoder.Decode(&slackbot, r.PostForm)
	check(err1)

	//Verify the request is originating from slack, belongs to the correct domain/team and the correct user is sending the request
	if (slackbot.Teamdomain == os.Getenv("slackteamdomain")) && (slackbot.Username == os.Getenv("slackusername") && (slackbot.Token == os.Getenv("slackverifytoken"))) {
		//Takes in the text and publishes message(s) against the target(s) in PubSub for the Subscription Workers to pick up
		bb := strings.Split(slackbot.Text, "|")
		tname, targets, responseurl := bb[0], strings.Split(bb[1], ","), slackbot.Responseurl

		proj := os.Getenv("project_id")
		ctx := context.Background()

		psclient, err := pubsub.NewClient(ctx, proj)
		check(err)

		topicName := os.Getenv("tools_topicname")
		topic := psclient.Topic(topicName)

		for _, element := range targets {

			mapD := map[string]string{"toolname": tname, "target": element, "responseurl": responseurl}
			mapB, _ := json.Marshal(mapD)

			r1 = topic.Publish(ctx, &pubsub.Message{
				Data: []byte(mapB),
			})

			id, err := r1.Get(ctx)
			check(err)

			fmt.Fprintf(w, "Job submitted to run "+tname+" against "+element+". Published Message ID: "+id+" \n")

		}
	} else {
		fmt.Fprintf(w, "Make sure the Slack team domain, Username and Verify token are all correct...")
	}

}
