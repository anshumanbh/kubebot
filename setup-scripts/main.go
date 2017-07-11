package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"

	"github.com/google/go-github/github"
)

var (
	gac     = flag.String("gac", "", "Location of the Google Application Credentials file")
	project = flag.String("project", "", "GCP Project ID")

	subscription = flag.String("subscription", "", "PubSub Subscription Name")
	topic        = flag.String("topic", "", "PubSub Topic Name")

	wfdataset = flag.String("wfdataset", "", "WFUZZ BigQuery Dataset Name")
	rsdataset = flag.String("rsdataset", "", "Repo-Supervisor BigQuery Dataset Name")
	wftable   = flag.String("wftable", "", "WFUZZ BigQuery Table Name")
	rstable   = flag.String("rstable", "", "Repo-Supervisor BigQuery Table Name")

	tool     = flag.String("tool", "", "Toolname")
	gittoken = flag.String("gittoken", "", "Github Personal Access Token")
)

func check(e error) {
	if e != nil {
		panic(e)
	} else if _, ok := e.(*github.RateLimitError); ok {
		log.Println("hit rate limit")
	} else if _, ok := e.(*github.AcceptedError); ok {
		log.Println("scheduled on GitHub side")
	}
}

// Info Function to show colored text
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func createBQ(ctx context.Context, ds string, tb string, schema bigquery.Schema, proj string) error {

	//check if dataset exists. if not, create it. also create the table
	//if it exists, check if table exists. if not, create it
	//if it exists, delete and create it with the schema again

	bqClient, err := bigquery.NewClient(ctx, proj)
	check(err)

	err = bqClient.Dataset(ds).Create(ctx)
	if err != nil {
		fmt.Println(ds + " already exists")
		err = bqClient.Dataset(ds).Table(tb).Create(ctx, schema)
		if err != nil {
			fmt.Println(tb + " already exists. Deleting it now..")
			err = bqClient.Dataset(ds).Table(tb).Delete(ctx)
			check(err)
			fmt.Println(tb + " deleted. Creating it again now..")
			err = bqClient.Dataset(ds).Table(tb).Create(ctx, schema)
			check(err)
			fmt.Println(tb + " recreated")
		} else if err == nil {
			fmt.Println(tb + " created")
		}
	} else if err == nil {
		fmt.Println(ds + " created")
		err = bqClient.Dataset(ds).Table(tb).Create(ctx, schema)
		check(err)
		fmt.Println(tb + " created")
	}
	return nil
}

func main() {
	//Parsing the flags
	flag.Parse()

	ctx := context.Background()

	if (*gac == "" || *project == "") && (*gittoken == "" || *tool == "") {
		fmt.Println("Need the GAC file location along with the GCP Project ID OR Need the Git token along with the tool name")
		os.Exit(1)
	} else {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *gac)
		//check if sub exists. delete it
		//check if topic exists. delete it
		//create both the topic and sub

		if *topic != "" && *subscription != "" {

			psClient, err := pubsub.NewClient(ctx, *project)
			check(err)

			sub := psClient.Subscription(*subscription)
			subExists, err := sub.Exists(ctx)
			check(err)

			if subExists {
				fmt.Println(*subscription + " Exists. Deleting it now..")
				err := sub.Delete(ctx)
				check(err)
				fmt.Println(*subscription + " deleted")
			}

			top := psClient.Topic(*topic)
			topicExists, err := top.Exists(ctx)
			check(err)

			if topicExists {
				fmt.Println(*topic + " Exists. Deleting it now..")
				err := top.Delete(ctx)
				check(err)
				fmt.Println(*topic + " deleted")
			}

			newTopic, err := psClient.CreateTopic(ctx, *topic)
			check(err)

			newtopicExists, err := newTopic.Exists(ctx)
			check(err)

			if newtopicExists {
				fmt.Println(*topic + " created")
			}

			newSub, err := psClient.CreateSubscription(ctx, *subscription, newTopic, 0, nil)
			check(err)

			newsubExists, err := newSub.Exists(ctx)
			check(err)

			if newsubExists {
				fmt.Println(*subscription + " created")
			}
		}

		if (*wfdataset != "" && *wftable != "") || (*rsdataset != "" && *rstable != "") {

			wfschema := bigquery.Schema{
				&bigquery.FieldSchema{Name: "ID", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Response", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Lines", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Word", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Chars", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Request", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Success", Required: false, Type: bigquery.StringFieldType},
			}

			rsschema := bigquery.Schema{
				&bigquery.FieldSchema{Name: "File", Required: false, Type: bigquery.StringFieldType},
				&bigquery.FieldSchema{Name: "Secret", Required: false, Type: bigquery.StringFieldType},
			}

			var schema bigquery.Schema
			var ds, tb string

			if *wfdataset != "" && *wftable != "" {
				schema = wfschema
				ds = *wfdataset
				tb = *wftable
				err := createBQ(ctx, ds, tb, schema, *project)
				check(err)
			}

			if *rsdataset != "" && *rstable != "" {
				schema = rsschema
				ds = *rsdataset
				tb = *rstable
				err := createBQ(ctx, ds, tb, schema, *project)
				check(err)
			}

		}

		//for all tools ready, check if repo exists. if it exists, dont do anything
		//if it does not exist, create it with a License

		if *tool != "" && *gittoken != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *gittoken})
			tc := oauth2.NewClient(ctx, ts)
			gitClient := github.NewClient(tc)

			repo := &github.Repository{
				Name:            github.String(*tool),
				Description:     github.String("Kubebot Tool Repo - " + *tool),
				Private:         github.Bool(false),
				LicenseTemplate: github.String("mit"),
			}

			// Creating a repo
			repocreate, _, err := gitClient.Repositories.Create(ctx, "", repo)
			if _, ok := err.(*github.RateLimitError); ok {
				fmt.Println("hit rate limit")
			} else if err != nil {
				fmt.Println(*tool + " repository already exists")
			} else if err == nil {
				fmt.Println(*repocreate.Name + " repo created")
			}
		}
	}

}
