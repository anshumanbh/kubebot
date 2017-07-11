# Sample Automation Workflow

## WFUZZ Basic Authentication Bruteforcing
### What does it do?
* This workflow combines running the tools `wfuzz` and `git-all-secrets` simultaneously.
* And, then once they finish, we have some bruteforced endpoints that are potentially vulnerable and we also have some secrets we have obtained by searching the target's Github repositories.
* It then runs a bruteforce attack against the basic authentication mechanism of the target against each endpoint with all the secrets obtained above.
* Finally, it sends back the results to Slack.

### How does it work?
* Initiate a request from Slack by typing a command like `/runautomation wfuzzbasicauthbrute|<www.target.com>`
* API server receives the request
* API server drops a message in the queue to start `wfuzzbasicauthbrute` tool
* The message is picked up by a subscription worker from the queue
* Subscription worker starts 2 GoRoutines:
    * First GoRoutine starts [gitallsecrets](https://github.com/anshumanbh/git-all-secrets) with the options `-token <> -org <target> -toolName repo-supervisor -output /tmp/results/results.json`. As soon as this is finished, the results are uploaded to BigQuery in the table `reposupervisor_test` under the dataset `reposupervisords` by the help of a utility [converttobq](https://hub.docker.com/r/abhartiya/utils_converttobq/)
    * Second GoRoutine starts [wfuzz](https://github.com/anshumanbh/wfuzz) with the options `-w /data/SecLists/Discovery/Web_Content/tomcat.txt --hc 404,429,400 -o csv http://<TARGET>/FUZZ/ /tmp/results/results.csv`. As soon as this is finished, the results are uploaded to BigQuery in the table `wfuzz_tomcat_test` under the dataset `wfuzzds` by the help of a utility [converttobq](https://hub.docker.com/r/abhartiya/utils_converttobq/)
    * All the above jobs are performed inside Docker containers and they are destroyed once they are all completed.
* After the tools finish running above, the subscription worker starts another GoRoutine:
    * This GoRoutine starts a utility [wfuzzbasicauthbrute](https://hub.docker.com/r/abhartiya/utils_wfuzzbasicauthbrute/) with the opttions `-target <target> -slackHook <slackhook>`.
    * This utility basically fetches all the secrets obtained from the `reposupervisor_test` table and stores it in a file. It then fetches all the endpoints obtained from the table `wfuzz_tomcat_test`.
    * For each endpoint (ENDPOINT) retrieved above, the utility does a bruteforce attack against the basic authentication mechanism with all the secrets retrieved above against the URL `http://TARGET/ENDPOINT`. This is done by using the tool [wfuzz](https://github.com/xmendez/wfuzz) with the options `./wfuzz.py -w <all-the-secrets-file> -o csv --basic "admin:FUZZ" --sc 200,403 http://TARGET/ENDPOINT`
    * Finally, for each response with a `200` or `403` status, indicating that the secret worked against that endpoint, the results are sent back to Slack via the incoming Slack webhook.