VERSION = 0.1.0
PROJECT_ID = kubebot-163519
TOOLLIST = enumall gitallsecrets gitrob gitsecrets gobuster nmap subbrute sublist3r trufflehog wfuzz
UTILSLIST = checkfile converttobq wfuzzbasicauthbrute
CREDS_FILEPATH = /Users/abhartiya/Desktop/personal-creds.json
TOPIC = tool_topic
SUBSCRIPTION = tool_sub
WFUZZ_DATASET = wfuzzds
REPOSUPERVISOR_DATASET = reposupervisords
WFUZZ_TABLE = wfuzz_tomcat_test
REPOSUPERVISOR_TABLE = reposupervisor_test
GITTOKEN =

.PHONY: all build

all: build

build: setup images deployments

setup:
	go get github.com/golang/dep
	go install github.com/golang/dep/cmd/dep
	cd api/ && dep init && dep ensure k8s.io/client-go@^2.0.0
	cd subscriptionworker/ && dep init && dep ensure k8s.io/client-go@^2.0.0
	cd utils/ && dep init
	go run setup-scripts/main.go -project $(PROJECT_ID) -gac $(CREDS_FILEPATH) -wfdataset $(WFUZZ_DATASET) -wftable $(WFUZZ_TABLE) -rsdataset $(REPOSUPERVISOR_DATASET) -rstable $(REPOSUPERVISOR_TABLE) -topic $(TOPIC) -subscription $(SUBSCRIPTION)
	for toolname in $(TOOLLIST)  ; do \
		if test $$toolname != gitrob ; then \
			go run setup-scripts/main.go -gittoken $(GITTOKEN) -tool $$toolname ; \
		fi ; \
    done

images:
	docker build -t us.gcr.io/$(PROJECT_ID)/api/api_kubebot:$(VERSION) api/
	docker build -t us.gcr.io/$(PROJECT_ID)/api/api_subscriptionworker:$(VERSION) subscriptionworker/
	for toolname in $(TOOLLIST)  ; do \
		if test $$toolname = gitrob; then \
			docker build -t us.gcr.io/$(PROJECT_ID)/tools/tools_gitrob_server:$(VERSION) tools/$$toolname/server/ ; \
			docker build -t us.gcr.io/$(PROJECT_ID)/tools/tools_gitrob:$(VERSION) tools/$$toolname/client/ ; \
		else \
			docker build -t us.gcr.io/$(PROJECT_ID)/tools/tools_$$toolname:$(VERSION) tools/$$toolname/ ; \
		fi; \
    done
	for utilname in $(UTILSLIST)  ; do \
		docker build -t us.gcr.io/$(PROJECT_ID)/utils/utils_$$utilname:$(VERSION) utils/$$utilname/ ; \
	done

deployments:
	kubectl apply -f config/kubebot-config/00-namespace.yaml
	kubectl create secret generic googlesecret --from-file=$(CREDS_FILEPATH) --namespace=kubebot-server