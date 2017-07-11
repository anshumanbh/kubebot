package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/pubsub"

	"golang.org/x/oauth2"

	dl "github.com/pmezard/go-difflib/difflib"
	git "gopkg.in/src-d/go-git.v4"

	"github.com/google/go-github/github"
)

var results []*pubsub.PublishResult
var diffstring string
var jval string

type error interface {
	Error() string
}

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

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func gitclone(repourl string, reponame string) (string, error) {

	Info("git clone " + repourl)
	buf := "/tmp/" + reponame

	_, err := git.PlainClone(buf, false, &git.CloneOptions{
		URL:      repourl,
		Progress: os.Stdout,
		// Auth:     gitssh.NewBasicAuth(os.Getenv("githubowner"), os.Getenv("gitpersonaltoken")),
	})

	CheckIfError(err)

	return buf, nil

}

func diff(downloadedfilepath string, newfilepath string) (string, error) {

	f1, err := ioutil.ReadFile(downloadedfilepath)
	CheckIfError(err)

	f2, err := ioutil.ReadFile(newfilepath)
	CheckIfError(err)

	diff := dl.UnifiedDiff{
		A:        dl.SplitLines(string(f1)),
		B:        dl.SplitLines(string(f2)),
		FromFile: "Original",
		ToFile:   "New",
		Context:  3,
		Eol:      "\n",
	}
	// result, err := dl.GetContextDiffString(diff)
	result, err := dl.GetUnifiedDiffString(diff)
	CheckIfError(err)

	return result, nil
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func overwrite(newfilepath string, downloadedfilepath string) {
	err1 := copyFileContents(newfilepath, downloadedfilepath)
	if err1 != nil {
		fmt.Printf("Unable to copy file%q\n", err1)
		os.Exit(1)
	}
	fmt.Printf("File contents successfully copied\n")
	return

}

func CheckGitError(err error) {

	if _, ok := err.(*github.RateLimitError); ok {
		fmt.Println("hit rate limit")
	}

	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)

}

func gitop(downloadedfilepath string, filename string, toolname string) {

	//AuthN to Github using the personal access token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("gitpersonaltoken")})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	//Retrieving the repo object. this will also ensure the repo exists. otherwise no bueno!
	repo, _, err := client.Repositories.Get(ctx, os.Getenv("githubowner"), toolname)
	CheckGitError(err)

	//Retrieving the reponame and ownername from the repo object
	reponame := *repo.Name
	ownername := *repo.Owner.Login

	//Specifying the options to retrieve file contents
	opt1 := &github.RepositoryContentGetOptions{
		Ref: "",
	}

	//Retrieving the SHA value of the file contents of the file
	//if no file exists, it will say so but continue with the flow and create a file
	filecontent, _, _, err := client.Repositories.GetContents(ctx, ownername, reponame, filename, opt1)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			fmt.Println("\nTrying to check the last updated file for its SHA value of the blob")
			fmt.Println("\nFile doesn't exist. Error: " + err.Error())
			fmt.Println("\nCreating a new file..")
		}
	}

	//Retrieving the SHA value of the file blob which is needed to update it
	SHAvalue := filecontent.GetSHA()
	fmt.Println("\nSHA Value of the Last Updated Blob: ", SHAvalue)

	//Now, reading the contents of the oldrun file that was updated with the contents of the newrun file. Storing it in data which is a byte array
	data, err := ioutil.ReadFile(downloadedfilepath)
	CheckIfError(err)

	//Specifying the options to update file contents
	opt2 := &github.RepositoryContentFileOptions{
		Message: github.String("Updating newrun file.."),
		Content: data,
		SHA:     github.String(SHAvalue),
	}

	//Updating the Github file now with the new contents
	newfile, _, err := client.Repositories.UpdateFile(ctx, ownername, reponame, filename, opt2)
	CheckGitError(err)

	Info("Now, pushing the udpated file to Github\n")

	fmt.Println("File created at " + *newfile.URL)

	return
}

