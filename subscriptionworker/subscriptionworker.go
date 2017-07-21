package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"golang.org/x/net/context"

	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/pkg/api/v1"
	jobv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/tools/clientcmd"

	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/subosito/gotenv"
	"google.golang.org/grpc/grpclog"
)

var (
	subscription     *pubsub.Subscription
	qmsg             Message
	dockertoolurl    string
	dockerutilurl    string
	resultsmountpath string
	runtooljob       *jobv1.Job
)

// init function to get rid of weird GRPC errors
func init() {
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
}

// check function to catch all errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// create function for some weird integer manipulation
func create(x int32) *int32 {
	return &x
}

// Info Function to show colored text
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Message Defining message struct
type Message struct {
	Options     string `json:"options"`
	Target      string `json:"target"`
	Toolname    string `json:"toolname"`
	Responseurl string `json:"responseurl"`
}

// deleteEmpty function to remove empty values in a string array
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// deleteJob function to remove the terminated/completed pods
func deleteJob(runtooljob *jobv1.Job, clientset *kubernetes.Clientset) error {
	fmt.Printf("Job submitted: %s\n", runtooljob.Name)

	// running a loop continuously to check the status of the job
	for {
		cjob, err := clientset.Batch().Jobs(os.Getenv("kubebotnamespace")).Get(runtooljob.Name)
		check(err)

		cjobstatus := cjob.Status.Succeeded
		fmt.Printf("Job: %s Status: %d\n", runtooljob.Name, cjobstatus)

		if cjobstatus == 1 {
			break
		}
		time.Sleep(10 * time.Second)
	}

	// as soon as the job is done, retrieve all the jobs
	fjob, err := clientset.Batch().Jobs(os.Getenv("kubebotnamespace")).Get(runtooljob.Name)
	check(err)

	// as soon as the job is done, retrieve all the pods
	cpods, err := clientset.Core().Pods(os.Getenv("kubebotnamespace")).List(v1.ListOptions{})
	check(err)

	// Delete Job that finished running
	if fjob.Status.Succeeded == 1 {
		fmt.Printf("Job: %s Status: %d", runtooljob.Name, fjob.Status.Succeeded)
		fmt.Println("")
		fmt.Printf("Deleting Job: %s", runtooljob.Name)
		fmt.Println("")
		clientset.Batch().Jobs(os.Getenv("kubebotnamespace")).Delete(runtooljob.Name, &v1.DeleteOptions{}) //deleting the job

		//Delete PODs that have been completed where the jobs ran
		for _, element := range cpods.Items {
			if element.Status.Phase == "Succeeded" && element.Labels["app"] == "jobWorker" && element.Labels["component"] == "jobs" {
				fmt.Printf("Job Pod: %s Status: %s", element.Name, element.Status.Phase)
				fmt.Println("")
				fmt.Printf("Deleting Job Pod: %s", element.Name)
				clientset.Core().Pods(os.Getenv("kubebotnamespace")).Delete(element.Name, &v1.DeleteOptions{}) //deleting the job pod
				fmt.Println("")
			}
		}

	}
	return nil
}

// singlejob function to run a single tool as a job
func singlejob(tool string, url string, argstringarray []string, clientset *kubernetes.Clientset, wg *sync.WaitGroup) {

	var envvars []v1.EnvVar

	//setting up environment values for the tool
	switch tool {
	case "wfuzzbasicauthbrute":
		envvars = []v1.EnvVar{
			{
				Name:  "PROJECT_ID",
				Value: os.Getenv("project_id"),
			},
			{
				Name:  "RS_DATASET_NAME",
				Value: os.Getenv("RS_DATASET_NAME"),
			},
			{
				Name:  "RS_TABLE_NAME",
				Value: os.Getenv("RS_TABLE_NAME"),
			},
			{
				Name:  "WF_DATASET_NAME",
				Value: os.Getenv("WF_DATASET_NAME"),
			},
			{
				Name:  "WF_TABLE_NAME",
				Value: os.Getenv("WF_TABLE_NAME"),
			},
			{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/secretstore/" + os.Getenv("gac_keyname"),
			},
		}
	default:
		envvars = []v1.EnvVar{} //some tools don't need any environment values
	}

	runtooljob, err := clientset.Batch().Jobs(os.Getenv("kubebotnamespace")).Create(&jobv1.Job{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "tool-job", //naming standard for the job pod that will start
		},
		Spec: jobv1.JobSpec{
			Parallelism: create(1),
			Completions: create(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{ // adding labels to the job pod
						"app":       "jobWorker",
						"component": "jobs",
					},
					Namespace: os.Getenv("kubebotnamespace"),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{

						{
							Name:            tool + "-container",
							Image:           url,
							ImagePullPolicy: "IfNotPresent",
							Env:             envvars,        //environment values for the tool
							Args:            argstringarray, //arguments to run the tool
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "secretstore",
									MountPath: "/secretstore",
								},
							},
						},
					},
					RestartPolicy: "OnFailure",
					Volumes: []v1.Volume{
						{
							Name: "secretstore",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "googlesecret",
								},
							},
						},
					},
				},
			},
		},
	})
	check(err)

	//Delete the tool pods
	err = deleteJob(runtooljob, clientset)
	check(err)

	wg.Done()
}

