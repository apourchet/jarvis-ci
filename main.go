package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	TokenPath     string
	HubSecretPath string
	BasePath      string
	OutputURI     string
	ServerPort    int

	CleanPrefix        string
	CleanThresholdHour string
)

func init() {
	flag.StringVar(&TokenPath, "t", "/jarvis-ci/token", "The API token for authenticating with GitHub")
	flag.StringVar(&HubSecretPath, "s", "/jarvis-ci/hubsecret", "The secret key for validating a request from GitHub")
	flag.StringVar(&BasePath, "b", "/jarvis-ci", "The root path of the webhooks Jarvis will make available to GitHub")
	flag.StringVar(&OutputURI, "output-uri", "https://jarvisci.org/outputs/", "The base uri of a build output")
	flag.IntVar(&ServerPort, "p", 8080, "The port that Jarvis should listen for webhooks on")
	flag.StringVar(&CleanPrefix, "clean-prefix", "old", "The prefix of the docker images that jarvis can clean periodically")
	flag.StringVar(&CleanThresholdHour, "clean-threshold", "2", "The age of a docker image to be cleaned up")
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()

	// Start the cleanup in background
	go startCleanup()

	// Read in the token
	token, err := ioutil.ReadFile(TokenPath)
	if err != nil {
		glog.Warningf("Failed to read token: %v", err)
	}

	// Read in the hub secret
	hubSecret, err := ioutil.ReadFile(HubSecretPath)
	if err != nil {
		glog.Warningf("Failed to read hub secret: %v", err)
		hubSecret = []byte("antoinesecret")
	}

	// Log the configuration
	printFlags()

	// Create the github client
	client := NewGithubClient(string(token), OutputURI)

	// Create the output handler
	outputhandler := DefaultOutputHandler()

	// Create the event handler
	eventhandler := NewEventHandler(RepoFullName, client, outputhandler)

	// Start the server
	http.HandleFunc(path.Join(BasePath, "/debug/status"), debug)
	http.HandleFunc(path.Join(BasePath, "/hook"), hook(hubSecret, eventhandler))
	http.HandleFunc(path.Join(BasePath, "/outputs")+"/", outputfunc(outputhandler))
	err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func startCleanup() {
	go func() {
		for {
			glog.Infof("Cleaning docker images...")
			out, err := exec.Command("/bin/bash", "/dockerclean.sh", CleanPrefix, CleanThresholdHour).CombinedOutput()
			glog.Infof("Cleaned up docker images: \n%v\n---\n%v", string(out), err)
			time.Sleep(30 * time.Minute)
		}
	}()
}

func outputfunc(outputhandler OutputHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		glog.Infof("Handling output request: %s", req.URL.Path)
		jobid := strings.TrimPrefix(req.URL.Path, "/outputs/")
		out := outputhandler.GetOutput(jobid)
		if out != "" {
			fmt.Fprintf(w, out)
		} else {
			fmt.Fprintf(w, "No output found for jobid '%s'.", jobid)
		}
	}
}

func hook(hubscrt []byte, eventhandler EventHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		glog.Infof("Handling hook request")
		glog.Infof("Request: %v", req)

		// Verify that it's coming from github
		payload, err := github.ValidatePayload(req, hubscrt)
		if err != nil {
			glog.Errorf("Failed to validate github hook payload: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Parse the event
		event, err := github.ParseWebHook(github.WebHookType(req), payload)
		if err != nil {
			glog.Errorf("Failed to parse github hook event: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Hook handled successfully
		fmt.Fprintf(w, "OK")

		// Process the event in the background
		go func() {
			switch event := event.(type) {
			case *github.PingEvent:
				err = eventhandler.OnPingEvent(event)
			case *github.PushEvent:
				err = eventhandler.OnPushEvent(event)
			}

			// If there is an error, log it
			if err != nil {
				glog.Errorf("Failed to handle github hook: %v", err)
				return
			}
		}()
	}
}

func debug(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling debug request")
	fmt.Fprintf(w, "OK")
}
