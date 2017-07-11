# Integration of tools

Let's say we need to integrate another tool in Kubebot. How do we do that?

* We need to first ensure the tool can take an input and produce the results in an output file.
* Next, we need to build a Docker image of the tool. Take a look at the Dockerfile(s) in the `tools` folder for different tools.
* Next, we push the image to a Docker registry where it can be downloaded from into the K8s cluster.
* Finally, we add a switch case statement in the [subscriptionworker.go](../subscriptionworker/subscriptionworker.go) file in the `singlejobwithinitcontainer` function.
* Create a github repository for this tool to store the results.