// singlejobwithinitcontainer function to run a single tool as an init container and a utility after that, all this as one job
func singlejobwithinitcontainer(target string, tool string, utility string, url string, initcontainerstrarray []string, clientset *kubernetes.Clientset, wg *sync.WaitGroup) {

	var initcontainerstring string
	var maincontainerstrarray []string
	var envvars []v1.EnvVar
	var filepath string
	var argbuffer bytes.Buffer

	//depending upon the tool, creating the arguments string array for the tool (init container)
	//also creating the arguments string array for the utility container that runs after the tool
	switch tool {
	case "repo-supervisor": //implemented in an automation workflow
		maincontainerstrarray = []string{"-toolName", tool, "-filePath", resultsmountpath + "/results.json"}

	case "wfuzz": //implemented in an automation workflow
		maincontainerstrarray = []string{"-toolName", tool, "-filePath", resultsmountpath + "/results.csv"}

	case "nmap":
		filepath = resultsmountpath + "/nmap_" + target + ".nmap"
		initcontainerstrarray = append(initcontainerstrarray, "-oN")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		initcontainerstrarray = append(initcontainerstrarray, target)
		maincontainerstrarray = []string{filepath}

	case "sublist3r":
		filepath = resultsmountpath + "/sublist3r_" + target
		initcontainerstrarray = append(initcontainerstrarray, "-d")
		initcontainerstrarray = append(initcontainerstrarray, target)
		initcontainerstrarray = append(initcontainerstrarray, "-o")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		maincontainerstrarray = []string{filepath}

	case "gobuster":
		filepath = resultsmountpath + "/gobuster_" + target
		initcontainerstrarray = append(initcontainerstrarray, "-u")
		initcontainerstrarray = append(initcontainerstrarray, target)
		initcontainerstrarray = append(initcontainerstrarray, "-o")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		maincontainerstrarray = []string{filepath}

	case "enumall":
		filepath = resultsmountpath + "/enumall_" + target
		initcontainerstrarray = append(initcontainerstrarray, "-o")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		initcontainerstrarray = append(initcontainerstrarray, target)
		maincontainerstrarray = []string{filepath}

	case "subbrute":
		filepath = resultsmountpath + "/subbrute_" + target
		initcontainerstrarray = append(initcontainerstrarray, "-o")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		initcontainerstrarray = append(initcontainerstrarray, target)
		maincontainerstrarray = []string{filepath}

	case "trufflehog":
		tname := strings.Split(target, "/")
		tn := tname[len(tname)-1]
		filepath = resultsmountpath + "/trufflehog_" + tn
		initcontainerstrarray = append(initcontainerstrarray, "-o")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		initcontainerstrarray = append(initcontainerstrarray, target)
		maincontainerstrarray = []string{filepath}

	case "gitsecrets":
		tname := strings.Split(target, "/")
		tn := tname[len(tname)-1]
		filepath = resultsmountpath + "/gitsecrets_" + tn
		initcontainerstrarray = append(initcontainerstrarray, "./run.sh")
		initcontainerstrarray = append(initcontainerstrarray, target)
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		maincontainerstrarray = []string{filepath}

	case "gitallsecrets":
		var gsp string
		if strings.Contains(target, "http") {
			gsp = strings.Split(target, "/")[3] + "_" + strings.Split(target, "/")[4]
		} else {
			gsp = target
		}
		filepath = resultsmountpath + "/gitallsecrets_" + gsp
		initcontainerstrarray = append(initcontainerstrarray, target)
		initcontainerstrarray = append(initcontainerstrarray, "-output")
		initcontainerstrarray = append(initcontainerstrarray, filepath)
		initcontainerstrarray = append(initcontainerstrarray, "-token")
		initcontainerstrarray = append(initcontainerstrarray, os.Getenv("gitpersonaltoken"))
		maincontainerstrarray = []string{filepath}

		// case "altdns":
		// 	filepath = "/results/altdns_" + qmsg.Target

		// 	optarray = append(optarray, "-s")
		// 	optarray = append(optarray, filepath)
		// 	optarray = append(optarray, "-i")
		// 	optarray = append(optarray, qmsg.Target)

	}

	//since init containers are a total PITA as of now, we need to form the string
	for _, str := range initcontainerstrarray {
		argbuffer.WriteString("\"")
		argbuffer.WriteString(str)
		argbuffer.WriteString("\"")
		argbuffer.WriteString(",")
	}
	initcontainerargstring := argbuffer.String()
	initcontainerargstring = strings.TrimRight(initcontainerargstring, ",")

	//forming the final init container string to be supplied to the PodSpec
	initcontainerstring = "[{ \"name\": \"tool-init-container\", \"image\": \"" +
		url +
		"\", \"imagePullPolicy\": \"IfNotPresent\", \"args\": [" +
		initcontainerargstring +
		"],\"volumeMounts\": [{\"name\": \"results\", \"mountPath\":\"" + resultsmountpath + "\"}]}]"

	//depending upon the utility, they might need different environment values
	switch utility {
	case "converttobq":
		envvars = []v1.EnvVar{
			{
				Name:  "PROJECT_ID",
				Value: os.Getenv("project_id"),
			},
			{
				Name:  "RS_DATASET_NAME",
				Value: os.Getenv("RS_DATASET_NAME"),
			},
			{
				Name:  "RS_TABLE_NAME",
				Value: os.Getenv("RS_TABLE_NAME"),
			},
			{
				Name:  "WF_DATASET_NAME",
				Value: os.Getenv("WF_DATASET_NAME"),
			},
			{
				Name:  "WF_TABLE_NAME",
				Value: os.Getenv("WF_TABLE_NAME"),
			},
			{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/secretstore/" + os.Getenv("gac_keyname"),
			},
		}
	case "checkfile":
		envvars = []v1.EnvVar{
			{
				Name:  "gitpersonaltoken",
				Value: os.Getenv("gitpersonaltoken"),
			},
			{
				Name:  "githubowner",
				Value: os.Getenv("githubowner"),
			},
			{
				Name:  "githubacturl",
				Value: os.Getenv("githubacturl"),
			},
			{
				Name:  "slackwebhookurl",
				Value: os.Getenv("slackhook"),
			},
		}
	}

	dockerutilurl = os.Getenv("dockerrepourl") + "/utils/utils_" + utility + ":" + os.Getenv("dockerrepoversion")

	runtooljob, err := clientset.Batch().Jobs(os.Getenv("kubebotnamespace")).Create(&jobv1.Job{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "tool-job", //naming standard for the job pod that will start
		},
		Spec: jobv1.JobSpec{
			Parallelism: create(1),
			Completions: create(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{ // adding labels to the job pod
						"app":       "jobWorker",
						"component": "jobs",
					},
					Namespace: os.Getenv("kubebotnamespace"),
					Annotations: map[string]string{ //defining the tool to run as an init container
						"pod.beta.kubernetes.io/init-containers": initcontainerstring,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{ //running utility as the container after the init container completes!

						{
							Name:            utility + "-container",
							Image:           dockerutilurl,
							ImagePullPolicy: "IfNotPresent",
							Env:             envvars,
							Args:            maincontainerstrarray,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "results",
									MountPath: resultsmountpath,
								},
								{
									Name:      "secretstore",
									MountPath: "/secretstore",
								},
							},
						},
					},
					RestartPolicy: "OnFailure",
					Volumes: []v1.Volume{
						{
							Name: "results",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium: "",
								},
							},
						},
						{
							Name: "secretstore",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "googlesecret",
								},
							},
						},
					},
				},
			},
		},
	})
	check(err)

	//Cleaning up
	err = deleteJob(runtooljob, clientset)
	check(err)

	wg.Done()

}

