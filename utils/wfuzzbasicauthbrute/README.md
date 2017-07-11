# Basic Authentication Bruteforcing of WFUZZ endpoints with secrets obtained from Repo-Supervisor

## Running Locally
* Complete the `.env.sample` file in the `wfuzzbasicauthbrute` folder with the appropriate values and rename it to `.env`.

* Now, in order to bruteforce the basic authentication mechanism with the data retrieved from `wfuzz` and `repo-supervisor`, type `go run bruteforce.go -target <> -slackHook <>`