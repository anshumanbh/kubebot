# TODO List

* Move from using public Github repos to private for diffing the files.
* Test cases. I should have done this from the get-go but oh well!!
* Automation of multiple tools together using Goroutines, waitgroups and channels. Check [this](https://abronan.com/introduction-to-goroutines-and-go-channels/).
* Init containers are out of beta so need to change some code in subscription worker.
* Figure out what to do with long outputs being sent to Slack. Slack doesn't like those.
* Explore the options parallelism and completions in job spec of the K8s pods/containers.