// run the tool
func tooljob(qmsg Message) {

	// config, err := rest.InClusterConfig()
	// check(err)

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("kubeconfig"))
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	check(err)

	//switching whether its an automation workflow or the default case which is a single tool
	switch qmsg.Toolname {
	case "wfuzzbasicauthbrute": //automation workflow

		//start 2 goroutines for each tool above
		//check the status of the pod and on terminated status, waitgroup.done. no channel needed because data is not being passed
		//data is just getting stored in BQ
		//control comes back to the main program
		//starts the bruteforcing pod, results being sent to slack

		var url string
		var argstring []string
		tools := []string{"repo-supervisor", "wfuzz"}

		var wg sync.WaitGroup
		for _, tool := range tools {
			switch tool {
			case "repo-supervisor":
				url = os.Getenv("dockerrepourl") + "/tools/tools_gitallsecrets:" + os.Getenv("dockerrepoversion")
				argstring = []string{"-token", os.Getenv("gitpersonaltoken"), "-org", strings.Split(qmsg.Target, ".")[1], "-toolName", "repo-supervisor", "-output", resultsmountpath + "/results.json"}
			case "wfuzz":
				url = os.Getenv("dockerrepourl") + "/tools/tools_wfuzz:" + os.Getenv("dockerrepoversion")
				if qmsg.Target == "defcon.kubebot.io" {
					qmsg.Target = "104.198.4.57"
				}
				argstring = []string{"-w", "/data/SecLists/Discovery/Web_Content/tomcat.txt", "--hc", "404,429,400", "-o", "csv", "http://" + qmsg.Target + "/FUZZ/", resultsmountpath + "/results.csv"}
			}
			wg.Add(1)
			go singlejobwithinitcontainer(qmsg.Target, tool, "converttobq", url, argstring, clientset, &wg)
		}
		wg.Wait()

		//both repo-supervisor and wfuzz have finished running by now
		//lets run the bruteforcing pod
		brutearr := []string{"-target", qmsg.Target, "-slackHook", os.Getenv("slackhook")}

		var wg1 sync.WaitGroup
		wg1.Add(1)
		go singlejob("wfuzzbasicauthbrute", os.Getenv("dockerrepourl")+"/utils/utils_wfuzzbasicauthbrute:"+os.Getenv("dockerrepoversion"), brutearr, clientset, &wg1)
		wg1.Wait()

	default: //running single tools
		var url string

		url = os.Getenv("dockerrepourl") + "/tools/tools_" + qmsg.Toolname + ":" + os.Getenv("dockerrepoversion")

		// Getting the options, splitting on whitespace and adding them in a string array called optarray
		opt := strings.Split(qmsg.Options, " ")
		optarray := make([]string, len(opt))
		for index, element := range opt {
			optarray[index] = element
		}
		optarray = deleteEmpty(optarray)

		if qmsg.Toolname == "gitrob" { //special case because gitrob is a snowflake
			// we have the options array that looks like - "analyze --no-banner --no-server --title=target target"
			// now starting a job type pod to run the gitrob client
			// we don't need no init containers and checkfile container for gitrob since we will store all results in the gitrob server

			optarray = append(optarray, "--title="+qmsg.Target)
			optarray = append(optarray, qmsg.Target)

			var wg sync.WaitGroup
			wg.Add(1)
			go singlejob(qmsg.Toolname, url, optarray, clientset, &wg)
			wg.Wait()

		} else { //for everything else

			// now starting a job type pod to run the tool
			// running the tool as init containers and the checkfile container after that
			// checkfile container git clones, compares the old fine with the new file, sends the diff
			var wg sync.WaitGroup
			wg.Add(1)
			go singlejobwithinitcontainer(qmsg.Target, qmsg.Toolname, "checkfile", url, optarray, clientset, &wg)
			wg.Wait()

		}

	}

}

func main() {

	err := gotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	proj := os.Getenv("project_id")
	subscriptionName := os.Getenv("tools_subname")
	resultsmountpath = os.Getenv("resultsmountpath")

	ctx := context.Background()

	psclient, err := pubsub.NewClient(ctx, proj)
	check(err)

	subscription = psclient.Subscription(subscriptionName)

	// Run the PubSub Subscription forever to keep receiving messages until a "Cancel" context is passed
	// Not passing the cancel context here hence the loop runs forever
	err = subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		fmt.Println("Got message: " + string(m.Data))

		if err := json.Unmarshal(m.Data, &qmsg); err != nil {
			fmt.Printf("Could not decode message data: %#v. ACKing the message as well\n", m.Data)
			m.Ack()
		}

		fmt.Println("Processing Message ID: " + m.ID)
		fmt.Println("Acknowledging the Message ID: " + m.ID + " in PubSub so that it is removed from the queue")
		m.Ack()
		fmt.Println("Message ID: " + m.ID + " ACKed\n\n")

		// creating a go routine to spawn and perform the job
		// while getting back to the main program and listening for more data on the topic
		go tooljob(qmsg)

	})
	// If the context is something else, panic!
	if err != context.Canceled {
		check(err)
	}

}