func sendtoslack(data string, toolname string, target string) {

	if data == "" {
		jval = "The output from: " + toolname + " against: " + target + " has not changed since the last run"

	} else {

		jval = "*" + "Output from: " + toolname + " against: " + target + "*" + "\n" + "```" + data + "```"
	}

	jsonstr := map[string]string{"text": jval}
	jsonValue, _ := json.Marshal(jsonstr)

	body := strings.NewReader(string(jsonValue))

	req, err := http.NewRequest("POST", os.Getenv("slackwebhookurl"), body)
	CheckIfError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	CheckIfError(err)
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		fmt.Println("Message sent to Slack!")
		return
	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func main() {

	//getting the filepath as a command line argument
	filepath := os.Args[1]
	fmt.Println("Filepath: " + filepath)

	filename := strings.Split(filepath, "/")[3]
	fmt.Println("Filename: " + filename)

	toolname := strings.Split(filename, "_")[0]
	fmt.Println("Toolname: " + toolname)

	target := strings.Split(filename, "_")[1]
	fmt.Println("Target: " + target)

	fmt.Println("")

	//subdomain bruteforcing tools that don't sort the output. Ugh!!
	subdomainbruteforcingTools := []string{"gobuster", "subbrute", "sublist3r"}

	value := false
	fsize := int64(0)

	for (value == false) || (fsize == int64(0)) {
		i, s, err := exists(filepath)
		CheckIfError(err)
		value = i
		fsize = s
	}

	fmt.Println("Newrun File exists:", value)
	fmt.Println("Newrun File size:", fsize)
	fmt.Println("")
	//now, we know that the file exists and that its size>0

	// if the tool is one of the tools in the subdomainbruteforcingTools string slice, we need to sort, uniq it and store the file again
	if contains(subdomainbruteforcingTools, toolname) {

		Info("Since the tool does not sort its output, sorting and uniq'ing the output and storing it at the same filepath again")

		// sort and uniq the newrun file
		cmd := exec.Command("sort", "-u", filepath)

		// open a temp out file for writing
		outfile, err := ioutil.TempFile("", "")
		if err != nil {
			return
		}
		defer outfile.Close()
		cmd.Stdout = outfile

		if err = cmd.Start(); err != nil { //running the command which actually sorts it, uniq's it, and pastes the stdout output in outfile
			return
		}
		cmd.Wait()

		// We have the temp file now that's the sorted version of newrun file @ filepath so we need to delete the file @ filepath
		err1 := os.Remove(filepath)
		CheckIfError(err1)

		//Now, that the file @ filepath is gone, we need to copy the contents of the temp outfile to a file @ filepath
		overwrite(outfile.Name(), filepath)

		//We now have a sorted file @ filepath i.e. the newrun file is now sorted for tools that don't provide a sorted output

	}

	//forming the repourl to pass onto gitclone function
	githubacturl := os.Getenv("githubacturl")
	repourl := githubacturl + "/" + toolname

	//git clone repo
	tmppath, err := gitclone(repourl, toolname)
	CheckIfError(err)

	Info("Cloned to: " + tmppath)
	downloadedfilepath := tmppath + "/" + filename

	//check if the downloaded file exists in the cloned repo or not
	dvalue, dsize, err := exists(downloadedfilepath)
	CheckIfError(err)

	//checking to see if the file existed or not, also checking if filesize>0
	if (dvalue == true) && (dsize > int64(0)) {
		fmt.Println("\nLastRun File exists:", dvalue)
		fmt.Println("LastRun File size:", dsize)
		fmt.Println()

		//diff both files since it already existed
		diffstring, err = diff(downloadedfilepath, filepath)
		CheckIfError(err)
		fmt.Println(diffstring)

	} else if (dvalue == true) && (dsize == int64(0)) {
		fmt.Println("The file exists in the cloned repo but the size is", dsize)

	} else {
		fmt.Println("The file does not exist in the cloned repo")
	}

	//copy aka overwrite the old file by the new file contents
	Info("Overwriting the old file with the new file")
	overwrite(filepath, downloadedfilepath)

	//git add, commit & push the new file aka update the contents of the file on github
	gitop(downloadedfilepath, filename, toolname)

	if (dvalue == true) && (dsize > int64(0)) {
		// send the file to Slack
		Info("Since there was a valid file with size>0, sending only the diff to Slack")
		fmt.Println("\nSending Data: " + diffstring)

		sendtoslack(diffstring, toolname, target)

	} else {

		//getting the data from the file to send
		data, err := ioutil.ReadFile(downloadedfilepath)
		CheckIfError(err)

		// send the file to Slack
		Info("Since there was no valid file with size>0, sending the new file to Slack")
		fmt.Println("\nSending Data: " + string(data))

		sendtoslack(string(data), toolname, target)
	}

	fmt.Println("Done")

}
