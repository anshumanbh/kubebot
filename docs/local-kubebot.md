# How to run Kubebot locally for development and testing?


## out-of-cluster config using Minikube (This is where you want to start ideally!)
Please watch [this](https://youtu.be/1XkWmT6HxMo) video and follow along to install Kubebot locally.

* Git clone this repository.

* Start minikube by typing `minikube start`

* Once minikube starts, type `eval $(minikube docker-env)`. You are now inside the minikube's Docker environment i.e. you are essentially inside a single node K8s cluster running locally which is nothing but a replica of what it would look like when you deploy the infrastructure remotely. But, this is much easier for development and testing since everything is local for now.

* Change the value of `PROJECT_ID`, `CREDS_FILEPATH` and `GITTOKEN` in the `Makefile` to your GCP Project ID, the location where you store the GCP Default Compute Engine Service Account's credentials file and your Personal Access Token of your Github account respectively.

* Rename the `.gitrobrc.sample` file in the `tools/gitrob/client` and `tools/gitrob/server` directories to `.gitrobrc` and make sure you add your environment variables correctly.

* Type `make build`. This will take a few minutes so please be patient. This will:
    * Install `dep`. Ensure we have all the libraries and the correct versions we need.
    * Creates the appropriate Google PubSub topic and subscription, Google BigQuery datasets and tables and Github repositories for all the tools.
    * Build the Docker images for the API server, Subscription Worker, all the Utilities and all the tools. If, for some reason, an image does not get built, please build it manually. You can view the images by typing `docker images`.
    * Create a namespace `kubebot-server` inside the K8s cluster.
    * Create a secret `googlesecret` that has the value of the GCP Default Compute Engine's Service Account credentials file. We need this for the utilities to interact with our Google PubSub Service.

* After the above command successfully completes, type `docker images` and the output should look like below with all the images built and ready to go:
![Docker Images](/imgs/docker_images.png)

* Start a NGROK tunnel locally by typing `./ngrok http 3636`. Store the Forwarding https URL shown on the screen.

* Navigate to `https://api.slack.com/apps` and for your Slack account, create an app by configuring a Slash command named `runtool` with the `Request URL` parameter having the value `$URL/api/v1/runtool` where `$URL` is the NGROK URL that you obtained above. Save this Slash command. Make sure you can call this slash command from a channel inside your Slack account. Also, grab the Slack verification token for this slash command. You will need this in the `.env` file below.

* Similarly, set one for a command named `runautomation` with the `Request URL` parameter having the value `$URL/api/v1/runautomation` where `$URL` is the NGROK URL that you obtained above.

![Slack Image](/imgs/slack.png)

* Rename the `.env.sample` file to `.env` and make sure you add your environment variables correctly.

* Open two terminals side by side and make sure you type in `eval $(minikube docker-env)` in both of them to ensure you are inside the Minikube's Docker environment in both the terminals.

* In one terminal, navigate to the `api` directory and type `go run *.go`

* In the other terminal, navigate to the `subscriptionworker` directory and type `go run subscriptionworker.go`

* Go to your Slack account in the channel where you can run the slash command and type `/runtool nmap|-Pn -p 1-1000|google.com` and wait for a few seconds. See the magic happen!!

![Kubebot Slack Image](/imgs/kubebot_slack.png)

* So what happened?
    * The Slash command from Slack was sent to the https URL you configured.
    * That basically sends it to the NGROK tunnel you have setup locally to hit your `localhost:3636` which is nothing but the API server running.
    * The API server receives the request and sees that it needs to run `nmap` against `google.com` with the options `-Pn -p 1-1000`.
    * The API server starts a nmap job inside a Docker container (on the Minikube K8s cluster running locally) providing it the necessary arguments.
    * When the job finishes, it goes to your Github account and sees if that scan was previously performed. If it was, it diffs the present run against the previous run and sends only the diff to your Slack channel via a webhook. If it wasn't previously run, then it stores the output on Github under the `nmap` repository and returns the output to your Slack channel via a webhook.
