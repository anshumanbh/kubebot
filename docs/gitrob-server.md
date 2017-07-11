# Gitrob Server

* We need to start the Gitrob server before we could use the gitrob Slack slash command.
* In order to do that, navigate to the `tools/gitrob` directory, rename `.gitrobrc.sample` file to `.gitrobrc` after replacing the github access token value.
* Build the gitrob Docker image by typing `docker build -t us.gcr.io/$PROJECT_ID/tools/tools_gitrob_server:0.1.0 .`
* Start the Gitrob server in a pod by typing `kubectl apply -f pod-server.yaml`. This will create the server pod and the service.
