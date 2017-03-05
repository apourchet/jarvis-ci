package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	TokenPath     string
	Token         []byte
	HubSecretPath string
	HubSecret     []byte

	RepoFullName string

	BasePath   string
	ServerPort int
)

func init() {
	flag.StringVar(&TokenPath, "t", "/jarvis-ci/token", "The API token for authenticating with GitHub")
	flag.StringVar(&HubSecretPath, "s", "/jarvis-ci/hubsecret", "The secret key for validating a request from GitHub")
	flag.StringVar(&RepoFullName, "repo", "", "The full name of the repository we are watching from GitHub")

	flag.StringVar(&BasePath, "b", "/jarvis-ci", "The root path of the webhooks Jarvis will make available to GitHub")
	flag.IntVar(&ServerPort, "p", 8080, "The port that Jarvis should listen for webhooks on")

	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()

	// Log the configuration
	logConfig()

	// Read in the token
	var err error
	Token, err = ioutil.ReadFile(TokenPath)
	if err != nil {
		glog.Warningf("Failed to read token: %v", err)
	}

	// Read in the hub secret
	HubSecret, err = ioutil.ReadFile(HubSecretPath)
	if err != nil {
		glog.Warningf("Failed to read hub secret: %v", err)
		HubSecret = []byte("mysecret")
	}

	// Start the server
	http.HandleFunc(fmt.Sprintf("%s/debug/status", BasePath), debug)
	http.HandleFunc(fmt.Sprintf("%s/hook", BasePath), hook)
	err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func logConfig() {
	glog.Infof("Server port: %d", ServerPort)
	glog.Infof("Base path: %s", BasePath)

	glog.Infof("Token path: %s", TokenPath)
	glog.Infof("Hub secret path: %s", HubSecretPath)
	glog.Infof("Repository full name: %s", RepoFullName)
}

func debug(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling debug request")
	fmt.Fprintf(w, "OK")
}

func hook(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling hook request")
	glog.Infof("Request: %v", req)

	// Verify that it's coming from github
	payload, err := github.ValidatePayload(req, HubSecret)
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

	// Switch on its type
	switch event := event.(type) {
	case *github.PingEvent:
		glog.Infof("Received ping event")
	case *github.PushEvent:
		glog.Infof("Received push event")
		err = handlePushEvent(event)
	}

	if err != nil {
		glog.Errorf("Failed to handle github hook: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Done
	fmt.Fprintf(w, "OK")
}

// Returns an error if it is a bad request
func handlePushEvent(event *github.PushEvent) error {
	// Make sure every field exists
	if event.Repo == nil {
		return fmt.Errorf("Missing PushEvent.Repo")
	}
	if event.HeadCommit == nil {
		return fmt.Errorf("Missing PushEvent head commit")
	}
	if event.Repo.FullName == nil {
		return fmt.Errorf("Missing PushEvent repo full name")
	}
	if event.HeadCommit.ID == nil {
		return fmt.Errorf("Missing PushEvent head commit ID")
	}

	// Check the full name for security
	fullName := *event.Repo.FullName
	if RepoFullName != "" && fullName != RepoFullName {
		return fmt.Errorf("Will not handle requests for this repository: %s", fullName)
	}

	// Construct the clone URL
	prefix := fmt.Sprintf("https://%s@github.com", Token)
	cloneURL := fmt.Sprintf("%s/%s.git", prefix, fullName)

	// Clone repository
	glog.Infof("Cloning from %s", cloneURL)
	err := exec.Command("git", "clone", cloneURL, "/tmp/clone").Run() // TODO cycle names
	if err != nil {
		glog.Errorf("Failed to clone directory: %s", cloneURL)
		return nil
	}

	// Execute test command
	cmd := exec.Command("make", "test")
	cmd.Dir = "/tmp/clone"
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Errorf("Test command failed: %v", err)
		glog.Errorf("Combined output: %s", out)
		return nil
	}

	// TODO post status
	return nil
}
