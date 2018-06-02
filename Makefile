VERSION = 0.1.0
PROJECT_ID = kubebot-163519
TOOLLIST = enumall gitallsecrets gitsecrets gobuster nmap subbrute sublist3r trufflehog wfuzz
UTILSLIST = checkfile converttobq wfuzzbasicauthbrute
CREDS_FILEPATH = /Users/redteam/Downloads/personal-creds.json
TOPIC = tool_topic
SUBSCRIPTION = tool_sub
WFUZZ_DATASET = wfuzzds
REPOSUPERVISOR_DATASET = reposupervisords
WFUZZ_TABLE = wfuzz_tomcat_test
REPOSUPERVISOR_TABLE = reposupervisor_test
GITTOKEN = <enter your github token>

.PHONY: all build

all: build

build: setup images deployments

setup:
	go run setup-scripts/main.go -project $(PROJECT_ID) -gac $(CREDS_FILEPATH) -wfdataset $(WFUZZ_DATASET) -wftable $(WFUZZ_TABLE) -rsdataset $(REPOSUPERVISOR_DATASET) -rstable $(REPOSUPERVISOR_TABLE) -topic $(TOPIC) -subscription $(SUBSCRIPTION)
	
images:
	docker build -t us.gcr.io/$(PROJECT_ID)/api/api_kubebot:$(VERSION) api/
	docker build -t us.gcr.io/$(PROJECT_ID)/api/api_subscriptionworker:$(VERSION) subscriptionworker/
	for toolname in $(TOOLLIST)  ; do \
		docker build -t us.gcr.io/$(PROJECT_ID)/tools/tools_$$toolname:$(VERSION) tools/$$toolname/ ; \
    done
	for utilname in $(UTILSLIST)  ; do \
		docker build -t us.gcr.io/$(PROJECT_ID)/utils/utils_$$utilname:$(VERSION) utils/$$utilname/ ; \
	done

deployments:
	kubectl apply -f config/kubebot-config/00-namespace.yaml
	kubectl create secret generic googlesecret --from-file=$(CREDS_FILEPATH) --namespace=kubebot-server
